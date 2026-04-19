package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TokenValidator - интерфейс для валидации токенов
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (string, error)
}

const (
	authorizationHeader = "authorization"
	bearerPrefix        = "bearer "
	loginKey            = "login"
)

type contextKey string

const LoginContextKey contextKey = "login"
const TokenContextKey contextKey = "token"

func AuthMiddleware(validator TokenValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Пропускаем методы Register и Login без проверки токена
		if strings.HasSuffix(info.FullMethod, "/Register") || strings.HasSuffix(info.FullMethod, "/Login") {
			return handler(ctx, req)
		}

		// Получаем метаданные из контекста
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Получаем токен из заголовка authorization
		authHeader := md.Get(authorizationHeader)
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := authHeader[0]
		if !strings.HasPrefix(strings.ToLower(token), bearerPrefix) {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		// Удаляем префикс "Bearer " (любой регистр)
		if len(token) > 7 {
			token = token[7:]
		}

		// Проверяем токен
		login, err := validator.ValidateToken(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Добавляем login и токен в контекст
		ctx = context.WithValue(ctx, LoginContextKey, login)
		ctx = context.WithValue(ctx, TokenContextKey, token)

		return handler(ctx, req)
	}
}

func GetLoginFromContext(ctx context.Context) (string, bool) {
	login, ok := ctx.Value(LoginContextKey).(string)
	return login, ok
}

func GetTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(TokenContextKey).(string)
	return token, ok
}
