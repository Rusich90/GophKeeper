package service

import (
	"context"

	"github.com/Rusich90/GophKeeper/internal/client/grpc"
	"github.com/Rusich90/GophKeeper/pkg/validator"
)

// AuthService определяет интерфейс для работы с авторизацией
type AuthService interface {
	Login(ctx context.Context, login, password string) (string, error)
	Register(ctx context.Context, login, password string) error
	Logout(ctx context.Context, token string) error
	Close() error
}

// authService реализация AuthService через gRPC клиент
type authService struct {
	grpcClient grpc.ClientInterface
	validator  validator.CredentialsValidator
}

// NewAuthService создает новый экземпляр сервиса авторизации
func NewAuthService(grpcClient grpc.ClientInterface) AuthService {
	return &authService{
		grpcClient: grpcClient,
		validator:  validator.NewCredentialsValidator(),
	}
}

// Login выполняет вход пользователя
func (s *authService) Login(ctx context.Context, login, password string) (string, error) {
	// Валидация учетных данных на клиенте
	if err := s.validator.ValidateCredentials(login, password); err != nil {
		return "", err
	}

	return s.grpcClient.Login(ctx, login, password)
}

// Register выполняет регистрацию нового пользователя
func (s *authService) Register(ctx context.Context, login, password string) error {
	// Валидация учетных данных на клиенте
	if err := s.validator.ValidateCredentials(login, password); err != nil {
		return err
	}

	return s.grpcClient.Register(ctx, login, password)
}

// Logout выполняет выход пользователя
func (s *authService) Logout(ctx context.Context, token string) error {
	return s.grpcClient.Logout(ctx, token)
}

// Close закрывает ресурсы сервиса авторизации
// Примечание: gRPC клиент закрывается контейнером, поэтому здесь ничего не делаем
func (s *authService) Close() error {
	return nil
}
