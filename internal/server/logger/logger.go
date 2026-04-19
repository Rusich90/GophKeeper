package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
)

// InitLogger инициализирует структурированный логгер
func InitLogger(env string) *slog.Logger {
	var log *slog.Logger

	var handler slog.Handler
	opts := &slog.HandlerOptions{}

	switch env {
	case "prod":
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	log = slog.New(handler)

	// Устанавливаем глобальный логгер
	slog.SetDefault(log)

	return log
}

// GRPCLoggingInterceptor логирует все входящие gRPC запросы
func GRPCLoggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		log.Info("gRPC request started",
			"method", info.FullMethod,
			"request", req,
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			log.Error("gRPC request failed",
				"method", info.FullMethod,
				"duration", duration,
				"error", err,
			)
		} else {
			log.Info("gRPC request completed",
				"method", info.FullMethod,
				"duration", duration,
				"response", resp,
			)
		}

		return resp, err
	}
}
