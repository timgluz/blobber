package secret

import (
	"context"
	"os"
)

type EnvSecretStore struct {
	envVar string
}

func NewEnvSecretStore(envVar string) *EnvSecretStore {
	return &EnvSecretStore{envVar: envVar}
}

func (s *EnvSecretStore) ValidateToken(ctx context.Context, token string) (bool, error) {
	envToken := os.Getenv(s.envVar)
	if token != envToken {
		return false, nil
	}

	return token == envToken, nil
}
