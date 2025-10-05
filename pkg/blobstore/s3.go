package blobstore

import (
	"context"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewS3Client(config S3Config, credsProvider aws.CredentialsProvider, logger *slog.Logger) (*s3.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if config.UsePathStyle {
			o.UsePathStyle = true
		}

		if config.Endpoint != "" {
			o.BaseEndpoint = aws.String(config.Endpoint)
		}

		if config.Region != "" {
			o.Region = config.Region
		}

		if credsProvider != nil {
			o.Credentials = credsProvider
		}
	})

	return client, nil
}

type S3BlobStore struct {
	Bucket string
	client *s3.Client
	logger *slog.Logger
}

func NewS3BlobStore(bucket string, client *s3.Client, logger *slog.Logger) (*S3BlobStore, error) {
	if bucket == "" {
		return nil, ErrNoValidBucket
	}

	if client == nil {
		return nil, ErrNoValidBlobClient
	}

	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return &S3BlobStore{bucket, client, logger}, nil
}

func (s *S3BlobStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.logger.Debug("Get", slog.String("key", key))

	response, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		s.logger.Error("GetObject failed", slog.String("key", key),
			slog.String("bucket", s.Bucket), slog.Any("error", err))
		return nil, err
	}

	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		s.logger.Error("ReadAll failed", slog.String("key", key), slog.String("bucket", s.Bucket), slog.Any("error", err))
		return nil, err
	}

	return content, nil
}
