package handler

import (
	"context"
	"errors"
	"io"
	"testing"

	"log/slog"

	"github.com/Rusich90/GophKeeper/internal/server/middleware"
	"github.com/Rusich90/GophKeeper/mocks"
	"github.com/Rusich90/GophKeeper/pkg/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNewSyncServer(t *testing.T) {
	mockService := mocks.NewSyncServiceInterface(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	server := NewSyncServer(mockService, log)

	assert.NotNil(t, server)
	assert.NotNil(t, server.syncService)
	assert.Equal(t, log, server.log)
}

func TestSyncServer_Pull_Success(t *testing.T) {
	mockService := mocks.NewSyncServiceInterface(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewSyncServer(mockService, log)

	expectedData := []byte("encrypted-data")
	expectedUpdatedAt := int64(1234567890)
	mockService.On("Pull", mock.Anything, "testuser").Return(expectedData, expectedUpdatedAt, nil)

	// Создаем контекст с login
	ctx := context.WithValue(context.Background(), middleware.LoginContextKey, "testuser")

	resp, err := server.Pull(ctx, &emptypb.Empty{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedData, resp.Data)
	assert.Equal(t, expectedUpdatedAt, resp.UpdatedAt)
	mockService.AssertExpectations(t)
}

func TestSyncServer_Pull_MissingLoginInContext(t *testing.T) {
	mockService := mocks.NewSyncServiceInterface(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewSyncServer(mockService, log)

	// Создаем контекст без login
	ctx := context.Background()

	resp, err := server.Pull(ctx, &emptypb.Empty{})

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "login not found in context")
}

func TestSyncServer_Pull_ServiceError(t *testing.T) {
	mockService := mocks.NewSyncServiceInterface(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewSyncServer(mockService, log)

	mockService.On("Pull", mock.Anything, "testuser").Return(nil, int64(0), errors.New("database error"))

	// Создаем контекст с login
	ctx := context.WithValue(context.Background(), middleware.LoginContextKey, "testuser")

	resp, err := server.Pull(ctx, &emptypb.Empty{})

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to pull data")
	mockService.AssertExpectations(t)
}

func TestSyncServer_Pull_EmptyData(t *testing.T) {
	mockService := mocks.NewSyncServiceInterface(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewSyncServer(mockService, log)

	expectedData := []byte{}
	expectedUpdatedAt := int64(1234567890)
	mockService.On("Pull", mock.Anything, "testuser").Return(expectedData, expectedUpdatedAt, nil)

	// Создаем контекст с login
	ctx := context.WithValue(context.Background(), middleware.LoginContextKey, "testuser")

	resp, err := server.Pull(ctx, &emptypb.Empty{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedData, resp.Data)
	assert.Equal(t, expectedUpdatedAt, resp.UpdatedAt)
	mockService.AssertExpectations(t)
}

func TestSyncServer_Push_Success(t *testing.T) {
	mockService := mocks.NewSyncServiceInterface(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewSyncServer(mockService, log)

	data := []byte("encrypted-data")
	expectedUpdatedAt := int64(1234567890)
	mockService.On("Push", mock.Anything, "testuser", data).Return(expectedUpdatedAt, nil)

	// Создаем контекст с login
	ctx := context.WithValue(context.Background(), middleware.LoginContextKey, "testuser")

	req := &pb.SyncRequest{
		Data: data,
	}

	resp, err := server.Push(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Data) // В ответе на Push данные не возвращаем
	assert.Equal(t, expectedUpdatedAt, resp.UpdatedAt)
	mockService.AssertExpectations(t)
}

func TestSyncServer_Push_MissingLoginInContext(t *testing.T) {
	mockService := mocks.NewSyncServiceInterface(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewSyncServer(mockService, log)

	// Создаем контекст без login
	ctx := context.Background()

	req := &pb.SyncRequest{
		Data: []byte("encrypted-data"),
	}

	resp, err := server.Push(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "login not found in context")
}

func TestSyncServer_Push_ServiceError(t *testing.T) {
	mockService := mocks.NewSyncServiceInterface(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewSyncServer(mockService, log)

	data := []byte("encrypted-data")
	mockService.On("Push", mock.Anything, "testuser", data).Return(int64(0), errors.New("database error"))

	// Создаем контекст с login
	ctx := context.WithValue(context.Background(), middleware.LoginContextKey, "testuser")

	req := &pb.SyncRequest{
		Data: data,
	}

	resp, err := server.Push(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to push data")
	mockService.AssertExpectations(t)
}

func TestSyncServer_Push_EmptyData(t *testing.T) {
	mockService := mocks.NewSyncServiceInterface(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewSyncServer(mockService, log)

	data := []byte{}
	expectedUpdatedAt := int64(1234567890)
	mockService.On("Push", mock.Anything, "testuser", data).Return(expectedUpdatedAt, nil)

	// Создаем контекст с login
	ctx := context.WithValue(context.Background(), middleware.LoginContextKey, "testuser")

	req := &pb.SyncRequest{
		Data: data,
	}

	resp, err := server.Push(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Data)
	assert.Equal(t, expectedUpdatedAt, resp.UpdatedAt)
	mockService.AssertExpectations(t)
}

func TestSyncServer_Push_LargeData(t *testing.T) {
	mockService := mocks.NewSyncServiceInterface(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewSyncServer(mockService, log)

	// Создаем большие данные (1MB)
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	expectedUpdatedAt := int64(1234567890)
	mockService.On("Push", mock.Anything, "testuser", data).Return(expectedUpdatedAt, nil)

	// Создаем контекст с login
	ctx := context.WithValue(context.Background(), middleware.LoginContextKey, "testuser")

	req := &pb.SyncRequest{
		Data: data,
	}

	resp, err := server.Push(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Data)
	assert.Equal(t, expectedUpdatedAt, resp.UpdatedAt)
	mockService.AssertExpectations(t)
}
