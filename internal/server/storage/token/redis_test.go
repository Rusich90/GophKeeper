package token

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *RedisTokenRepo) {
	t.Helper()

	s, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	repo := &RedisTokenRepo{client: client}

	t.Cleanup(func() {
		client.Close()
		s.Close()
	})

	return s, repo
}

func TestRedisTokenRepo_StoreToken(t *testing.T) {
	tests := []struct {
		name    string
		login   string
		token   string
		wantErr bool
	}{
		{
			name:    "successful store",
			login:   "testuser",
			token:   "test-token-123",
			wantErr: false,
		},
		{
			name:    "empty login",
			login:   "",
			token:   "test-token-123",
			wantErr: false, // Redis allows empty keys
		},
		{
			name:    "empty token",
			login:   "testuser",
			token:   "",
			wantErr: false, // Redis allows empty values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, repo := setupTestRedis(t)
			ctx := context.Background()

			err := repo.StoreToken(ctx, tt.login, tt.token)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedisTokenRepo_GetToken(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		setupToken  string
		wantToken   string
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful get",
			login:      "testuser",
			setupToken: "test-token-123",
			wantToken:  "test-token-123",
			wantErr:    false,
		},
		{
			name:        "token not found",
			login:       "nonexistent",
			setupToken:  "",
			wantToken:   "",
			wantErr:     true,
			errContains: "token not found",
		},
		{
			name:       "empty login",
			login:      "",
			setupToken: "test-token-123",
			wantToken:  "test-token-123",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, repo := setupTestRedis(t)
			ctx := context.Background()

			if tt.setupToken != "" {
				err := repo.StoreToken(ctx, tt.login, tt.setupToken)
				require.NoError(t, err)
			}

			token, err := repo.GetToken(ctx, tt.login)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func TestRedisTokenRepo_DeleteToken(t *testing.T) {
	tests := []struct {
		name       string
		login      string
		setupToken string
		wantErr    bool
	}{
		{
			name:       "successful delete",
			login:      "testuser",
			setupToken: "test-token-123",
			wantErr:    false,
		},
		{
			name:       "delete nonexistent token",
			login:      "nonexistent",
			setupToken: "",
			wantErr:    false, // Redis DEL doesn't error on missing keys
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, repo := setupTestRedis(t)
			ctx := context.Background()

			if tt.setupToken != "" {
				err := repo.StoreToken(ctx, tt.login, tt.setupToken)
				require.NoError(t, err)
			}

			err := repo.DeleteToken(ctx, tt.login)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify token is deleted
				_, err = repo.GetToken(ctx, tt.login)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "token not found")
			}
		})
	}
}

func TestRedisTokenRepo_TokenExpiration(t *testing.T) {
	t.Run("token expires after configured time", func(t *testing.T) {
		s, repo := setupTestRedis(t)
		ctx := context.Background()

		login := "testuser"
		token := "test-token-123"

		err := repo.StoreToken(ctx, login, token)
		require.NoError(t, err)

		// Token should exist immediately
		retrievedToken, err := repo.GetToken(ctx, login)
		require.NoError(t, err)
		assert.Equal(t, token, retrievedToken)

		// Fast forward time past expiration
		s.FastForward(tokenExpiration + time.Second)

		// Token should be expired
		_, err = repo.GetToken(ctx, login)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token not found")
	})
}

func TestRedisTokenRepo_Ping(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		_, repo := setupTestRedis(t)
		ctx := context.Background()

		err := repo.Ping(ctx)
		assert.NoError(t, err)
	})
}

func TestRedisTokenRepo_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		s, repo := setupTestRedis(t)

		err := repo.Close()
		assert.NoError(t, err)

		// Verify connection is closed by trying to ping
		ctx := context.Background()
		err = repo.Ping(ctx)
		assert.Error(t, err)

		s.Close()
	})
}

func TestRedisTokenRepo_ConcurrentOperations(t *testing.T) {
	t.Run("concurrent store and get", func(t *testing.T) {
		_, repo := setupTestRedis(t)
		ctx := context.Background()

		const numGoroutines = 100
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				login := "user" + string(rune(id))
				token := "token" + string(rune(id))

				err := repo.StoreToken(ctx, login, token)
				assert.NoError(t, err)

				retrieved, err := repo.GetToken(ctx, login)
				assert.NoError(t, err)
				assert.Equal(t, token, retrieved)

				done <- true
			}(i)
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}

func TestNewRedisTokenRepo(t *testing.T) {
	t.Run("creates new repository", func(t *testing.T) {
		addr := "localhost:6379"
		repo := NewRedisTokenRepo(addr)

		assert.NotNil(t, repo)
		assert.NotNil(t, repo.client)
		assert.Equal(t, addr, repo.client.Options().Addr)

		repo.Close()
	})
}
