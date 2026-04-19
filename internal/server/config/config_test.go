package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig(t *testing.T) {
	tests := []struct {
		name     string
		setupEnv func()
		cleanup  func()
		want     *Config
	}{
		{
			name: "default values",
			setupEnv: func() {
				// Очищаем переменные окружения для тестирования дефолтных значений
				os.Unsetenv("GRPC_PORT")
				os.Unsetenv("DATABASE_DSN")
				os.Unsetenv("ENV")
				os.Unsetenv("REDIS_ADDR")
				os.Unsetenv("JWT_SECRET")
				os.Unsetenv("TOKEN_TTL")
			},
			cleanup: func() {
				os.Unsetenv("GRPC_PORT")
				os.Unsetenv("DATABASE_DSN")
				os.Unsetenv("ENV")
				os.Unsetenv("REDIS_ADDR")
				os.Unsetenv("JWT_SECRET")
				os.Unsetenv("TOKEN_TTL")
			},
			want: &Config{
				GRPCPort:    DefaultGRPCPort,
				DatabaseDSN: DefaultDatabaseDSN,
				Environment: DefaultEnvironment,
				RedisAddr:   DefaultRedisAddr,
				JWTSecret:   DefaultJWTSecret,
				TokenTTL:    DefaultTokenTTL,
			},
		},
		{
			name: "custom values from env",
			setupEnv: func() {
				os.Setenv("GRPC_PORT", "8080")
				os.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/test")
				os.Setenv("ENV", "production")
				os.Setenv("REDIS_ADDR", "redis:6379")
				os.Setenv("JWT_SECRET", "custom-secret")
				os.Setenv("TOKEN_TTL", "1h")
			},
			cleanup: func() {
				os.Unsetenv("GRPC_PORT")
				os.Unsetenv("DATABASE_DSN")
				os.Unsetenv("ENV")
				os.Unsetenv("REDIS_ADDR")
				os.Unsetenv("JWT_SECRET")
				os.Unsetenv("TOKEN_TTL")
			},
			want: &Config{
				GRPCPort:    "8080",
				DatabaseDSN: "postgres://user:pass@localhost:5432/test",
				Environment: "production",
				RedisAddr:   "redis:6379",
				JWTSecret:   "custom-secret",
				TokenTTL:    "1h",
			},
		},
		{
			name: "partial env override",
			setupEnv: func() {
				os.Setenv("GRPC_PORT", "9000")
				os.Setenv("JWT_SECRET", "another-secret")
				os.Unsetenv("DATABASE_DSN")
				os.Unsetenv("ENV")
				os.Unsetenv("REDIS_ADDR")
				os.Unsetenv("TOKEN_TTL")
			},
			cleanup: func() {
				os.Unsetenv("GRPC_PORT")
				os.Unsetenv("JWT_SECRET")
			},
			want: &Config{
				GRPCPort:    "9000",
				DatabaseDSN: DefaultDatabaseDSN,
				Environment: DefaultEnvironment,
				RedisAddr:   DefaultRedisAddr,
				JWTSecret:   "another-secret",
				TokenTTL:    DefaultTokenTTL,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanup()

			got := InitConfig()

			assert.Equal(t, tt.want.GRPCPort, got.GRPCPort)
			assert.Equal(t, tt.want.DatabaseDSN, got.DatabaseDSN)
			assert.Equal(t, tt.want.Environment, got.Environment)
			assert.Equal(t, tt.want.RedisAddr, got.RedisAddr)
			assert.Equal(t, tt.want.JWTSecret, got.JWTSecret)
			assert.Equal(t, tt.want.TokenTTL, got.TokenTTL)
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		setupEnv      func()
		cleanup       func()
		expectedValue string
	}{
		{
			name:          "env variable exists",
			key:           "TEST_VAR",
			defaultValue:  "default",
			setupEnv:      func() { os.Setenv("TEST_VAR", "value") },
			cleanup:       func() { os.Unsetenv("TEST_VAR") },
			expectedValue: "value",
		},
		{
			name:          "env variable does not exist",
			key:           "NON_EXISTENT_VAR",
			defaultValue:  "default",
			setupEnv:      func() {},
			cleanup:       func() {},
			expectedValue: "default",
		},
		{
			name:          "env variable is empty string",
			key:           "EMPTY_VAR",
			defaultValue:  "default",
			setupEnv:      func() { os.Setenv("EMPTY_VAR", "") },
			cleanup:       func() { os.Unsetenv("EMPTY_VAR") },
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanup()

			got := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expectedValue, got)
		})
	}
}

func TestConfigConstants(t *testing.T) {
	assert.Equal(t, "50051", DefaultGRPCPort)
	assert.Equal(t, "postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable", DefaultDatabaseDSN)
	assert.Equal(t, "dev", DefaultEnvironment)
	assert.Equal(t, "localhost:6379", DefaultRedisAddr)
	assert.Equal(t, "your-secret-key-change-in-production", DefaultJWTSecret)
	assert.Equal(t, "10m", DefaultTokenTTL)
}

func TestInitConfigReturnsNotNil(t *testing.T) {
	cfg := InitConfig()
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.GRPCPort)
	require.NotNil(t, cfg.DatabaseDSN)
	require.NotNil(t, cfg.Environment)
	require.NotNil(t, cfg.RedisAddr)
	require.NotNil(t, cfg.JWTSecret)
	require.NotNil(t, cfg.TokenTTL)
}
