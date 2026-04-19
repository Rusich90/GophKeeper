package grpc

import (
	"context"
)

// ClientInterface определяет интерфейс для gRPC клиента
type ClientInterface interface {
	// Auth methods
	Login(ctx context.Context, login, password string) (string, error)
	Register(ctx context.Context, login, password string) error
	Logout(ctx context.Context, token string) error
	
	// Sync methods
	Pull(ctx context.Context, token string) ([]byte, int64, error)
	Push(ctx context.Context, token string, data []byte) (int64, error)
	
	// Connection management
	Close() error
}