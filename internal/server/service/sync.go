package service

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/Rusich90/GophKeeper/internal/server/storage/user"
)

// SyncServiceInterface определяет интерфейс для сервиса синхронизации
type SyncServiceInterface interface {
	Pull(ctx context.Context, login string) ([]byte, int64, error)
	Push(ctx context.Context, login string, encryptedData []byte) (int64, error)
}

type SyncService struct {
	userRepo user.UserRepository
	log      *slog.Logger
}

func NewSyncService(userRepo user.UserRepository, log *slog.Logger) *SyncService {
	return &SyncService{
		userRepo: userRepo,
		log:      log,
	}
}

// Pull возвращает зашифрованные данные пользователя и время последнего обновления
func (s *SyncService) Pull(ctx context.Context, login string) ([]byte, int64, error) {
	s.log.Debug("Pull request", "login", login)

	encryptedData, updatedAt, err := s.userRepo.GetUserData(ctx, login)
	if err != nil {
		s.log.Error("Failed to get user data", "login", login, "error", err)
		return nil, 0, fmt.Errorf("failed to get user data: %w", err)
	}

	s.log.Debug("Pull successful", "login", login, "data_size", len(encryptedData), "updated_at", updatedAt)
	return encryptedData, updatedAt, nil
}

// Push обновляет зашифрованные данные пользователя и возвращает новое время обновления
func (s *SyncService) Push(ctx context.Context, login string, encryptedData []byte) (int64, error) {
	s.log.Debug("Push request", "login", login, "data_size", len(encryptedData))

	updatedAt := time.Now().Unix()
	err := s.userRepo.UpdateUserData(ctx, login, encryptedData, updatedAt)
	if err != nil {
		s.log.Error("Failed to update user data", "login", login, "error", err)
		return 0, fmt.Errorf("failed to update user data: %w", err)
	}

	s.log.Debug("Push successful", "login", login, "updated_at", updatedAt)
	return updatedAt, nil
}
