package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	DefaultGRPCPort    string = "50051"
	DefaultDatabaseDSN string = "postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable"
	DefaultEnvironment string = "dev"
	DefaultRedisAddr   string = "localhost:6379"
	DefaultJWTSecret   string = "your-secret-key-change-in-production"
	DefaultTokenTTL    string = "10m"
)

type Config struct {
	GRPCPort    string
	DatabaseDSN string
	Environment string
	RedisAddr   string
	JWTSecret   string
	TokenTTL    string
}

// InitConfig загружает конфигурацию из .env файла или переменных окружения
func InitConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	return &Config{
		GRPCPort:    getEnv("GRPC_PORT", DefaultGRPCPort),
		DatabaseDSN: getEnv("DATABASE_DSN", DefaultDatabaseDSN),
		Environment: getEnv("ENV", DefaultEnvironment),
		RedisAddr:   getEnv("REDIS_ADDR", DefaultRedisAddr),
		JWTSecret:   getEnv("JWT_SECRET", DefaultJWTSecret),
		TokenTTL:    getEnv("TOKEN_TTL", DefaultTokenTTL),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
