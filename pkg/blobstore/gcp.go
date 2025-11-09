package blobstore

import (
	"context"
	"io"
	"log/slog"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GCPConfig struct {
	ProjectID       string `yaml:"project_id"`
	Endpoint        string `yaml:"endpoint"`
	Bucket          string `yaml:"bucket"`
	CredentialsPath string `yaml:"credentials_path"`
}

func NewGCPClient(config GCPConfig, credsProvider CredentialsProvider, logger *slog.Logger) (*storage.Client, error) {
	creds, err := credsProvider.Retrieve(context.Background())
	if err != nil {
		return nil, err
	}

	if creds.APIKey == "" && len(creds.CredentialsJSON) == 0 {
		return nil, ErrNoValidCredentials
	}

	var options []option.ClientOption
	if creds.APIKey != "" {
		options = append(options, option.WithAPIKey(creds.APIKey))
	}

	if len(creds.CredentialsJSON) != 0 {
		options = append(options, option.WithCredentialsJSON(creds.CredentialsJSON))
	}

	if config.Endpoint != "" {
		options = append(options, option.WithEndpoint(config.Endpoint))
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, options...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

type GCPBlobStore struct {
	Bucket string

	client *storage.Client
	logger *slog.Logger
}

func NewGCPBlobStore(bucket string, client *storage.Client, logger *slog.Logger) (*GCPBlobStore, error) {
	if bucket == "" {
		return nil, ErrNoValidBucket
	}

	if client == nil {
		return nil, ErrNoValidBlobClient
	}

	if logger == nil {
		return nil, ErrNoValidLogger
	}

	return &GCPBlobStore{
		Bucket: bucket,
		client: client,
		logger: logger,
	}, nil
}

func (s *GCPBlobStore) Ping(ctx context.Context) error {
	defer ctx.Done()

	_, err := s.client.Bucket(s.Bucket).Attrs(ctx)
	return err
}

func (s *GCPBlobStore) Has(ctx context.Context, key string) error {
	defer ctx.Done()

	bucket := s.client.Bucket(s.Bucket)
	obj := bucket.Object(key)
	if _, err := obj.Attrs(ctx); err != nil {
		s.logger.Error("object does not exist", slog.String("key", key), slog.Any("error", err))
		return ErrBlobNotFound
	}

	return nil
}

func (s *GCPBlobStore) Get(ctx context.Context, key string) ([]byte, error) {
	defer ctx.Done()

	bucket := s.client.Bucket(s.Bucket)
	obj := bucket.Object(key)
	objAttrs, err := obj.Attrs(ctx)
	if err != nil {
		s.logger.Error("reading object attributes failed", slog.String("key", key), slog.Any("error", err))
		return nil, err
	}

	s.logger.Debug("object attributes", slog.String("key", key), slog.Int64("size", objAttrs.Size))

	ctx2, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	reader, err := obj.NewReader(ctx2)
	if err != nil {
		s.logger.Error("getting object reader failed", slog.String("key", key), slog.Any("error", err))
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		s.logger.Error("reading object data failed", slog.String("key", key), slog.Any("error", err))
		return nil, err
	}

	return data, nil
}

func (s *GCPBlobStore) Put(ctx context.Context, key string, data []byte) error {
	defer ctx.Done()

	bucket := s.client.Bucket(s.Bucket)
	obj := bucket.Object(key)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	if _, err := writer.Write(data); err != nil {
		s.logger.Error("Put failed", slog.String("key", key), slog.Any("error", err))
		return err
	}

	return nil
}

func (s *GCPBlobStore) Delete(ctx context.Context, key string) error {
	defer ctx.Done()

	bucket := s.client.Bucket(s.Bucket)
	obj := bucket.Object(key)
	if _, err := obj.Attrs(ctx); err != nil {
		s.logger.Error("object does not exist", slog.String("key", key), slog.Any("error", err))
		return ErrBlobNotFound
	}

	if err := obj.Delete(ctx); err != nil {
		s.logger.Error("Delete failed", slog.String("key", key), slog.Any("error", err))
		return err
	}

	return nil
}

func (s *GCPBlobStore) List(ctx context.Context, prefix string) ([]string, error) {
	defer ctx.Done()

	query := &storage.Query{Prefix: prefix}
	it := s.client.Bucket(s.Bucket).Objects(ctx, query)

	var keys []string
	for {
		attr, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			s.logger.Error("List failed", slog.String("prefix", prefix), slog.Any("error", err))
			return nil, err
		}

		keys = append(keys, attr.Name)
	}

	return keys, nil
}
