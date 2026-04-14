package handler

import (
	"context"
	"errors"

	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Rusich90/GophKeeper/internal/server/middleware"
	"github.com/Rusich90/GophKeeper/internal/server/service"
	"github.com/Rusich90/GophKeeper/pkg/pb"
	"github.com/Rusich90/GophKeeper/pkg/validator"
)

type AuthServer struct {
	pb.UnimplementedAuthServer
	authService *service.AuthService
	log         *slog.Logger
	validator   validator.CredentialsValidator
}

func NewAuthServer(authService *service.AuthService, log *slog.Logger) *AuthServer {
	return &AuthServer{
		authService: authService,
		log:         log,
		validator:   validator.NewCredentialsValidator(),
	}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	s.log.Info("Register request received", "login", req.Login)

	// Дополнительная валидация на уровне handler для защиты от некорректных запросов
	if err := s.validator.ValidateCredentials(req.Login, req.Password); err != nil {
		s.log.Warn("Invalid credentials in request", "login", req.Login, "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := s.authService.Register(ctx, req.Login, req.Password)
	if err != nil {
		s.log.Error("Register failed", "login", req.Login, "error", err)
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	s.log.Info("Register successful", "login", req.Login)
	return &pb.RegisterResponse{
		Message: "Registration successful",
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	s.log.Info("Login request received", "login", req.Login)

	// Дополнительная валидация на уровне handler для защиты от некорректных запросов
	if err := s.validator.ValidateCredentials(req.Login, req.Password); err != nil {
		s.log.Warn("Invalid credentials in request", "login", req.Login, "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := s.authService.Login(ctx, req.Login, req.Password)
	if err != nil {
		s.log.Error("Login failed", "login", req.Login, "error", err)
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid login or password")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	s.log.Info("Login successful", "login", req.Login)
	return &pb.LoginResponse{
		Token: token,
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, _ *emptypb.Empty) (*pb.LogoutResponse, error) {
	s.log.Info("Logout request received")

	// Получаем login из контекста (уже проверенный middleware)
	login, ok := middleware.GetLoginFromContext(ctx)
	if !ok {
		s.log.Error("Login not found in context")
		return nil, status.Error(codes.Internal, "login not found in context")
	}

	err := s.authService.Logout(ctx, login)
	if err != nil {
		s.log.Error("Logout failed", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.log.Info("Logout successful", "login", login)
	return &pb.LogoutResponse{
		Message: "Logout successful",
	}, nil
}
