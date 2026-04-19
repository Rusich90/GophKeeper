package handler

import (
	"context"
	"fmt"

	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Rusich90/GophKeeper/internal/server/middleware"
	"github.com/Rusich90/GophKeeper/pkg/pb"
)

// SyncServiceInterface определяет интерфейс для сервиса синхронизации
type SyncServiceInterface interface {
	Pull(ctx context.Context, login string) ([]byte, int64, error)
	Push(ctx context.Context, login string, encryptedData []byte) (int64, error)
}

type SyncServer struct {
	pb.UnimplementedSyncServer
	syncService SyncServiceInterface
	log         *slog.Logger
}

func NewSyncServer(syncService SyncServiceInterface, log *slog.Logger) *SyncServer {
	return &SyncServer{
		syncService: syncService,
		log:         log,
	}
}

// Pull возвращает зашифрованные данные пользователя с сервера
func (s *SyncServer) Pull(ctx context.Context, _ *emptypb.Empty) (*pb.SyncResponse, error) {
	s.log.Info("Pull request received")

	// Получаем login из контекста (уже проверенный middleware)
	login, ok := middleware.GetLoginFromContext(ctx)
	if !ok {
		s.log.Error("Login not found in context")
		return nil, status.Error(codes.Internal, "login not found in context")
	}

	// Получаем данные пользователя
	encryptedData, updatedAt, err := s.syncService.Pull(ctx, login)
	if err != nil {
		s.log.Error("Pull failed", "login", login, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to pull data: %v", err))
	}

	s.log.Info("Pull successful", "login", login, "data_size", len(encryptedData), "updated_at", updatedAt)
	return &pb.SyncResponse{
		Data:      encryptedData,
		UpdatedAt: updatedAt,
	}, nil
}

// Push отправляет зашифрованные данные пользователя на сервер
func (s *SyncServer) Push(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	s.log.Info("Push request received", "data_size", len(req.Data))

	// Получаем login из контекста (уже проверенный middleware)
	login, ok := middleware.GetLoginFromContext(ctx)
	if !ok {
		s.log.Error("Login not found in context")
		return nil, status.Error(codes.Internal, "login not found in context")
	}

	// Обновляем данные пользователя
	updatedAt, err := s.syncService.Push(ctx, login, req.Data)
	if err != nil {
		s.log.Error("Push failed", "login", login, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to push data: %v", err))
	}

	s.log.Info("Push successful", "login", login, "updated_at", updatedAt)
	return &pb.SyncResponse{
		Data:      nil, // В ответе на Push данные не возвращаем
		UpdatedAt: updatedAt,
	}, nil
}
