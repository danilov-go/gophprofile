package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/sync/errgroup"

	"github.com/danilov-go/gophprofile/internal/auth"
	"github.com/danilov-go/gophprofile/internal/models"
)

const (
	testBucket = "avatars"
	userID     = "userID"
)

type AvatarSuite struct {
	suite.Suite

	pg        *postgres.PostgresContainer
	kafka     *kafka.KafkaContainer
	minio     *minio.MinioContainer
	s3        *awss3.Client
	net       string
	server    testcontainers.Container
	worker    testcontainers.Container
	netw      *testcontainers.DockerNetwork
	secretKey string
}

func (s *AvatarSuite) SetupSuite() {
	ctx := s.T().Context()
	var err error
	s.netw, err = network.New(ctx)
	s.Require().NoError(err)
	g, gCtx := errgroup.WithContext(ctx)
	var s3Host string
	g.Go(func() (err error) {
		s.pg, err = postgres.Run(gCtx, "postgres:16-alpine",
			postgres.WithDatabase("test"), postgres.WithUsername("postgres"), postgres.WithPassword("postgres"),
			network.WithNetwork([]string{"postgres"}, s.netw),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
		)
		return err
	})
	g.Go(func() (err error) {
		s.kafka, err = kafka.Run(gCtx, "confluentinc/confluent-local:7.8.0",
			kafka.WithClusterID("test"),
			network.WithNetwork([]string{"kafka"}, s.netw),
		)
		return err
	})
	g.Go(func() (err error) {
		s.minio, err = minio.Run(gCtx, "pgsty/minio:latest",
			minio.WithUsername("minioadmin"), minio.WithPassword("minioadmin"),
			network.WithNetwork([]string{"minio"}, s.netw),
		)
		if err == nil {
			s3Host, err = s.minio.ConnectionString(gCtx)
		}
		return err
	})
	s.Require().NoError(g.Wait())
	s3URL := "http://" + s3Host
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("minioadmin", "minioadmin", "")),
	)
	s.Require().NoError(err)
	s.s3 = awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(s3URL)
		o.UsePathStyle = true
	})
	_, err = s.s3.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(testBucket)})
	s.Require().NoError(err)
	pgURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("postgres", "postgres"),
		Host:     "postgres:5432",
		Path:     "test",
		RawQuery: "sslmode=disable",
	}
	s.secretKey = "test-secret-keys"
	env := map[string]string{
		"ADDRESS":        ":8080",
		"DATABASE_DSN":   pgURL.String(),
		"S3_ENDPOINT":    "http://minio:9000",
		"KAFKA_BROKERS":  "kafka:9092",
		"S3_BUCKET_NAME": testBucket,
		"S3_ACCESS_KEY":  "minioadmin",
		"S3_SECRET_KEY":  "minioadmin",
		"KAFKA_TOPIC":    "avatar-events",
		"KAFKA_GROUP_ID": "e2e-group",
		"WEB_PATH":       "./web/static",
		"RETRY_DELAY":    "100ms",
		"SECRET_KEY":     s.secretKey,
	}
	dir := s.T().TempDir()
	serverPath := filepath.Join(dir, "server")
	buildServer := exec.CommandContext(ctx, "go", "build", "-o", "server", "./cmd/server")
	buildServer.Dir = ".."
	buildServer.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64")
	s.Require().NoError(buildServer.Run())
	s.Require().NoError(os.Rename("../server", serverPath))
	workerPath := filepath.Join(dir, "worker")
	buildWorker := exec.CommandContext(ctx, "go", "build", "-o", "worker", "./cmd/worker")
	buildWorker.Dir = ".."
	buildWorker.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64")
	s.Require().NoError(buildWorker.Run())
	s.Require().NoError(os.Rename("../worker", workerPath))
	s.worker = s.startContainer(ctx, workerPath, "worker", env, nil, nil)
	s.server = s.startContainer(ctx, serverPath, "server", env, []string{"8080/tcp"}, wait.ForHTTP("/health").WithPort("8080/tcp"))
	s.net, err = s.server.PortEndpoint(ctx, "8080/tcp", "http")
	s.Require().NoError(err)
}

func (s *AvatarSuite) startContainer(ctx context.Context, binPath, name string, env map[string]string, ports []string, waitStrategy wait.Strategy) testcontainers.Container {
	s.T().Helper()
	path := "/app/" + name
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "alpine:latest",
			ExposedPorts: ports,
			Networks:     []string{s.netw.Name},
			WorkingDir:   "/app",
			Env:          env,
			Files: []testcontainers.ContainerFile{
				{HostFilePath: binPath, ContainerFilePath: path, FileMode: 0o755},
				{HostFilePath: "../migrations", ContainerFilePath: "/app/migrations", FileMode: 0o755},
			},
			Cmd:        []string{path},
			WaitingFor: waitStrategy,
		},
		Started: true,
	})
	s.Require().NoError(err)
	return container
}

