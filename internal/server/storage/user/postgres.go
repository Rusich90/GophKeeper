package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rusich90/GophKeeper/internal/server/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type PGUserRepo struct {
	pool *pgxpool.Pool
}

func NewPGUserRepo(pool *pgxpool.Pool) *PGUserRepo {
	return &PGUserRepo{pool: pool}
}

func (r *PGUserRepo) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2)`

	_, err := r.pool.Exec(ctx, query, user.Login, user.PasswordHash)
	return err
}

func (r *PGUserRepo) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `SELECT login, password_hash FROM users WHERE login = $1`

	var user model.User
	err := r.pool.QueryRow(ctx, query, login).Scan(&user.Login, &user.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", ErrUserNotFound, login)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

func (r *PGUserRepo) GetUserData(ctx context.Context, login string) ([]byte, int64, error) {
	query := `SELECT encrypted_data, updated_at FROM users WHERE login = $1`

	var encryptedData []byte
	var updatedAt int64
	err := r.pool.QueryRow(ctx, query, login).Scan(&encryptedData, &updatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, fmt.Errorf("%w: %s", ErrUserNotFound, login)
		}
		return nil, 0, fmt.Errorf("failed to get user data: %w", err)
	}

	return encryptedData, updatedAt, nil
}

func (r *PGUserRepo) UpdateUserData(ctx context.Context, login string, encryptedData []byte, updatedAt int64) error {
	query := `UPDATE users SET encrypted_data = $1, updated_at = $2 WHERE login = $3`

	_, err := r.pool.Exec(ctx, query, encryptedData, updatedAt, login)
	return err
}
