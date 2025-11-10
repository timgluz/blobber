package blobstore

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

type AzureConfig struct {
	TenantID  string `yaml:"tenant_id"`
	Endpoint  string `yaml:"endpoint"` // e.g., "https://<account_name>.blob.core.windows.net/"
	Container string `yaml:"container"`
}

func NewAzureClient(config AzureConfig, credsProvider CredentialsProvider, logger *slog.Logger) (*container.Client, error) {
	creds, err := credsProvider.Retrieve(context.Background())
	if err != nil {
		return nil, err
	}

	azCreds, err := azidentity.NewClientSecretCredential(creds.TenantID, creds.ClientID, creds.ClientSecret, nil)
	if err != nil {
		return nil, err
	}

	if config.Endpoint[len(config.Endpoint)-1] == '/' {
		config.Endpoint = config.Endpoint[:len(config.Endpoint)-1]
	}

	containerURL := fmt.Sprintf("%s/%s", config.Endpoint, config.Container)
	client, err := container.NewClient(containerURL, azCreds, nil)
	if err != nil {
		return nil, err
	}

	return client, nil

}

type AzureBlobStore struct {
	Container string

	containerClient *container.Client
	logger          *slog.Logger
}

func NewAzureBlobStore(container string, client *container.Client, logger *slog.Logger) (*AzureBlobStore, error) {
	return &AzureBlobStore{
		Container:       container,
		containerClient: client,
		logger:          logger,
	}, nil
}

func (s *AzureBlobStore) Ping(ctx context.Context) error {
	defer ctx.Done()
	s.logger.Info("Pinging Azure Blob Store")

	_, err := s.containerClient.GetProperties(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to ping Azure Blob Store", "error", err)
		return err
	}

	s.logger.Info("Successfully pinged Azure Blob Store")
	return nil
}

func (s *AzureBlobStore) Has(ctx context.Context, key string) error {
	return fmt.Errorf("Has method not implemented yet")
}

func (s *AzureBlobStore) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, fmt.Errorf("Get method not implemented yet")
}

func (s *AzureBlobStore) Put(ctx context.Context, key string, data []byte) error {
	return fmt.Errorf("Put method not implemented yet")
}

func (s *AzureBlobStore) Delete(ctx context.Context, key string) error {
	return fmt.Errorf("Delete method not implemented yet")
}

func (s *AzureBlobStore) List(ctx context.Context, prefix string) ([]string, error) {
	return nil, fmt.Errorf("List method not implemented yet")
}
