package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig содержит настройки пула соединений к PostgreSQL
type PoolConfig struct {
	// ConnectTimeout - таймаут на установление соединения
	ConnectTimeout time.Duration
	// MaxConns - максимальное количество соединений в пуле
	MaxConns int
	// MinConns - минимальное количество соединений в пуле
	MinConns int
	// MaxConnLifetime - максимальное время жизни соединения
	MaxConnLifetime time.Duration
	// MaxConnIdleTime - максимальное время простоя соединения
	MaxConnIdleTime time.Duration
	// HealthCheckPeriod - период проверки здоровья соединений
	HealthCheckPeriod time.Duration
}

// defaultPoolConfig - настройки пула по умолчанию
var defaultPoolConfig = PoolConfig{
	ConnectTimeout:    10 * time.Second,
	MaxConns:          25,
	MinConns:          5,
	MaxConnLifetime:   1 * time.Hour,
	MaxConnIdleTime:   30 * time.Minute,
	HealthCheckPeriod: 1 * time.Minute,
}

// NewConnection создает настроенный пул соединений к PostgreSQL
func NewConnection(ctx context.Context, dsn string, cfg *PoolConfig) (*pgxpool.Pool, error) {
	// Используем конфиг по умолчанию если не передан
	if cfg == nil {
		cfg = &defaultPoolConfig
	}

	// Создаем контекст с таймаутом подключения
	ctx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	// Настраиваем конфигурацию пула
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Применяем настройки пула
	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = int32(cfg.MinConns)
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Создаем пул
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
