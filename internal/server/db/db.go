package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultConnectionTimeout = 10 * time.Second
)

// NewConnection создает пул соединений к PostgreSQL
func NewConnection(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	// Создаем контекст с timeout для подключения
	ctx, cancel := context.WithTimeout(ctx, defaultConnectionTimeout)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return pool, nil
}
