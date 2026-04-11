package handler

import (
	"context"
	"errors"

	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Rusich90/GophKeeper/internal/server/service"
	"github.com/Rusich90/GophKeeper/pkg/pb"
)

type AuthServer struct {
	pb.UnimplementedAuthServer
	authService *service.AuthService
	log         *slog.Logger
}

func NewAuthServer(authService *service.AuthService, log *slog.Logger) *AuthServer {
	return &AuthServer{
		authService: authService,
		log:         log,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	s.log.Info("Register request received", "login", req.Login)

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
