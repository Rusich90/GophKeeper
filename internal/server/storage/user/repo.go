package user

import (
	"context"

	"github.com/Rusich90/GophKeeper/internal/server/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByLogin(ctx context.Context, login string) (*model.User, error)
}
