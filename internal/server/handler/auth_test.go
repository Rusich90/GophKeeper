package handler

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Rusich90/GophKeeper/internal/server/middleware"
	"github.com/Rusich90/GophKeeper/internal/server/service"
	"github.com/Rusich90/GophKeeper/mocks"
	"github.com/Rusich90/GophKeeper/pkg/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log/slog"
)

func TestNewAuthServer(t *testing.T) {
	mockService := mocks.NewAuthService(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	server := NewAuthServer(mockService, log)

	assert.NotNil(t, server)
	assert.NotNil(t, server.authService)
	assert.Equal(t, log, server.log)
	assert.NotNil(t, server.validator)
}

func TestAuthServer_Register_Success(t *testing.T) {
	mockService := mocks.NewAuthService(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewAuthServer(mockService, log)

	mockService.On("Register", mock.Anything, "testuser", "Password123!").Return(nil)

	req := &pb.RegisterRequest{
		Login:    "testuser",
		Password: "Password123!",
	}

	resp, err := server.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Registration successful", resp.Message)
	mockService.AssertExpectations(t)
}

func TestAuthServer_Register_InvalidCredentials(t *testing.T) {
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
			mockService := mocks.NewAuthService(t)
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			server := NewAuthServer(mockService, log)

			req := &pb.RegisterRequest{
				Login:    tt.login,
				Password: tt.password,
			}

			resp, err := server.Register(context.Background(), req)

			assert.Error(t, err)
			assert.Nil(t, resp)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
			assert.Contains(t, st.Message(), tt.expectedErr)
		})
	}
}

func TestAuthServer_Register_UserAlreadyExists(t *testing.T) {
	mockService := mocks.NewAuthService(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewAuthServer(mockService, log)

	mockService.On("Register", mock.Anything, "existinguser", "Password123!").Return(service.ErrUserAlreadyExists)

	req := &pb.RegisterRequest{
		Login:    "existinguser",
		Password: "Password123!",
	}

	resp, err := server.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
	assert.Contains(t, st.Message(), "user already exists")
	mockService.AssertExpectations(t)
}

func TestAuthServer_Register_ServiceError(t *testing.T) {
	mockService := mocks.NewAuthService(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewAuthServer(mockService, log)

	mockService.On("Register", mock.Anything, "testuser", "Password123!").Return(errors.New("database error"))

	req := &pb.RegisterRequest{
		Login:    "testuser",
		Password: "Password123!",
	}

	resp, err := server.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "database error")
	mockService.AssertExpectations(t)
}

func TestAuthServer_Login_Success(t *testing.T) {
	mockService := mocks.NewAuthService(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewAuthServer(mockService, log)

	expectedToken := "valid-jwt-token"
	mockService.On("Login", mock.Anything, "testuser", "Password123!").Return(expectedToken, nil)

	req := &pb.LoginRequest{
		Login:    "testuser",
		Password: "Password123!",
	}

	resp, err := server.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedToken, resp.Token)
	mockService.AssertExpectations(t)
}

func TestAuthServer_Login_InvalidCredentials(t *testing.T) {
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
			mockService := mocks.NewAuthService(t)
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			server := NewAuthServer(mockService, log)

			req := &pb.LoginRequest{
				Login:    tt.login,
				Password: tt.password,
			}

			resp, err := server.Login(context.Background(), req)

			assert.Error(t, err)
			assert.Nil(t, resp)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
			assert.Contains(t, st.Message(), tt.expectedErr)
		})
	}
}

func TestAuthServer_Login_InvalidCredentialsError(t *testing.T) {
	mockService := mocks.NewAuthService(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewAuthServer(mockService, log)

	mockService.On("Login", mock.Anything, "testuser", "WrongPassword123!").Return("", service.ErrInvalidCredentials)

	req := &pb.LoginRequest{
		Login:    "testuser",
		Password: "WrongPassword123!",
	}

	resp, err := server.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "invalid login or password")
	mockService.AssertExpectations(t)
}

func TestAuthServer_Login_ServiceError(t *testing.T) {
	mockService := mocks.NewAuthService(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewAuthServer(mockService, log)

	mockService.On("Login", mock.Anything, "testuser", "Password123!").Return("", errors.New("database error"))

	req := &pb.LoginRequest{
		Login:    "testuser",
		Password: "Password123!",
	}

	resp, err := server.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "database error")
	mockService.AssertExpectations(t)
}

func TestAuthServer_Logout_Success(t *testing.T) {
	mockService := mocks.NewAuthService(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewAuthServer(mockService, log)

	mockService.On("Logout", mock.Anything, "testuser").Return(nil)

	// Создаем контекст с login
	ctx := context.WithValue(context.Background(), middleware.LoginContextKey, "testuser")

	resp, err := server.Logout(ctx, &emptypb.Empty{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Logout successful", resp.Message)
	mockService.AssertExpectations(t)
}

func TestAuthServer_Logout_MissingLoginInContext(t *testing.T) {
	mockService := mocks.NewAuthService(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewAuthServer(mockService, log)

	// Создаем контекст без login
	ctx := context.Background()

	resp, err := server.Logout(ctx, &emptypb.Empty{})

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "login not found in context")
}

func TestAuthServer_Logout_ServiceError(t *testing.T) {
	mockService := mocks.NewAuthService(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewAuthServer(mockService, log)

	mockService.On("Logout", mock.Anything, "testuser").Return(errors.New("redis error"))

	// Создаем контекст с login
	ctx := context.WithValue(context.Background(), middleware.LoginContextKey, "testuser")

	resp, err := server.Logout(ctx, &emptypb.Empty{})

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "redis error")
	mockService.AssertExpectations(t)
}
