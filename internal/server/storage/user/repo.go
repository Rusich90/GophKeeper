package user

import (
	"context"

	"github.com/Rusich90/GophKeeper/internal/server/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByLogin(ctx context.Context, login string) (*model.User, error)
	GetUserData(ctx context.Context, login string) ([]byte, int64, error)
	UpdateUserData(ctx context.Context, login string, encryptedData []byte, updatedAt int64) error
}
