package blobstore

import "context"

type Credentials struct {
	// AWS
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	// GCP
	APIKey          string
	CredentialsJSON []byte
}

type CredentialsProvider interface {
	Retrieve(ctx context.Context) (Credentials, error)
}
