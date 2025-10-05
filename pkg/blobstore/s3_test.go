package blobstore_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timgluz/blobber/pkg/blobstore"
)

const expectedSuccessContent = `{"success": true}`

func TestS3BlobStore_Get(t *testing.T) {
	logger := initTestLogger(t)
	credsProvider := blobstore.NewEnvS3Credentials()
	s3Config := loadTestS3Config(t, "../../tests/fixtures/r2_config.yaml")

	s3Client, err := blobstore.NewS3Client(*s3Config, credsProvider, logger)
	require.NoError(t, err, "failed to create S3 client")

	store, err := blobstore.NewS3BlobStore(s3Config.Bucket, s3Client, logger)
	require.NoError(t, err, "failed to create S3 blob store")

	ctx := context.Background()
	actualContent, err := store.Get(ctx, "test.json")

	require.NoError(t, err)
	logger.Info("fetched content", "content", string(actualContent))
	assert.Equal(t, []byte(expectedSuccessContent), actualContent)
}

func loadTestS3Config(t *testing.T, configPath string) *blobstore.S3Config {
	t.Helper()

	provider := blobstore.NewYamlS3Config(configPath)
	config, err := provider.Retrieve(nil)
	if err != nil {
		t.Fatalf("failed to load R2 config: %v", err)
	}
	return config
}

func initTestLogger(t *testing.T) *slog.Logger {
	t.Helper()

	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	return logger
}