func (s *AvatarSuite) Test1UploadAndDeleteAvatar() {
	ctx := s.T().Context()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("image", "avatar.png")
	s.Require().NoError(err)
	s.Require().NoError(png.Encode(part, image.NewRGBA(image.Rect(0, 0, 800, 800))))
	s.Require().NoError(mw.Close())
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.net+"/api/v1/avatars", &buf)
	s.Require().NoError(err)
	token, err := auth.GenerateToken(userID, s.secretKey)
	s.Require().NoError(err)
	uploadReq.Header.Set("Authorization", "Bearer "+token)
	uploadReq.Header.Set("Content-Type", mw.FormDataContentType())
	uploadResp, err := http.DefaultClient.Do(uploadReq)
	s.Require().NoError(err)
	defer func() { _ = uploadResp.Body.Close() }()
	if uploadResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(uploadResp.Body)
		logs, _ := s.server.Logs(ctx)
		logBytes, _ := io.ReadAll(logs)
		s.T().Logf("response body: %s\nserver logs:\n%s", string(body), string(logBytes))
	}
	s.Require().Equal(http.StatusCreated, uploadResp.StatusCode)
	s.Require().Equal(http.StatusCreated, uploadResp.StatusCode)
	var resp models.UploadResponse
	s.Require().NoError(json.NewDecoder(uploadResp.Body).Decode(&resp))
	s.Require().NotEmpty(resp.ID)
	viewReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.net+"/api/v1/avatars/"+resp.ID, nil)
	s.Require().NoError(err)
	viewResp, err := http.DefaultClient.Do(viewReq)
	s.Require().NoError(err)
	defer func() { _ = viewResp.Body.Close() }()
	s.Require().Equal(http.StatusOK, viewResp.StatusCode)
	s.Require().Eventually(func() bool {
		metaReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.net+"/api/v1/avatars/"+resp.ID+"/metadata", nil)
		if err != nil {
			return false
		}
		metaResp, err := http.DefaultClient.Do(metaReq)
		if err != nil {
			return false
		}
		defer func() { _ = metaResp.Body.Close() }()
		var meta models.MetaDataResponse
		return metaResp.StatusCode == http.StatusOK &&
			json.NewDecoder(metaResp.Body).Decode(&meta) == nil &&
			len(meta.Thumbnails) == 2
	}, 20*time.Second, 200*time.Millisecond)
	list, err := s.s3.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{Bucket: aws.String(testBucket)})
	s.Require().NoError(err)
	s.Require().Len(list.Contents, 3)
	deleteReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.net+"/api/v1/avatars/"+resp.ID, nil)
	s.Require().NoError(err)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteResp, err := http.DefaultClient.Do(deleteReq)
	s.Require().NoError(err)
	defer func() { _ = deleteResp.Body.Close() }()
	s.Require().Equal(http.StatusNoContent, deleteResp.StatusCode)
	s.Require().Eventually(func() bool {
		list, err = s.s3.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{Bucket: aws.String(testBucket)})
		return err == nil && len(list.Contents) == 0
	}, 20*time.Second, 200*time.Millisecond)
}

func (s *AvatarSuite) Test2GracefulShutdown() {
	ctx := s.T().Context()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("image", "avatar.png")
	s.Require().NoError(err)
	s.Require().NoError(png.Encode(part, image.NewRGBA(image.Rect(0, 0, 800, 800))))
	s.Require().NoError(mw.Close())
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.net+"/api/v1/avatars", &buf)
	s.Require().NoError(err)
	token, err := auth.GenerateToken(userID, s.secretKey)
	s.Require().NoError(err)
	uploadReq.Header.Set("Authorization", "Bearer "+token)
	uploadReq.Header.Set("Content-Type", mw.FormDataContentType())
	type resp struct {
		status int
		err    error
	}
	respCh := make(chan resp, 1)
	go func() {
		uploadResp, err := http.DefaultClient.Do(uploadReq)
		if err != nil {
			respCh <- resp{err: err}
			return
		}
		defer func() { _ = uploadResp.Body.Close() }()
		respCh <- resp{status: uploadResp.StatusCode}
	}()
	time.Sleep(30 * time.Millisecond)
	stopTimeout := 10 * time.Second
	s.Require().NoError(s.server.Stop(ctx, &stopTimeout))
	select {
	case res := <-respCh:
		s.Require().NoError(res.err)
		s.Require().Equal(http.StatusCreated, res.status)
	case <-time.After(5 * time.Second):
		s.FailNow("таймаут ожидания ответа")
	}
	state, err := s.server.State(ctx)
	s.Require().NoError(err)
	s.Require().Equal(0, state.ExitCode)
}

func (s *AvatarSuite) TearDownSuite() {
	ctx := context.Background()
	for _, c := range []testcontainers.Container{
		s.server,
		s.worker,
		s.pg,
		s.kafka,
		s.minio,
	} {
		if c != nil {
			_ = c.Terminate(ctx)
		}
	}
	if s.netw != nil {
		_ = s.netw.Remove(ctx)
	}
}

func TestAvatarSuite(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	suite.Run(t, new(AvatarSuite))
}
