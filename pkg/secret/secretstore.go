package secret

import "context"

type AuthProviderType string

const (
	AuthProviderEnv  AuthProviderType = "env"
	AuthProviderFile AuthProviderType = "file"
)

type SecretStore interface {
	ValidateToken(ctx context.Context, token string) (bool, error)
}
