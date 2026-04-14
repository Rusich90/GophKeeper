package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCPort    string
	DatabaseDSN string
	Environment string
	RedisAddr   string
	JWTSecret   string
}

// InitConfig загружает конфигурацию из .env файла или переменных окружения
func InitConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	return &Config{
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		DatabaseDSN: getEnv("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable"),
		Environment: getEnv("ENV", "dev"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
