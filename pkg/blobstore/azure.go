package blobstore

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

type AzureConfig struct {
	TenantID  string `yaml:"tenant_id"`
	Endpoint  string `yaml:"endpoint"` // e.g., "https://<account_name>.blob.core.windows.net/"
	Container string `yaml:"container"`
}

func NewAzureClient(config AzureConfig, credsProvider CredentialsProvider, logger *slog.Logger) (*azblob.Client, error) {
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

	client, err := azblob.NewClient(config.Endpoint, azCreds, nil)
	if err != nil {
		return nil, err
	}

	return client, nil

}

type AzureBlobStore struct {
	Container string

	client *azblob.Client
	logger *slog.Logger
}

func (s *AzureBlobStore) getContainerClient() *container.Client {
	return s.client.ServiceClient().NewContainerClient(s.Container)
}

func (s *AzureBlobStore) getBlobClient(key string) *blob.Client {
	return s.getContainerClient().NewBlobClient(key)
}

func NewAzureBlobStore(container string, client *azblob.Client, logger *slog.Logger) (*AzureBlobStore, error) {
	return &AzureBlobStore{
		Container: container,
		client:    client,
		logger:    logger,
	}, nil
}

func (s *AzureBlobStore) Ping(ctx context.Context) error {
	defer ctx.Done()
	s.logger.Info("Pinging Azure Blob Store")

	_, err := s.getContainerClient().GetProperties(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to ping Azure Blob Store", "error", err)
		return err
	}

	s.logger.Info("Successfully pinged Azure Blob Store")
	return nil
}

func (s *AzureBlobStore) List(ctx context.Context, prefix string) ([]string, error) {
	pager := s.getContainerClient().NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	var keys []string
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, blobItem := range page.Segment.BlobItems {
			keys = append(keys, *blobItem.Name)
		}
	}

	return keys, nil
}

func (s *AzureBlobStore) Has(ctx context.Context, key string) error {
	blobClient := s.getBlobClient(key)

	if _, err := blobClient.GetProperties(ctx, nil); err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return ErrBlobNotFound
		}

		return err
	}

	return nil
}

func (s *AzureBlobStore) Get(ctx context.Context, key string) ([]byte, error) {
	blobClient := s.getBlobClient(key)

	getResp, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return nil, ErrBlobNotFound
		}
		return nil, err
	}
	defer getResp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(getResp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *AzureBlobStore) Put(ctx context.Context, key string, data []byte) error {

	_, err := s.client.UploadBuffer(ctx, s.Container, key, data, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *AzureBlobStore) Delete(ctx context.Context, key string) error {
	blobClient := s.getBlobClient(key)

	if _, err := blobClient.Delete(ctx, nil); err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return ErrBlobNotFound
		}
		return err
	}

	return nil
}
