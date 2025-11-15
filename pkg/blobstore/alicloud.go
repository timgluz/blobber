package blobstore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

func NewAlicloudClient(config AlicloudConfig, credsProvider CredentialsProvider, logger *slog.Logger) (*oss.Client, error) {
	creds, err := credsProvider.Retrieve(context.Background())
	if err != nil {
		return nil, err
	}

	staticCreds := credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, "")

	cfg := oss.LoadDefaultConfig().
		WithRegion(config.Region).
		WithEndpoint(config.Endpoint).
		WithCredentialsProvider(staticCreds)

	return oss.NewClient(cfg), nil
}

type AlicloudConfig struct {
	Region   string `yaml:"region"`
	Bucket   string `yaml:"bucket"`
	Endpoint string `yaml:"endpoint"`
}

type AlicloudBlobStore struct {
	Config AlicloudConfig

	client *oss.Client
	logger *slog.Logger
}

func NewAlicloudBlobStore(config AlicloudConfig, client *oss.Client, logger *slog.Logger) (*AlicloudBlobStore, error) {
	return &AlicloudBlobStore{
		Config: config,
		client: client,
		logger: logger,
	}, nil
}

// Ping checks the connectivity to the blob store.
func (s *AlicloudBlobStore) Ping(ctx context.Context) error {
	ok, err := s.client.IsBucketExist(ctx, s.Config.Bucket)
	if err != nil {
		return fmt.Errorf("failed to ping Alicloud OSS: %w", err)
	}

	if !ok {
		return fmt.Errorf("bucket %s does not exist", s.Config.Bucket)
	}

	return nil
}

func (s *AlicloudBlobStore) List(ctx context.Context, prefix string) ([]string, error) {
	result, err := s.client.ListObjects(ctx, &oss.ListObjectsRequest{
		Bucket: oss.Ptr(s.Config.Bucket),
		Prefix: oss.Ptr(prefix),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list objects with prefix %s: %w", prefix, err)
	}

	var keys []string
	for _, object := range result.Contents {
		if object.Key != nil {
			keys = append(keys, oss.ToString(object.Key))
		}
	}

	return keys, nil
}

func (s *AlicloudBlobStore) Has(ctx context.Context, key string) error {
	ok, err := s.client.IsObjectExist(ctx, s.Config.Bucket, key)
	if err != nil {
		return fmt.Errorf("failed to check existence of object %s: %w", key, err)
	}

	if !ok {
		return ErrBlobNotFound
	}

	return nil
}

func (s *AlicloudBlobStore) Get(ctx context.Context, key string) ([]byte, error) {
	res, err := s.client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(s.Config.Bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", key, err)
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s data: %w", key, err)
	}

	return data, nil
}

func (s *AlicloudBlobStore) Put(ctx context.Context, key string, data []byte) error {
	buf := bytes.NewReader(data)

	_, err := s.client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket: oss.Ptr(s.Config.Bucket),
		Key:    oss.Ptr(key),
		Body:   buf,
	})
	if err != nil {
		return fmt.Errorf("failed to put object %s: %w", key, err)
	}

	return nil
}

func (s *AlicloudBlobStore) Delete(ctx context.Context, key string) error {
	if err := s.Has(ctx, key); err != nil {
		return ErrBlobNotFound
	}

	_, err := s.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(s.Config.Bucket),
		Key:    oss.Ptr(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", key, err)
	}

	return nil
}
