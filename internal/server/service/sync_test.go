package service

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Rusich90/GophKeeper/internal/server/storage/user"
	"github.com/Rusich90/GophKeeper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestNewSyncService(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	assert.NotNil(t, service)
	assert.Equal(t, mockUserRepo, service.userRepo)
	assert.Equal(t, log, service.log)
}

func TestSyncService_Pull_Success(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	expectedData := []byte("encrypted-data")
	expectedUpdatedAt := int64(1234567890)
	mockUserRepo.On("GetUserData", mock.Anything, "testuser").Return(expectedData, expectedUpdatedAt, nil)

	data, updatedAt, err := service.Pull(context.Background(), "testuser")

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	assert.Equal(t, expectedUpdatedAt, updatedAt)
	mockUserRepo.AssertExpectations(t)
}

func TestSyncService_Pull_EmptyData(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	expectedData := []byte{}
	expectedUpdatedAt := int64(1234567890)
	mockUserRepo.On("GetUserData", mock.Anything, "testuser").Return(expectedData, expectedUpdatedAt, nil)

	data, updatedAt, err := service.Pull(context.Background(), "testuser")

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	assert.Equal(t, expectedUpdatedAt, updatedAt)
	mockUserRepo.AssertExpectations(t)
}

func TestSyncService_Pull_UserNotFound(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	mockUserRepo.On("GetUserData", mock.Anything, "nonexistent").Return(nil, int64(0), user.ErrUserNotFound)

	data, updatedAt, err := service.Pull(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "failed to get user data")
	mockUserRepo.AssertExpectations(t)
}

func TestSyncService_Pull_DatabaseError(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	mockUserRepo.On("GetUserData", mock.Anything, "testuser").Return(nil, int64(0), errors.New("database error"))

	data, updatedAt, err := service.Pull(context.Background(), "testuser")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "failed to get user data")
	mockUserRepo.AssertExpectations(t)
}

func TestSyncService_Push_Success(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	data := []byte("encrypted-data")
	mockUserRepo.On("UpdateUserData", mock.Anything, "testuser", data, mock.AnythingOfType("int64")).Return(nil)

	updatedAt, err := service.Push(context.Background(), "testuser", data)

	assert.NoError(t, err)
	assert.NotZero(t, updatedAt)
	mockUserRepo.AssertExpectations(t)
}

func TestSyncService_Push_EmptyData(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	data := []byte{}
	mockUserRepo.On("UpdateUserData", mock.Anything, "testuser", data, mock.AnythingOfType("int64")).Return(nil)

	updatedAt, err := service.Push(context.Background(), "testuser", data)

	assert.NoError(t, err)
	assert.NotZero(t, updatedAt)
	mockUserRepo.AssertExpectations(t)
}

func TestSyncService_Push_LargeData(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	// Создаем большие данные (1MB)
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	mockUserRepo.On("UpdateUserData", mock.Anything, "testuser", data, mock.AnythingOfType("int64")).Return(nil)

	updatedAt, err := service.Push(context.Background(), "testuser", data)

	assert.NoError(t, err)
	assert.NotZero(t, updatedAt)
	mockUserRepo.AssertExpectations(t)
}

func TestSyncService_Push_DatabaseError(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	data := []byte("encrypted-data")
	mockUserRepo.On("UpdateUserData", mock.Anything, "testuser", data, mock.AnythingOfType("int64")).Return(errors.New("database error"))

	updatedAt, err := service.Push(context.Background(), "testuser", data)

	assert.Error(t, err)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "failed to update user data")
	mockUserRepo.AssertExpectations(t)
}

func TestSyncService_Push_UpdatedAtIncrement(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	data := []byte("encrypted-data")
	mockUserRepo.On("UpdateUserData", mock.Anything, "testuser", data, mock.AnythingOfType("int64")).Return(nil)

	// Делаем несколько последовательных вызовов Push
	var timestamps []int64
	for i := 0; i < 3; i++ {
		updatedAt, err := service.Push(context.Background(), "testuser", data)
		require.NoError(t, err)
		timestamps = append(timestamps, updatedAt)
	}

	// Проверяем, что временные метки увеличиваются
	assert.True(t, timestamps[1] >= timestamps[0])
	assert.True(t, timestamps[2] >= timestamps[1])
	mockUserRepo.AssertExpectations(t)
}

func TestSyncService_PullPushIntegration(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSyncService(mockUserRepo, log)

	// Тестируем последовательность Pull -> Push -> Pull
	initialData := []byte("initial-data")
	initialUpdatedAt := int64(1234567890)
	mockUserRepo.On("GetUserData", mock.Anything, "testuser").Return(initialData, initialUpdatedAt, nil).Once()

	// Первый Pull
	data, updatedAt, err := service.Pull(context.Background(), "testuser")
	require.NoError(t, err)
	assert.Equal(t, initialData, data)
	assert.Equal(t, initialUpdatedAt, updatedAt)

	// Push новых данных
	newData := []byte("new-data")
	mockUserRepo.On("UpdateUserData", mock.Anything, "testuser", newData, mock.AnythingOfType("int64")).Return(nil).Once()

	newUpdatedAt, err := service.Push(context.Background(), "testuser", newData)
	require.NoError(t, err)
	assert.NotZero(t, newUpdatedAt)

	// Второй Pull с обновленными данными
	mockUserRepo.On("GetUserData", mock.Anything, "testuser").Return(newData, newUpdatedAt, nil).Once()

	updatedData, updatedUpdatedAt, err := service.Pull(context.Background(), "testuser")
	require.NoError(t, err)
	assert.Equal(t, newData, updatedData)
	assert.Equal(t, newUpdatedAt, updatedUpdatedAt)

	mockUserRepo.AssertExpectations(t)
}
