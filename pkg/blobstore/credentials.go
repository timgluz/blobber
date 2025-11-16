package blobstore

import "context"

type AuthProviderType string

const (
	AuthProviderEnv  AuthProviderType = "env"
	AuthProviderFile AuthProviderType = "file"
)

type Credentials struct {
	// AWS, Alicloud
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string

	// GCP
	APIKey          string
	CredentialsJSON []byte

	// Azure
	TenantID     string
	ClientID     string
	ClientSecret string
}

type CredentialsProvider interface {
	Retrieve(ctx context.Context) (Credentials, error)
}
