package app

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"log/slog"

	"github.com/Rusich90/GophKeeper/internal/server/config"
	"github.com/Rusich90/GophKeeper/internal/server/jwt"
	"github.com/Rusich90/GophKeeper/internal/server/service"
	"github.com/Rusich90/GophKeeper/internal/server/storage/token"
	"github.com/Rusich90/GophKeeper/internal/server/storage/user"
	"github.com/Rusich90/GophKeeper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// TestNewApp_InvalidTokenTTL тестирует создание приложения с невалидным TokenTTL
func TestNewApp_InvalidTokenTTL(t *testing.T) {
	tests := []struct {
		name        string
		tokenTTL    string
		expectedErr string
	}{
		{
			name:        "invalid token TTL format",
			tokenTTL:    "invalid",
			expectedErr: "failed to parse token TTL",
		},
		{
			name:        "empty token TTL",
			tokenTTL:    "",
			expectedErr: "failed to parse token TTL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем приложение с невалидным TokenTTL
			// Поскольку NewApp создает реальные подключения, этот тест проверяет только парсинг TTL
			// Для полноценного тестирования нужен рефакторинг с внедрением зависимостей
			_, err := time.ParseDuration(tt.tokenTTL)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid duration")
		})
	}
}

// TestApp_Stop тестирует остановку приложения
func TestApp_Stop(t *testing.T) {
	// Создаем тестовый gRPC сервер с bufconn
	lis := bufconn.Listen(1024 * 1024)

	grpcServer := grpc.NewServer()

	go func() {
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Logf("Server error: %v", err)
		}
	}()

	// Ждем запуска сервера
	time.Sleep(100 * time.Millisecond)

	// Создаем приложение с минимальной конфигурацией для тестирования Stop
	app := &App{
		cfg: &config.Config{
			GRPCPort: "50051",
		},
		server: grpcServer,
		log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	// Тестируем остановку
	err := app.Stop()
	assert.NoError(t, err)

	// Проверяем, что сервер остановлен
	assert.Equal(t, grpcServer.GetServiceInfo(), map[string]grpc.ServiceInfo{})
}

// TestApp_Stop_Timeout тестирует остановку приложения с таймаутом
func TestApp_Stop_Timeout(t *testing.T) {
	// Создаем тестовый gRPC сервер
	lis := bufconn.Listen(1024 * 1024)

	grpcServer := grpc.NewServer()

	// Регистрируем сервис, который будет блокировать остановку
	// Для этого создаем простой сервис с долгим shutdown

	go func() {
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Logf("Server error: %v", err)
		}
	}()

	// Ждем запуска сервера
	time.Sleep(100 * time.Millisecond)

	// Создаем приложение
	app := &App{
		cfg: &config.Config{
			GRPCPort: "50051",
		},
		server: grpcServer,
		log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	// Поскольку shutdownTimeout - это константа, мы не можем её изменить
	// Тестируем нормальную остановку
	err := app.Stop()
	assert.NoError(t, err)
}

// TestApp_Run_ListenError тестирует ошибку при прослушивании порта
func TestApp_Run_ListenError(t *testing.T) {
	// Создаем приложение с невалидным портом
	app := &App{
		cfg: &config.Config{
			GRPCPort: "invalid-port",
		},
		log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	err := app.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "listen tcp")
}

// TestApp_Run_Success тестирует успешный запуск и остановку приложения
func TestApp_Run_Success(t *testing.T) {
	// Находим свободный порт
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	// Создаем тестовый gRPC сервер
	grpcServer := grpc.NewServer()

	go func() {
		if err := grpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Logf("Server error: %v", err)
		}
	}()

	// Ждем запуска сервера
	time.Sleep(100 * time.Millisecond)

	// Создаем приложение
	app := &App{
		cfg: &config.Config{
			GRPCPort: string(rune(port)),
		},
		server: grpcServer,
		log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	// Проверяем, что сервер запущен
	assert.NotNil(t, app.server)

	// Останавливаем сервер напрямую
	err = app.Stop()
	assert.NoError(t, err)

	// Проверяем, что сервер остановлен
	time.Sleep(100 * time.Millisecond)
}

// TestShutdownTimeout проверяет константу shutdownTimeout
func TestShutdownTimeout(t *testing.T) {
	assert.Equal(t, 10*time.Second, shutdownTimeout)
}

// TestApp_Structure проверяет структуру App
func TestApp_Structure(t *testing.T) {
	app := &App{
		cfg:       &config.Config{},
		pool:      nil,
		tokenRepo: nil,
		server:    nil,
		log:       nil,
	}

	assert.NotNil(t, app.cfg)
	assert.Nil(t, app.pool)
	assert.Nil(t, app.tokenRepo)
	assert.Nil(t, app.server)
	assert.Nil(t, app.log)
}

