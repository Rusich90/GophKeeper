package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MockTokenValidator - мок для тестирования middleware
type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func TestAuthMiddleware_RegisterAndLogin(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
	}{
		{
			name:       "Register method should pass",
			fullMethod: "/auth.AuthService/Register",
		},
		{
			name:       "Login method should pass",
			fullMethod: "/auth.AuthService/Login",
		},
		{
			name:       "Register with different path",
			fullMethod: "/some.package/Register",
		},
		{
			name:       "Login with different path",
			fullMethod: "/some.package/Login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок для TokenValidator
			validator := new(MockTokenValidator)

			// Создаем middleware
			middleware := AuthMiddleware(validator)

			// Создаем контекст без метаданных
			ctx := context.Background()

			// Создаем handler, который должен быть вызван
			handlerCalled := false
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				return "response", nil
			}

			// Создаем info
			info := &grpc.UnaryServerInfo{
				FullMethod: tt.fullMethod,
			}

			// Вызываем middleware
			resp, err := middleware(ctx, "request", info, handler)

			// Проверяем, что handler был вызван
			assert.True(t, handlerCalled, "Handler should be called for Register and Login methods")
			assert.NoError(t, err)
			assert.Equal(t, "response", resp)
		})
	}
}

func TestAuthMiddleware_MissingMetadata(t *testing.T) {
	validator := new(MockTokenValidator)
	middleware := AuthMiddleware(validator)

	ctx := context.Background()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/auth.AuthService/SomeMethod",
	}

	resp, err := middleware(ctx, "request", info, handler)

	assert.Nil(t, resp)
	assert.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "missing metadata")
}

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	validator := new(MockTokenValidator)
	middleware := AuthMiddleware(validator)

	// Создаем контекст с метаданными, но без заголовка авторизации
	md := metadata.New(map[string]string{
		"other-header": "value",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/auth.AuthService/SomeMethod",
	}

	resp, err := middleware(ctx, "request", info, handler)

	assert.Nil(t, resp)
	assert.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "missing authorization header")
}

func TestAuthMiddleware_InvalidAuthorizationHeaderFormat(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		expectedError string
	}{
		{
			name:          "no bearer prefix",
			authHeader:    "invalid-token",
			expectedError: "invalid authorization header format",
		},
		{
			name:          "wrong prefix",
			authHeader:    "Basic token",
			expectedError: "invalid authorization header format",
		},
		{
			name:          "empty header",
			authHeader:    "",
			expectedError: "invalid authorization header format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := new(MockTokenValidator)
			middleware := AuthMiddleware(validator)

			md := metadata.New(map[string]string{
				"authorization": tt.authHeader,
			})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", nil
			}

			info := &grpc.UnaryServerInfo{
				FullMethod: "/auth.AuthService/SomeMethod",
			}

			resp, err := middleware(ctx, "request", info, handler)

			assert.Nil(t, resp)
			assert.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.Unauthenticated, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	validator := new(MockTokenValidator)

	// Настраиваем мок для возврата ошибки при валидации токена
	validator.On("ValidateToken", mock.Anything, "invalid-token").Return("", errors.New("invalid token"))

	middleware := AuthMiddleware(validator)

	md := metadata.New(map[string]string{
		"authorization": "Bearer invalid-token",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/auth.AuthService/SomeMethod",
	}

	resp, err := middleware(ctx, "request", info, handler)

	assert.Nil(t, resp)
	assert.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "invalid or expired token")
}

func TestAuthMiddleware_SuccessfulAuthentication(t *testing.T) {
	validator := new(MockTokenValidator)

	// Настраиваем мок для успешной валидации токена
	expectedLogin := "testuser"
	validator.On("ValidateToken", mock.Anything, "valid-token").Return(expectedLogin, nil)

	middleware := AuthMiddleware(validator)

	md := metadata.New(map[string]string{
		"authorization": "Bearer valid-token",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handlerCalled := false
	var receivedCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		receivedCtx = ctx
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/auth.AuthService/SomeMethod",
	}

	resp, err := middleware(ctx, "request", info, handler)

	assert.True(t, handlerCalled, "Handler should be called")
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)

	// Проверяем, что login и токен добавлены в контекст
	login, ok := GetLoginFromContext(receivedCtx)
	assert.True(t, ok)
	assert.Equal(t, expectedLogin, login)

	token, ok := GetTokenFromContext(receivedCtx)
	assert.True(t, ok)
	assert.Equal(t, "valid-token", token)
}

func TestAuthMiddleware_BearerCaseInsensitive(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "uppercase Bearer",
			authHeader: "Bearer valid-token",
		},
		{
			name:       "lowercase bearer",
			authHeader: "bearer valid-token",
		},
		{
			name:       "mixed case",
			authHeader: "BeArEr valid-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := new(MockTokenValidator)

			// Настраиваем мок для успешной валидации токена
			// Токен будет передан без префикса "Bearer "
			validator.On("ValidateToken", mock.Anything, "valid-token").Return("testuser", nil)

			middleware := AuthMiddleware(validator)

			md := metadata.New(map[string]string{
				"authorization": tt.authHeader,
			})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			handlerCalled := false
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				return "response", nil
			}

			info := &grpc.UnaryServerInfo{
				FullMethod: "/auth.AuthService/SomeMethod",
			}

			resp, err := middleware(ctx, "request", info, handler)

			assert.True(t, handlerCalled, "Handler should be called")
			assert.NoError(t, err)
			assert.Equal(t, "response", resp)
		})
	}
}

func TestGetLoginFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		want   string
		wantOk bool
	}{
		{
			name:   "login exists in context",
			ctx:    context.WithValue(context.Background(), LoginContextKey, "testuser"),
			want:   "testuser",
			wantOk: true,
		},
		{
			name:   "login does not exist in context",
			ctx:    context.Background(),
			want:   "",
			wantOk: false,
		},
		{
			name:   "login is not a string",
			ctx:    context.WithValue(context.Background(), LoginContextKey, 123),
			want:   "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			login, ok := GetLoginFromContext(tt.ctx)
			assert.Equal(t, tt.want, login)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestGetTokenFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		want   string
		wantOk bool
	}{
		{
			name:   "token exists in context",
			ctx:    context.WithValue(context.Background(), TokenContextKey, "valid-token"),
			want:   "valid-token",
			wantOk: true,
		},
		{
			name:   "token does not exist in context",
			ctx:    context.Background(),
			want:   "",
			wantOk: false,
		},
		{
			name:   "token is not a string",
			ctx:    context.WithValue(context.Background(), TokenContextKey, 123),
			want:   "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, ok := GetTokenFromContext(tt.ctx)
			assert.Equal(t, tt.want, token)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}
