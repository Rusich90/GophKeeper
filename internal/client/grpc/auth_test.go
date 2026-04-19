package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/Rusich90/GophKeeper/mocks"
	"github.com/Rusich90/GophKeeper/pkg/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TestNewClient_Success проверяет успешное создание gRPC клиента
func TestNewClient_Success(t *testing.T) {
	// Этот тест требует запущенного gRPC сервера, поэтому мы пропускаем его
	// В реальном проекте можно использовать testcontainers или встроенный сервер
	t.Skip("Requires running gRPC server")
}

// TestNewClient_ConnectionFailed проверяет ошибку при невозможности подключиться к серверу
func TestNewClient_ConnectionFailed(t *testing.T) {
	// Этот тест требует запущенного gRPC сервера, поэтому мы пропускаем его
	t.Skip("Requires running gRPC server")
}

// TestClient_Login_Success проверяет успешный вход пользователя
func TestClient_Login_Success(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	expectedToken := "test-jwt-token-123"
	mockAuthClient.On("Login", mock.Anything, mock.MatchedBy(func(req *pb.LoginRequest) bool {
		return req.Login == "testuser" && req.Password == "Password123!"
	})).Return(&pb.LoginResponse{Token: expectedToken}, nil)

	token, err := client.Login(context.Background(), "testuser", "Password123!")

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Login_Unauthenticated проверяет ошибку неверных учетных данных
func TestClient_Login_Unauthenticated(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Unauthenticated, "invalid credentials")
	mockAuthClient.On("Login", mock.Anything, mock.MatchedBy(func(req *pb.LoginRequest) bool {
		return req.Login == "testuser" && req.Password == "WrongPassword"
	})).Return(nil, grpcErr)

	token, err := client.Login(context.Background(), "testuser", "WrongPassword")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "invalid login or password")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Login_Unavailable проверяет ошибку недоступности сервера
func TestClient_Login_Unavailable(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Unavailable, "server unavailable")
	mockAuthClient.On("Login", mock.Anything, mock.Anything).Return(nil, grpcErr)

	token, err := client.Login(context.Background(), "testuser", "Password123!")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "server is unavailable")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Login_OtherError проверяет другие gRPC ошибки
func TestClient_Login_OtherError(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Internal, "internal server error")
	mockAuthClient.On("Login", mock.Anything, mock.Anything).Return(nil, grpcErr)

	token, err := client.Login(context.Background(), "testuser", "Password123!")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "login failed")
	assert.Contains(t, err.Error(), "internal server error")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Login_ConnectionError проверяет ошибку соединения
func TestClient_Login_ConnectionError(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	connErr := errors.New("connection refused")
	mockAuthClient.On("Login", mock.Anything, mock.Anything).Return(nil, connErr)

	token, err := client.Login(context.Background(), "testuser", "Password123!")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "connection error")
	assert.Contains(t, err.Error(), "connection refused")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Register_Success проверяет успешную регистрацию пользователя