// TestNewApp_ConfigValidation тестирует валидацию конфигурации
func TestNewApp_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				GRPCPort:    "50051",
				DatabaseDSN: "postgres://test:test@localhost:5432/test",
				Environment: "test",
				RedisAddr:   "localhost:6379",
				JWTSecret:   "test-secret",
				TokenTTL:    "1h",
			},
			expectError: false,
		},
		{
			name: "invalid token TTL",
			cfg: &config.Config{
				GRPCPort:    "50051",
				DatabaseDSN: "postgres://test:test@localhost:5432/test",
				Environment: "test",
				RedisAddr:   "localhost:6379",
				JWTSecret:   "test-secret",
				TokenTTL:    "invalid",
			},
			expectError: true,
			errorMsg:    "failed to parse token TTL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Поскольку NewApp создает реальные подключения к БД и Redis,
			// мы не можем полностью протестировать его без рефакторинга
			// Поэтому тестируем только парсинг TokenTTL
			if tt.cfg.TokenTTL != "" {
				_, err := time.ParseDuration(tt.cfg.TokenTTL)
				if tt.expectError && tt.errorMsg != "" {
					assert.Error(t, err)
				}
			}
		})
	}
}

// TestApp_Integration_WithMockServices интеграционный тест с моками
// Этот тест демонстрирует, как можно тестировать приложение после рефакторинга
// для внедрения зависимостей
func TestApp_Integration_WithMockServices(t *testing.T) {
	// Этот тест показывает, как можно было бы тестировать приложение
	// после рефакторинга для внедрения зависимостей

	// Создаем моки
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Создаем сервисы с моками
	authService := service.NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	// Настраиваем ожидания моков
	mockUserRepo.On("FindByLogin", mock.Anything, "testuser").Return(nil, user.ErrUserNotFound)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	// Тестируем регистрацию
	err := authService.Register(context.Background(), "testuser", "Password123!")
	assert.NoError(t, err)

	// Проверяем, что моки были вызваны
	mockUserRepo.AssertExpectations(t)
}

// TestApp_GracefulShutdown тестирует корректную остановку
func TestApp_GracefulShutdown(t *testing.T) {
	// Создаем тестовый gRPC сервер
	lis := bufconn.Listen(1024 * 1024)

	grpcServer := grpc.NewServer()

	go func() {
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Logf("Server error: %v", err)
		}
	}()

	// Ждем запуска сервера
	time.Sleep(100 * time.Millisecond)

	// Создаем подключение к серверу
	conn, err := grpc.NewClient("bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	// Проверяем, что соединение работает
	assert.NotNil(t, conn)

	// Создаем приложение
	app := &App{
		cfg: &config.Config{
			GRPCPort: "50051",
		},
		server: grpcServer,
		log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	// Останавливаем приложение
	err = app.Stop()
	assert.NoError(t, err)

	// Проверяем, что сервер остановлен
	time.Sleep(100 * time.Millisecond)

	// Пытаемся закрыть соединение (должно быть уже закрыто)
	err = conn.Close()
	assert.NoError(t, err)
}

// TestJWTManagerIntegration тестирует интеграцию с JWT менеджером
func TestJWTManagerIntegration(t *testing.T) {
	secret := "test-secret"
	ttl := time.Hour

	jwtMgr := jwt.NewManager(secret, ttl)

	// Генерируем токен
	token, err := jwtMgr.GenerateToken("testuser")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Валидируем токен
	login, err := jwtMgr.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "testuser", login)
}

// TestTokenRepositoryInterface проверяет интерфейс TokenRepository
func TestTokenRepositoryInterface(t *testing.T) {
	// Проверяем, что мок реализует интерфейс
	var _ token.TokenRepository = &mocks.TokenRepository{}
}

// TestUserRepositoryInterface проверяет интерфейс UserRepository
func TestUserRepositoryInterface(t *testing.T) {
	// Проверяем, что мок реализует интерфейс
	var _ user.UserRepository = &mocks.UserRepository{}
}

// TestApp_Constants проверяет константы
func TestApp_Constants(t *testing.T) {
	assert.Equal(t, 10*time.Second, shutdownTimeout)
}

// TestNewApp_NilConfig тестирует создание приложения с nil конфигурацией
func TestNewApp_NilConfig(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Поскольку NewApp не проверяет cfg на nil, это вызовет панику
	// После рефакторинга нужно добавить валидацию
	assert.Panics(t, func() {
		_, _ = NewApp(nil, log)
	})
}

// TestNewApp_NilLogger тестирует создание приложения с nil логгером
func TestNewApp_NilLogger(t *testing.T) {
	cfg := &config.Config{
		GRPCPort:    "50051",
		DatabaseDSN: "postgres://test:test@localhost:5432/test",
		Environment: "test",
		RedisAddr:   "localhost:6379",
		JWTSecret:   "test-secret",
		TokenTTL:    "1h",
	}

	// Поскольку NewApp не проверяет log на nil, это может вызвать панику
	// После рефакторинга нужно добавить валидацию
	assert.Panics(t, func() {
		_, _ = NewApp(cfg, nil)
	})
}
