package service

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Rusich90/GophKeeper/internal/server/model"
	"github.com/Rusich90/GophKeeper/internal/server/storage/user"
	"github.com/Rusich90/GophKeeper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

func TestNewAuthService(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	assert.NotNil(t, service)
	assert.Equal(t, mockUserRepo, service.userRepo)
	assert.Equal(t, mockTokenRepo, service.tokenRepo)
	assert.Equal(t, mockJwtMgr, service.jwtMgr)
	assert.Equal(t, log, service.log)
	assert.NotNil(t, service.validator)
}

func TestAuthService_Register_Success(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	// Мок возвращает nil, когда пользователь не найден
	mockUserRepo.On("FindByLogin", mock.Anything, "testuser").Return(nil, user.ErrUserNotFound)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	err := service.Register(context.Background(), "testuser", "Password123!")

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_InvalidCredentials(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		password    string
		expectedErr string
	}{
		{
			name:        "empty login",
			login:       "",
			password:    "Password123!",
			expectedErr: "логин не может быть пустым",
		},
		{
			name:        "empty password",
			login:       "testuser",
			password:    "",
			expectedErr: "пароль не может быть пустым",
		},
		{
			name:        "short login",
			login:       "ab",
			password:    "Password123!",
			expectedErr: "логин должен содержать минимум 3 символа",
		},
		{
			name:        "short password",
			login:       "testuser",
			password:    "pass",
			expectedErr: "пароль должен содержать минимум 6 символов",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewUserRepository(t)
			mockTokenRepo := mocks.NewTokenRepository(t)
			mockJwtMgr := mocks.NewTokenManager(t)
			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

			err := service.Register(context.Background(), tt.login, tt.password)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	existingUser := &model.User{
		Login:        "existinguser",
		PasswordHash: "hashedpassword",
	}
	mockUserRepo.On("FindByLogin", mock.Anything, "existinguser").Return(existingUser, nil)

	err := service.Register(context.Background(), "existinguser", "Password123!")

	assert.Error(t, err)
	assert.Equal(t, ErrUserAlreadyExists, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_CreateError(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	mockUserRepo.On("FindByLogin", mock.Anything, "testuser").Return(nil, user.ErrUserNotFound)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(errors.New("database error"))

	err := service.Register(context.Background(), "testuser", "Password123!")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_FindByLoginError(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	mockUserRepo.On("FindByLogin", mock.Anything, "testuser").Return(nil, errors.New("database error"))

	err := service.Register(context.Background(), "testuser", "Password123!")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check user existence")
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	// Создаем хеш пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	require.NoError(t, err)

	existingUser := &model.User{
		Login:        "testuser",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.On("FindByLogin", mock.Anything, "testuser").Return(existingUser, nil)
	mockJwtMgr.On("GenerateToken", "testuser").Return("test-token-123", nil)
	mockTokenRepo.On("StoreToken", mock.Anything, "testuser", "test-token-123").Return(nil)

	token, err := service.Login(context.Background(), "testuser", "Password123!")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockUserRepo.AssertExpectations(t)
	mockJwtMgr.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		password    string
		expectedErr string
	}{
		{
			name:        "empty login",
			login:       "",
			password:    "Password123!",
			expectedErr: "логин не может быть пустым",
		},
		{
			name:        "empty password",
			login:       "testuser",
			password:    "",
			expectedErr: "пароль не может быть пустым",
		},
		{
			name:        "short login",
			login:       "ab",
			password:    "Password123!",
			expectedErr: "логин должен содержать минимум 3 символа",
		},
		{
			name:        "short password",
			login:       "testuser",
			password:    "pass",
			expectedErr: "пароль должен содержать минимум 6 символов",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewUserRepository(t)
			mockTokenRepo := mocks.NewTokenRepository(t)
			mockJwtMgr := mocks.NewTokenManager(t)
			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

			token, err := service.Login(context.Background(), tt.login, tt.password)

			assert.Error(t, err)
			assert.Empty(t, token)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	mockUserRepo.On("FindByLogin", mock.Anything, "nonexistent").Return(nil, user.ErrUserNotFound)

	token, err := service.Login(context.Background(), "nonexistent", "Password123!")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	// Создаем хеш пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("CorrectPassword123!"), bcrypt.DefaultCost)
	require.NoError(t, err)

	existingUser := &model.User{
		Login:        "testuser",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.On("FindByLogin", mock.Anything, "testuser").Return(existingUser, nil)

	token, err := service.Login(context.Background(), "testuser", "WrongPassword123!")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_StoreTokenError(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	// Создаем хеш пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	require.NoError(t, err)

	existingUser := &model.User{
		Login:        "testuser",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.On("FindByLogin", mock.Anything, "testuser").Return(existingUser, nil)
	mockJwtMgr.On("GenerateToken", "testuser").Return("test-token-123", nil)
	mockTokenRepo.On("StoreToken", mock.Anything, "testuser", "test-token-123").Return(errors.New("redis error"))

	token, err := service.Login(context.Background(), "testuser", "Password123!")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to store token")
	mockUserRepo.AssertExpectations(t)
	mockJwtMgr.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_Logout_Success(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	mockTokenRepo.On("DeleteToken", mock.Anything, "testuser").Return(nil)

	err := service.Logout(context.Background(), "testuser")

	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_Logout_DeleteTokenError(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	mockTokenRepo.On("DeleteToken", mock.Anything, "testuser").Return(errors.New("redis error"))

	err := service.Logout(context.Background(), "testuser")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete token")
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	testToken := "valid-test-token-123"

	mockJwtMgr.On("ValidateToken", testToken).Return("testuser", nil)
	mockTokenRepo.On("GetToken", mock.Anything, "testuser").Return(testToken, nil)

	login, err := service.ValidateToken(context.Background(), testToken)

	assert.NoError(t, err)
	assert.Equal(t, "testuser", login)
	mockJwtMgr.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken_InvalidJWT(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	mockJwtMgr.On("ValidateToken", "invalid-token").Return("", errors.New("invalid token"))

	login, err := service.ValidateToken(context.Background(), "invalid-token")

	assert.Error(t, err)
	assert.Empty(t, login)
	assert.Contains(t, err.Error(), "invalid token")
	mockJwtMgr.AssertExpectations(t)
}

func TestAuthService_ValidateToken_TokenNotFound(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	testToken := "valid-test-token-123"

	mockJwtMgr.On("ValidateToken", testToken).Return("testuser", nil)
	mockTokenRepo.On("GetToken", mock.Anything, "testuser").Return("", errors.New("token not found"))

	login, err := service.ValidateToken(context.Background(), testToken)

	assert.Error(t, err)
	assert.Empty(t, login)
	assert.Contains(t, err.Error(), "token not found or expired")
	mockJwtMgr.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken_TokenMismatch(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockTokenRepo := mocks.NewTokenRepository(t)
	mockJwtMgr := mocks.NewTokenManager(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewAuthService(mockUserRepo, mockTokenRepo, mockJwtMgr, log)

	testToken := "valid-test-token-123"

	mockJwtMgr.On("ValidateToken", testToken).Return("testuser", nil)
	// Возвращаем другой токен из хранилища
	mockTokenRepo.On("GetToken", mock.Anything, "testuser").Return("different-token", nil)

	login, err := service.ValidateToken(context.Background(), testToken)

	assert.Error(t, err)
	assert.Empty(t, login)
	assert.Contains(t, err.Error(), "invalid token")
	mockJwtMgr.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}
