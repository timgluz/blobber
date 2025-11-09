package blobstore

import (
	"context"
	"os"
)

type EnvGCPCredentials struct {
	APIKeyVar string
}

func NewEnvGCPCredentials() *EnvGCPCredentials {
	return &EnvGCPCredentials{
		APIKeyVar: "GCP_API_KEY",
	}
}
func (p EnvGCPCredentials) Retrieve(ctx context.Context) (Credentials, error) {
	apiKey := os.Getenv(p.APIKeyVar)
	if apiKey == "" {
		return Credentials{}, ErrNoValidCredentials
	}

	return Credentials{
		APIKey: apiKey,
	}, nil
}

type JSONFileGCPCredentials struct {
	FilePath string
}

func NewJSONFileGCPCredentials(filePath string) *JSONFileGCPCredentials {
	return &JSONFileGCPCredentials{
		FilePath: filePath,
	}
}

func (p JSONFileGCPCredentials) Retrieve(ctx context.Context) (Credentials, error) {
	data, err := os.ReadFile(p.FilePath)
	if err != nil {
		return Credentials{}, ErrNoValidCredentials
	}

	return Credentials{
		CredentialsJSON: data,
	}, nil
}
