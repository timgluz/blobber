package blobstore

import (
	"bytes"
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

func (s *S3BlobStore) Put(ctx context.Context, key string, data []byte) error {
	s.logger.Debug("Put", slog.String("key", key), slog.Int("size", len(data)))

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	if err != nil {
		s.logger.Error("PutObject failed", slog.String("key", key), slog.String("bucket", s.Bucket), slog.Any("error", err))
		return err
	}

	return nil
}

func (s *S3BlobStore) Delete(ctx context.Context, key string) error {
	s.logger.Debug("Delete", slog.String("key", key))

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		s.logger.Error("DeleteObject failed", slog.String("key", key), slog.String("bucket", s.Bucket), slog.Any("error", err))
		return err
	}
	return nil
}

func (s *S3BlobStore) List(ctx context.Context, prefix string) ([]string, error) {
	s.logger.Debug("List", slog.String("prefix", prefix))

	var keys []string
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			s.logger.Error("ListObjectsV2 failed", slog.String("prefix", prefix), slog.String("bucket", s.Bucket), slog.Any("error", err))
			return nil, err
		}

		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}

	return keys, nil
}
