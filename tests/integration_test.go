package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
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
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/sync/errgroup"

	"github.com/danilov-go/gophprofile/internal/models"
)

const (
	testBucket      = "avatars"
	envS3BucketName = "S3_BUCKET_NAME=" + testBucket
	envS3AccessKey  = "S3_ACCESS_KEY=minioadmin"
	envS3Key        = "S3_SECRET_KEY=minioadmin"
	envKafkaTopic   = "KAFKA_TOPIC=avatar-events"
	envKafkaGroupID = "KAFKA_GROUP_ID=e2e-group"
	envWebPath      = "WEB_PATH=./web/static"
	envRetryDelay   = "WORKER_RETRY_DELAY=100ms"
	envDatabaseDSN  = "DATABASE_DSN="
	envS3Endpoint   = "S3_ENDPOINT="
	envKafkaBrokers = "KAFKA_BROKERS="
	envAddress      = "ADDRESS="
)

type AvatarSuite struct {
	suite.Suite

	pg     *postgres.PostgresContainer
	kafka  *kafka.KafkaContainer
	minio  *minio.MinioContainer
	s3     *awss3.Client
	net    string
	server *exec.Cmd
	worker *exec.Cmd
}

func (s *AvatarSuite) SetupSuite() {
	ctx := s.T().Context()
	g, gCtx := errgroup.WithContext(ctx)
	var pgDSN, s3Host string
	var brokers []string
	g.Go(func() (err error) {
		s.pg, err = postgres.Run(gCtx, "postgres:16-alpine",
			postgres.WithDatabase("test"), postgres.WithUsername("postgres"), postgres.WithPassword("postgres"),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
		)
		if err == nil {
			pgDSN, err = s.pg.ConnectionString(gCtx, "sslmode=disable")
		}
		return err
	})
	g.Go(func() (err error) {
		s.kafka, err = kafka.Run(gCtx, "confluentinc/confluent-local:7.8.0", kafka.WithClusterID("test"))
		if err == nil {
			brokers, err = s.kafka.Brokers(gCtx)
		}
		return err
	})
	g.Go(func() (err error) {
		s.minio, err = minio.Run(gCtx, "pgsty/minio:latest", minio.WithUsername("minioadmin"), minio.WithPassword("minioadmin"))
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
	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	addr := l.Addr().String()
	_ = l.Close()
	s.net = "http://" + addr
	env := append(os.Environ(),
		envS3BucketName,
		envS3AccessKey,
		envS3Key,
		envKafkaTopic,
		envKafkaGroupID,
		envWebPath,
		envRetryDelay,
		envAddress+addr,
		envDatabaseDSN+pgDSN,
		envS3Endpoint+s3URL,
		envKafkaBrokers+strings.Join(brokers, ","),
	)
	s.worker = exec.CommandContext(ctx, "go", "run", "./cmd/worker")
	s.worker.Dir = ".."
	s.worker.Env = env
	s.worker.WaitDelay = time.Second
	s.Require().NoError(s.worker.Start())
	s.server = exec.CommandContext(ctx, "go", "run", "./cmd/server")
	s.server.Dir = ".."
	s.server.Env = env
	s.server.WaitDelay = time.Second
	s.Require().NoError(s.server.Start())
	s.Require().Eventually(func() bool {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.net+"/health", nil)
		if err != nil {
			return false
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 30*time.Second, 200*time.Millisecond)
}

func (s *AvatarSuite) TestUploadAndDeleteAvatar() {
	ctx := s.T().Context()
	userID := "userID"
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("image", "avatar.png")
	s.Require().NoError(err)
	s.Require().NoError(png.Encode(part, image.NewRGBA(image.Rect(0, 0, 800, 800))))
	s.Require().NoError(mw.Close())
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.net+"/api/v1/avatars", &buf)
	s.Require().NoError(err)
	uploadReq.Header.Set("X-User-Id", userID)
	uploadReq.Header.Set("Content-Type", mw.FormDataContentType())
	uploadResp, err := http.DefaultClient.Do(uploadReq)
	s.Require().NoError(err)
	defer func() { _ = uploadResp.Body.Close() }()
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
	deleteReq.Header.Set("X-User-Id", userID)
	deleteResp, err := http.DefaultClient.Do(deleteReq)
	s.Require().NoError(err)
	defer func() { _ = deleteResp.Body.Close() }()
	s.Require().Equal(http.StatusNoContent, deleteResp.StatusCode)
	s.Require().Eventually(func() bool {
		list, err = s.s3.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{Bucket: aws.String(testBucket)})
		return err == nil && len(list.Contents) == 0
	}, 20*time.Second, 200*time.Millisecond)
}

func (s *AvatarSuite) TearDownSuite() {
	for _, cmd := range []*exec.Cmd{s.server, s.worker} {
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}
	ctx := context.Background()
	for _, c := range []testcontainers.Container{s.pg, s.kafka, s.minio} {
		if c != nil {
			_ = c.Terminate(ctx)
		}
	}
}

func TestAvatarSuite(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	suite.Run(t, new(AvatarSuite))
}
