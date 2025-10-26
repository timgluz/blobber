package blobstore

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type StaticS3Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func NewStaticS3Credentials(accessKeyID, secretAccessKey, sessionToken string) *StaticS3Credentials {
	return &StaticS3Credentials{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
	}
}

func (c StaticS3Credentials) Retrieve(ctx context.Context) (aws.Credentials, error) {
	if c.AccessKeyID == "" || c.SecretAccessKey == "" {
		return aws.Credentials{}, ErrNoValidCredentials
	}

	return aws.Credentials{
		AccessKeyID:     c.AccessKeyID,
		SecretAccessKey: c.SecretAccessKey,
		SessionToken:    c.SessionToken,
		Source:          "StaticProvider",
	}, nil
}

type EnvS3Credentials struct {
	AccessKeyIDVar     string
	SecretAccessKeyVar string
	SessionTokenVar    string
}

func NewEnvS3Credentials() *EnvS3Credentials {
	return &EnvS3Credentials{
		AccessKeyIDVar:     "AWS_ACCESS_KEY_ID",
		SecretAccessKeyVar: "AWS_SECRET_ACCESS_KEY",
		SessionTokenVar:    "AWS_SESSION_TOKEN",
	}
}

func (p EnvS3Credentials) Retrieve(ctx context.Context) (aws.Credentials, error) {

	accessKeyID := os.Getenv(p.AccessKeyIDVar)
	if accessKeyID == "" {
		return aws.Credentials{}, ErrNoValidCredentials
	}

	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		return aws.Credentials{}, ErrNoValidCredentials
	}

	sessionToken := os.Getenv("AWS_SESSION_TOKEN")

	return aws.Credentials{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
		Source:          "EnvS3Provider",
	}, nil
}
