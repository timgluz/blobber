package blobstore

import (
	"context"
	"os"
)

type EnvAzureCredentials struct {
	TenantIDVar     string
	ClientIDVar     string
	ClientSecretVar string
}

func NewEnvAzureCredentials() *EnvAzureCredentials {
	return &EnvAzureCredentials{
		TenantIDVar:     "AZURE_TENANT_ID",
		ClientIDVar:     "AZURE_CLIENT_ID",
		ClientSecretVar: "AZURE_CLIENT_SECRET",
	}
}

func (p EnvAzureCredentials) Retrieve(ctx context.Context) (Credentials, error) {
	tenantID := os.Getenv(p.TenantIDVar)
	if tenantID == "" {
		return Credentials{}, ErrNoValidCredentials
	}

	clientID := os.Getenv(p.ClientIDVar)
	if clientID == "" {
		return Credentials{}, ErrNoValidCredentials
	}

	clientSecret := os.Getenv(p.ClientSecretVar)
	if clientSecret == "" {
		return Credentials{}, ErrNoValidCredentials
	}

	return Credentials{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}
