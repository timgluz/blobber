package secret

import "context"

type SecretStore interface {
	ValidateToken(ctx context.Context, token string) (bool, error)
}
