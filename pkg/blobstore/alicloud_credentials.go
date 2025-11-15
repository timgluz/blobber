package blobstore

import (
	"context"
	"os"
)

type EnvAlicloudCredentials struct {
	AccessKeyIDVar     string
	SecretAccessKeyVar string
}

func NewEnvAlicloudCredentials() *EnvAlicloudCredentials {
	return &EnvAlicloudCredentials{
		AccessKeyIDVar:     "OSS_ACCESS_KEY_ID",
		SecretAccessKeyVar: "OSS_SECRET_ACCESS_KEY",
	}
}

func (p EnvAlicloudCredentials) Retrieve(ctx context.Context) (Credentials, error) {
	accessKeyID := os.Getenv(p.AccessKeyIDVar)
	if accessKeyID == "" {
		return Credentials{}, ErrNoValidCredentials
	}

	secretAccessKey := os.Getenv(p.SecretAccessKeyVar)
	if secretAccessKey == "" {
		return Credentials{}, ErrNoValidCredentials
	}

	return Credentials{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}, nil
}
