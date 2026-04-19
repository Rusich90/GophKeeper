package token

import "context"

type TokenRepository interface {
	StoreToken(ctx context.Context, login string, token string) error
	GetToken(ctx context.Context, login string) (string, error)
	DeleteToken(ctx context.Context, login string) error
}
