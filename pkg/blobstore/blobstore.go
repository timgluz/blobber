package blobstore

import "context"

type BlobStoreType string

const (
	BlobStoreTypeS3  BlobStoreType = "s3"
	BlobStoreTypeGCP BlobStoreType = "gcp"
)

type BlobStore interface {
	// Ping checks the connectivity to the blob store.
	Ping(context.Context) error
	// Has checks if a blob with the given key exists.
	Has(context.Context, string) error
	Get(context context.Context, key string) ([]byte, error)
	Put(context context.Context, key string, data []byte) error
	Delete(context context.Context, key string) error
	List(context context.Context, prefix string) ([]string, error)
}
