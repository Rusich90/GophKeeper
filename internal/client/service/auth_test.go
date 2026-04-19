package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Rusich90/GophKeeper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewAuthService(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)

	// Создаем сервис с моком
	service := NewAuthService(mockGrpcClient)

	// Проверяем, что сервис создан
	assert.NotNil(t, service)

	// Преобразуем к конкретному типу для проверки полей
	authService, ok := service.(*authService)
	require.True(t, ok)

	assert.Equal(t, mockGrpcClient, authService.grpcClient)
	assert.NotNil(t, authService.validator)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)
	mockValidator := mocks.NewCredentialsValidator(t)

	// Создаем сервис с моками
	service := NewAuthService(mockGrpcClient)

	// Настраиваем моки
	mockValidator.On("ValidateCredentials", "testuser", "Password123!").Return(nil)
	mockGrpcClient.On("Login", mock.Anything, "testuser", "Password123!").Return("test-token-123", nil)

	// Заменяем валидатор в сервисе на мок
	authService, ok := service.(*authService)
	require.True(t, ok)
	authService.validator = mockValidator

	// Выполняем тест
	token, err := service.Login(context.Background(), "testuser", "Password123!")

	// Проверяем результаты
	assert.NoError(t, err)
	assert.Equal(t, "test-token-123", token)

	// Проверяем, что моки были вызваны
	mockValidator.AssertExpectations(t)
	mockGrpcClient.AssertExpectations(t)
}

func TestAuthService_Login_ValidationFailed(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)
	mockValidator := mocks.NewCredentialsValidator(t)

	// Создаем сервис с моками
	service := NewAuthService(mockGrpcClient)

	// Заменяем валидатор в сервисе на мок
	authService, ok := service.(*authService)
	require.True(t, ok)
	authService.validator = mockValidator

	// Настраиваем моки
	validationErr := errors.New("validation failed")
	mockValidator.On("ValidateCredentials", "invalid", "pass").Return(validationErr)

	// Выполняем тест
	token, err := service.Login(context.Background(), "invalid", "pass")

	// Проверяем результаты
	assert.Error(t, err)
	assert.Equal(t, validationErr, err)
	assert.Empty(t, token)

	// Проверяем, что моки были вызваны
	mockValidator.AssertExpectations(t)
	// Убедимся, что gRPC клиент не был вызван
	mockGrpcClient.AssertNotCalled(t, "Login")
}

func TestAuthService_Login_GrpcFailed(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)
	mockValidator := mocks.NewCredentialsValidator(t)

	// Создаем сервис с моками
	service := NewAuthService(mockGrpcClient)

	// Заменяем валидатор в сервисе на мок
	authService, ok := service.(*authService)
	require.True(t, ok)
	authService.validator = mockValidator

	// Настраиваем моки
	grpcErr := errors.New("grpc connection failed")
	mockValidator.On("ValidateCredentials", "testuser", "Password123!").Return(nil)
	mockGrpcClient.On("Login", mock.Anything, "testuser", "Password123!").Return("", grpcErr)

	// Выполняем тест
	token, err := service.Login(context.Background(), "testuser", "Password123!")

	// Проверяем результаты
	assert.Error(t, err)
	assert.Equal(t, grpcErr, err)
	assert.Empty(t, token)

	// Проверяем, что моки были вызваны
	mockValidator.AssertExpectations(t)
	mockGrpcClient.AssertExpectations(t)
}

func TestAuthService_Register_Success(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)
	mockValidator := mocks.NewCredentialsValidator(t)

	// Создаем сервис с моками
	service := NewAuthService(mockGrpcClient)

	// Заменяем валидатор в сервисе на мок
	authService, ok := service.(*authService)
	require.True(t, ok)
	authService.validator = mockValidator

	// Настраиваем моки
	mockValidator.On("ValidateCredentials", "testuser", "Password123!").Return(nil)
	mockGrpcClient.On("Register", mock.Anything, "testuser", "Password123!").Return(nil)

	// Выполняем тест
	err := service.Register(context.Background(), "testuser", "Password123!")

	// Проверяем результаты
	assert.NoError(t, err)

	// Проверяем, что моки были вызваны
	mockValidator.AssertExpectations(t)
	mockGrpcClient.AssertExpectations(t)
}

func TestAuthService_Register_ValidationFailed(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)
	mockValidator := mocks.NewCredentialsValidator(t)

	// Создаем сервис с моками
	service := NewAuthService(mockGrpcClient)

	// Заменяем валидатор в сервисе на мок
	authService, ok := service.(*authService)
	require.True(t, ok)
	authService.validator = mockValidator

	// Настраиваем моки
	validationErr := errors.New("validation failed")
	mockValidator.On("ValidateCredentials", "invalid", "pass").Return(validationErr)

	// Выполняем тест
	err := service.Register(context.Background(), "invalid", "pass")

	// Проверяем результаты
	assert.Error(t, err)
	assert.Equal(t, validationErr, err)

	// Проверяем, что моки были вызваны
	mockValidator.AssertExpectations(t)
	// Убедимся, что gRPC клиент не был вызван
	mockGrpcClient.AssertNotCalled(t, "Register")
}

func TestAuthService_Register_GrpcFailed(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)
	mockValidator := mocks.NewCredentialsValidator(t)

	// Создаем сервис с моками
	service := NewAuthService(mockGrpcClient)

	// Заменяем валидатор в сервисе на мок
	authService, ok := service.(*authService)
	require.True(t, ok)
	authService.validator = mockValidator

	// Настраиваем моки
	grpcErr := errors.New("grpc connection failed")
	mockValidator.On("ValidateCredentials", "testuser", "Password123!").Return(nil)
	mockGrpcClient.On("Register", mock.Anything, "testuser", "Password123!").Return(grpcErr)

	// Выполняем тест
	err := service.Register(context.Background(), "testuser", "Password123!")

	// Проверяем результаты
	assert.Error(t, err)
	assert.Equal(t, grpcErr, err)

	// Проверяем, что моки были вызваны
	mockValidator.AssertExpectations(t)
	mockGrpcClient.AssertExpectations(t)
}

func TestAuthService_Logout_Success(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)
	
	// Создаем сервис с моками
	service := NewAuthService(mockGrpcClient)

	// Настраиваем моки
	mockGrpcClient.On("Logout", mock.Anything, "test-token-123").Return(nil)

	// Выполняем тест
	err := service.Logout(context.Background(), "test-token-123")

	// Проверяем результаты
	assert.NoError(t, err)

	// Проверяем, что моки были вызваны
	mockGrpcClient.AssertExpectations(t)
}

func TestAuthService_Logout_GrpcFailed(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)
	
	// Создаем сервис с моками
	service := NewAuthService(mockGrpcClient)

	// Настраиваем моки
	grpcErr := errors.New("grpc connection failed")
	mockGrpcClient.On("Logout", mock.Anything, "test-token-123").Return(grpcErr)

	// Выполняем тест
	err := service.Logout(context.Background(), "test-token-123")

	// Проверяем результаты
	assert.Error(t, err)
	assert.Equal(t, grpcErr, err)

	// Проверяем, что моки были вызваны
	mockGrpcClient.AssertExpectations(t)
}

func TestAuthService_Close_Success(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)
	
	// Создаем сервис с моками
	service := NewAuthService(mockGrpcClient)

	// Выполняем тест
	err := service.Close()

	// Проверяем результаты
	assert.NoError(t, err)

	// Проверяем, что gRPC клиент не был закрыт (это делает контейнер)
	mockGrpcClient.AssertNotCalled(t, "Close")
}