func TestClient_Register_Success(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	mockAuthClient.On("Register", mock.Anything, mock.MatchedBy(func(req *pb.RegisterRequest) bool {
		return req.Login == "newuser" && req.Password == "Password123!"
	})).Return(&pb.RegisterResponse{Message: "Registration successful"}, nil)

	err := client.Register(context.Background(), "newuser", "Password123!")

	assert.NoError(t, err)
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Register_AlreadyExists проверяет ошибку существующего пользователя
func TestClient_Register_AlreadyExists(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.AlreadyExists, "user already exists")
	mockAuthClient.On("Register", mock.Anything, mock.MatchedBy(func(req *pb.RegisterRequest) bool {
		return req.Login == "existinguser"
	})).Return(nil, grpcErr)

	err := client.Register(context.Background(), "existinguser", "Password123!")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user already exists")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Register_Unavailable проверяет ошибку недоступности сервера при регистрации
func TestClient_Register_Unavailable(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Unavailable, "server unavailable")
	mockAuthClient.On("Register", mock.Anything, mock.Anything).Return(nil, grpcErr)

	err := client.Register(context.Background(), "newuser", "Password123!")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server is unavailable")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Register_OtherError проверяет другие gRPC ошибки при регистрации
func TestClient_Register_OtherError(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Internal, "database error")
	mockAuthClient.On("Register", mock.Anything, mock.Anything).Return(nil, grpcErr)

	err := client.Register(context.Background(), "newuser", "Password123!")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registration failed")
	assert.Contains(t, err.Error(), "database error")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Register_ConnectionError проверяет ошибку соединения при регистрации
func TestClient_Register_ConnectionError(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	connErr := errors.New("connection refused")
	mockAuthClient.On("Register", mock.Anything, mock.Anything).Return(nil, connErr)

	err := client.Register(context.Background(), "newuser", "Password123!")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection error")
	assert.Contains(t, err.Error(), "connection refused")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Logout_Success проверяет успешный выход пользователя
func TestClient_Logout_Success(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	mockAuthClient.On("Logout", mock.Anything, &emptypb.Empty{}).Return(&pb.LogoutResponse{Message: "Logout successful"}, nil)

	err := client.Logout(context.Background(), "test-token-123")

	assert.NoError(t, err)
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Logout_Unavailable проверяет ошибку недоступности сервера при выходе
func TestClient_Logout_Unavailable(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Unavailable, "server unavailable")
	mockAuthClient.On("Logout", mock.Anything, mock.Anything, mock.Anything).Return(nil, grpcErr)

	err := client.Logout(context.Background(), "test-token-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server is unavailable")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Logout_OtherError проверяет другие gRPC ошибки при выходе
func TestClient_Logout_OtherError(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Internal, "redis error")
	mockAuthClient.On("Logout", mock.Anything, mock.Anything, mock.Anything).Return(nil, grpcErr)

	err := client.Logout(context.Background(), "test-token-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logout failed")
	assert.Contains(t, err.Error(), "redis error")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Logout_ConnectionError проверяет ошибку соединения при выходе
func TestClient_Logout_ConnectionError(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	connErr := errors.New("connection refused")
	mockAuthClient.On("Logout", mock.Anything, mock.Anything, mock.Anything).Return(nil, connErr)

	err := client.Logout(context.Background(), "test-token-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection error")
	assert.Contains(t, err.Error(), "connection refused")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Pull_Success проверяет успешное получение данных с сервера
func TestClient_Pull_Success(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	expectedData := []byte("encrypted-data")
	expectedUpdatedAt := int64(1234567890)
	mockSyncClient.On("Pull", mock.Anything, &emptypb.Empty{}).Return(&pb.SyncResponse{
		Data:      expectedData,
		UpdatedAt: expectedUpdatedAt,
	}, nil)

	data, updatedAt, err := client.Pull(context.Background(), "test-token-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	assert.Equal(t, expectedUpdatedAt, updatedAt)
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Pull_Unauthenticated проверяет ошибку аутентификации при получении данных
func TestClient_Pull_Unauthenticated(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Unauthenticated, "invalid token")
	mockSyncClient.On("Pull", mock.Anything, mock.Anything, mock.Anything).Return(nil, grpcErr)

	data, updatedAt, err := client.Pull(context.Background(), "invalid-token")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "authentication failed")
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Pull_Unavailable проверяет ошибку недоступности сервера при получении данных
func TestClient_Pull_Unavailable(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Unavailable, "server unavailable")
	mockSyncClient.On("Pull", mock.Anything, mock.Anything, mock.Anything).Return(nil, grpcErr)

	data, updatedAt, err := client.Pull(context.Background(), "test-token-123")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "server is unavailable")
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Pull_OtherError проверяет другие gRPC ошибки при получении данных
func TestClient_Pull_OtherError(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Internal, "database error")
	mockSyncClient.On("Pull", mock.Anything, mock.Anything, mock.Anything).Return(nil, grpcErr)

	data, updatedAt, err := client.Pull(context.Background(), "test-token-123")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "pull failed")
	assert.Contains(t, err.Error(), "database error")
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Pull_ConnectionError проверяет ошибку соединения при получении данных
func TestClient_Pull_ConnectionError(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	connErr := errors.New("connection refused")
	mockSyncClient.On("Pull", mock.Anything, mock.Anything, mock.Anything).Return(nil, connErr)

	data, updatedAt, err := client.Pull(context.Background(), "test-token-123")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "connection error")
	assert.Contains(t, err.Error(), "connection refused")
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Push_Success проверяет успешную отправку данных на сервер
func TestClient_Push_Success(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	testData := []byte("encrypted-data")
	expectedUpdatedAt := int64(1234567890)
	mockSyncClient.On("Push", mock.Anything, mock.MatchedBy(func(req *pb.SyncRequest) bool {
		return string(req.Data) == string(testData)
	})).Return(&pb.SyncResponse{
		UpdatedAt: expectedUpdatedAt,
	}, nil)

	updatedAt, err := client.Push(context.Background(), "test-token-123", testData)

	assert.NoError(t, err)
	assert.Equal(t, expectedUpdatedAt, updatedAt)
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Push_Unauthenticated проверяет ошибку аутентификации при отправке данных
func TestClient_Push_Unauthenticated(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Unauthenticated, "invalid token")
	mockSyncClient.On("Push", mock.Anything, mock.Anything, mock.Anything).Return(nil, grpcErr)

	updatedAt, err := client.Push(context.Background(), "invalid-token", []byte("data"))

	assert.Error(t, err)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "authentication failed")
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Push_Unavailable проверяет ошибку недоступности сервера при отправке данных
func TestClient_Push_Unavailable(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Unavailable, "server unavailable")
	mockSyncClient.On("Push", mock.Anything, mock.Anything, mock.Anything).Return(nil, grpcErr)

	updatedAt, err := client.Push(context.Background(), "test-token-123", []byte("data"))

	assert.Error(t, err)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "server is unavailable")
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Push_OtherError проверяет другие gRPC ошибки при отправке данных
func TestClient_Push_OtherError(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	grpcErr := status.Error(codes.Internal, "database error")
	mockSyncClient.On("Push", mock.Anything, mock.Anything, mock.Anything).Return(nil, grpcErr)

	updatedAt, err := client.Push(context.Background(), "test-token-123", []byte("data"))

	assert.Error(t, err)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "push failed")
	assert.Contains(t, err.Error(), "database error")
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Push_ConnectionError проверяет ошибку соединения при отправке данных
func TestClient_Push_ConnectionError(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	connErr := errors.New("connection refused")
	mockSyncClient.On("Push", mock.Anything, mock.Anything, mock.Anything).Return(nil, connErr)

	updatedAt, err := client.Push(context.Background(), "test-token-123", []byte("data"))

	assert.Error(t, err)
	assert.Zero(t, updatedAt)
	assert.Contains(t, err.Error(), "connection error")
	assert.Contains(t, err.Error(), "connection refused")
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Close_NilConn проверяет закрытие с nil соединением
func TestClient_Close_NilConn(t *testing.T) {
	client := &Client{
		conn: nil,
	}

	err := client.Close()

	assert.NoError(t, err)
}

// TestClientInterface_Implementation проверяет, что Client реализует ClientInterface
func TestClientInterface_Implementation(t *testing.T) {
	var _ ClientInterface = (*Client)(nil)
}

// TestClient_Login_ContextCancellation проверяет отмену контекста
func TestClient_Login_ContextCancellation(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Сразу отменяем контекст

	mockAuthClient.On("Login", mock.Anything, mock.Anything).Return(nil, context.Canceled)

	token, err := client.Login(ctx, "testuser", "Password123!")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "connection error")
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Pull_EmptyData проверяет получение пустых данных
func TestClient_Pull_EmptyData(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	mockSyncClient.On("Pull", mock.Anything, &emptypb.Empty{}).Return(&pb.SyncResponse{
		Data:      []byte{},
		UpdatedAt: 0,
	}, nil)

	data, updatedAt, err := client.Pull(context.Background(), "test-token-123")

	assert.NoError(t, err)
	assert.Equal(t, []byte{}, data)
	assert.Equal(t, int64(0), updatedAt)
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Push_EmptyData проверяет отправку пустых данных
func TestClient_Push_EmptyData(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	expectedUpdatedAt := int64(1234567890)
	mockSyncClient.On("Push", mock.Anything, mock.MatchedBy(func(req *pb.SyncRequest) bool {
		return len(req.Data) == 0
	})).Return(&pb.SyncResponse{
		UpdatedAt: expectedUpdatedAt,
	}, nil)

	updatedAt, err := client.Push(context.Background(), "test-token-123", []byte{})

	assert.NoError(t, err)
	assert.Equal(t, expectedUpdatedAt, updatedAt)
	mockSyncClient.AssertExpectations(t)
}

// TestClient_Login_MultipleCalls проверяет множественные вызовы Login
func TestClient_Login_MultipleCalls(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	// Настраиваем мок для двух вызовов
	mockAuthClient.On("Login", mock.Anything, mock.MatchedBy(func(req *pb.LoginRequest) bool {
		return req.Login == "user1"
	})).Return(&pb.LoginResponse{Token: "token1"}, nil)

	mockAuthClient.On("Login", mock.Anything, mock.MatchedBy(func(req *pb.LoginRequest) bool {
		return req.Login == "user2"
	})).Return(&pb.LoginResponse{Token: "token2"}, nil)

	token1, err1 := client.Login(context.Background(), "user1", "Password123!")
	token2, err2 := client.Login(context.Background(), "user2", "Password123!")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, "token1", token1)
	assert.Equal(t, "token2", token2)
	mockAuthClient.AssertExpectations(t)
}

// TestClient_Pull_Push_Integration проверяет интеграцию Pull и Push
func TestClient_Pull_Push_Integration(t *testing.T) {
	mockAuthClient := mocks.NewAuthClient(t)
	mockSyncClient := mocks.NewSyncClient(t)

	client := &Client{
		authClient: mockAuthClient,
		syncClient: mockSyncClient,
	}

	// Pull
	pullData := []byte("original-data")
	pullUpdatedAt := int64(1234567890)
	mockSyncClient.On("Pull", mock.Anything, &emptypb.Empty{}).Return(&pb.SyncResponse{
		Data:      pullData,
		UpdatedAt: pullUpdatedAt,
	}, nil)

	data, updatedAt, err := client.Pull(context.Background(), "test-token-123")
	require.NoError(t, err)
	require.Equal(t, pullData, data)
	require.Equal(t, pullUpdatedAt, updatedAt)

	// Push
	pushData := []byte("updated-data")
	pushUpdatedAt := int64(1234567900)
	mockSyncClient.On("Push", mock.Anything, mock.MatchedBy(func(req *pb.SyncRequest) bool {
		return string(req.Data) == string(pushData)
	})).Return(&pb.SyncResponse{
		UpdatedAt: pushUpdatedAt,
	}, nil)

	newUpdatedAt, err := client.Push(context.Background(), "test-token-123", pushData)
	require.NoError(t, err)
	require.Equal(t, pushUpdatedAt, newUpdatedAt)

	mockSyncClient.AssertExpectations(t)
}