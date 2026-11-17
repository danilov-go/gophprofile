package s3

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/danilov-go/gophprofile/internal/config"
)

// S3Client определяет клиент для работы с S3.
type S3Client struct {
	client     *s3.Client
	bucketName *string
}

// NewS3Client создаёт новый экземпляр S3Client.
func NewS3Client(ctx context.Context, cfg config.S3Config) (*S3Client, error) {
	client := s3.New(s3.Options{
		Region:       cfg.Region,
		UsePathStyle: true,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		BaseEndpoint: aws.String(cfg.Endpoint),
	})
	name := aws.String(cfg.BucketName)
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: name,
	})
	if err != nil {
		var bucketAlreadyExists *types.BucketAlreadyExists
		var bucketAlreadyOwnedByYou *types.BucketAlreadyOwnedByYou
		if !errors.As(err, &bucketAlreadyExists) && !errors.As(err, &bucketAlreadyOwnedByYou) {
			return nil, err
		}
	}
	return &S3Client{
		client:     client,
		bucketName: name,
	}, nil
}

// Upload загружает данные в S3.
func (s *S3Client) Upload(ctx context.Context, s3Key string, imageData []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: s.bucketName,
		Key:    aws.String(s3Key),
		Body:   bytes.NewReader(imageData),
	})
	if err != nil {
		return err
	}
	return nil
}

// Download скачивает файл из S3.
func (s *S3Client) Download(ctx context.Context, s3Key string) ([]byte, error) {
	image, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: s.bucketName,
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return nil, err
	}
	defer image.Body.Close()
	return io.ReadAll(image.Body)
}

// Delete удаляет файл из S3.
func (s *S3Client) Delete(ctx context.Context, s3Key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: s.bucketName,
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return err
	}
	return nil
}

// Ping проверяет доступность хранилища.
func (s *S3Client) Ping(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: s.bucketName,
	})
	return err
}
