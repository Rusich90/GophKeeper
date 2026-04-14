package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	tokenExpiration = 10 * time.Minute
)

type RedisTokenRepo struct {
	client *redis.Client
}

func NewRedisTokenRepo(addr string) *RedisTokenRepo {
	return &RedisTokenRepo{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (r *RedisTokenRepo) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisTokenRepo) Close() error {
	return r.client.Close()
}

func (r *RedisTokenRepo) StoreToken(ctx context.Context, login string, token string) error {
	return r.client.Set(ctx, login, token, tokenExpiration).Err()
}

func (r *RedisTokenRepo) GetToken(ctx context.Context, login string) (string, error) {
	token, err := r.client.Get(ctx, login).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("token not found")
		}
		return "", err
	}
	return token, nil
}

func (r *RedisTokenRepo) DeleteToken(ctx context.Context, login string) error {
	return r.client.Del(ctx, login).Err()
}
