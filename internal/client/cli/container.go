package cli

import (
	"fmt"

	"github.com/Rusich90/GophKeeper/internal/client/config"
	"github.com/Rusich90/GophKeeper/internal/client/grpc"
	"github.com/Rusich90/GophKeeper/internal/client/service"
	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/Rusich90/GophKeeper/internal/client/ui"
)

// Container содержит все зависимости CLI приложения
type Container struct {
	UI             *ui.UI
	Config         *config.Config
	GRPCClient     *grpc.Client
	AuthService    service.AuthService
	SecretService  service.SecretService
	SyncService    service.SyncService
	SessionStorage *storage.SessionStorage
	SecretStorage  *storage.SecretStorage
}

// NewContainer создает новый контейнер зависимостей
func NewContainer(cfg *config.Config, verbose bool) (*Container, error) {
	// Инициализация UI
	uiInstance := ui.NewUI(verbose)

	// Инициализация gRPC клиента
	grpcClient, err := grpc.NewClient(cfg.ServerAddr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к серверу: %w", err)
	}

	// Инициализация сервисов
	authService := service.NewAuthService(grpcClient)
	sessionStorage := storage.NewSessionStorage()
	secretService := service.NewSecretService()
	syncService := service.NewSyncService(grpcClient, sessionStorage)
	secretStorage := storage.NewSecretStorage()

	return &Container{
		UI:             uiInstance,
		Config:         cfg,
		GRPCClient:     grpcClient,
		AuthService:    authService,
		SecretService:  secretService,
		SyncService:    syncService,
		SessionStorage: sessionStorage,
		SecretStorage:  secretStorage,
	}, nil
}

// Close закрывает все ресурсы контейнера
func (c *Container) Close() error {
	var errs []error

	// Закрываем gRPC клиент
	if c.GRPCClient != nil {
		if err := c.GRPCClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("ошибка закрытия gRPC клиента: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("ошибки при закрытии ресурсов: %v", errs)
	}
	return nil
}
