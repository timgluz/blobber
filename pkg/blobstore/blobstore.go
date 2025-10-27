package blobstore

import "context"

type BlobStore interface {
	// Ping checks the connectivity to the blob store.
	Ping(context.Context) error

	Get(context context.Context, key string) ([]byte, error)
	Put(context context.Context, key string, data []byte) error
	Delete(context context.Context, key string) error
	List(context context.Context, prefix string) ([]string, error)
}
