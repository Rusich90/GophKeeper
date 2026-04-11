package app

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Rusich90/GophKeeper/internal/server/config"
	"github.com/Rusich90/GophKeeper/internal/server/db"
	"github.com/Rusich90/GophKeeper/internal/server/handler"
	"github.com/Rusich90/GophKeeper/internal/server/logger"
	"github.com/Rusich90/GophKeeper/internal/server/service"
	"github.com/Rusich90/GophKeeper/internal/server/storage/user"
	"github.com/Rusich90/GophKeeper/pkg/pb"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	shutdownTimeout = 10 * time.Second
)

type App struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	server *grpc.Server
	log    *slog.Logger
}

func NewApp(cfg *config.Config, log *slog.Logger) (*App, error) {
	a := &App{
		cfg: cfg,
		log: log,
	}

	ctx := context.Background()
	pool, err := db.NewConnection(ctx, cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	a.pool = pool

	userRepo := user.NewPGUserRepo(a.pool)
	authService := service.NewAuthService(userRepo, a.log)
	authServer := handler.NewAuthServer(authService, a.log)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logger.GRPCLoggingInterceptor(a.log),
		),
	)

	pb.RegisterAuthServer(grpcServer, authServer)

	reflection.Register(grpcServer)

	a.server = grpcServer

	return a, nil
}

func (a *App) Run() error {
	lis, err := net.Listen("tcp", ":"+a.cfg.GRPCPort)
	if err != nil {
		return err
	}

	a.log.Info("Server is listening", "port", a.cfg.GRPCPort)

	serverErr := make(chan error, 1)

	go func() {
		a.log.Info("Starting gRPC server...")
		if err := a.server.Serve(lis); err != nil {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		a.log.Error("Server error", "error", err)
		return err
	case sig := <-quit:
		a.log.Info("Received shutdown signal", "signal", sig.String())
	}

	a.log.Info("Shutting down server...")
	if err := a.Stop(); err != nil {
		a.log.Error("Error during shutdown", "error", err)
		return err
	}

	return nil
}

func (a *App) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		a.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		a.log.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		a.log.Warn("gRPC server shutdown timeout, forcing stop")
		a.server.Stop()
	}

	a.pool.Close()
	a.log.Info("Database pool closed")
	return nil
}
