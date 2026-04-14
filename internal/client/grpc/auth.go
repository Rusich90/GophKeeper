package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Rusich90/GophKeeper/pkg/pb"
)

// Client представляет gRPC клиент для работы с сервером
type Client struct {
	conn   *grpc.ClientConn
	client pb.AuthClient
}

// NewClient создает новый gRPC клиент и подключается к серверу
func NewClient(serverAddr string) (*Client, error) {
	// Создаем соединение с блокирующим подключением
	// WithBlock заставляет клиента ждать фактического подключения
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewAuthClient(conn),
	}, nil
}

// Login выполняет вход пользователя через gRPC
func (c *Client) Login(ctx context.Context, login, password string) (string, error) {
	req := &pb.LoginRequest{
		Login:    login,
		Password: password,
	}

	// Вызываем RPC метод
	resp, err := c.client.Login(ctx, req)
	if err != nil {
		// Пытаемся извлечь статус ошибки от gRPC
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				return "", fmt.Errorf("invalid login or password")
			case codes.Unavailable:
				return "", fmt.Errorf("server is unavailable")
			default:
				return "", fmt.Errorf("login failed: %s", st.Message())
			}
		}
		return "", fmt.Errorf("connection error: %w", err)
	}

	return resp.Token, nil
}

// Register выполняет регистрацию нового пользователя через gRPC
func (c *Client) Register(ctx context.Context, login, password string) error {
	req := &pb.RegisterRequest{
		Login:    login,
		Password: password,
	}

	_, err := c.client.Register(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return fmt.Errorf("user already exists")
			case codes.Unavailable:
				return fmt.Errorf("server is unavailable")
			default:
				return fmt.Errorf("registration failed: %s", st.Message())
			}
		}
		return fmt.Errorf("connection error: %w", err)
	}

	return nil
}

// Logout выполняет выход пользователя через gRPC
func (c *Client) Logout(ctx context.Context, token string) error {
	// Добавляем токен в метаданные для middleware
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.client.Logout(ctx, &emptypb.Empty{})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unavailable:
				return fmt.Errorf("server is unavailable")
			default:
				return fmt.Errorf("logout failed: %s", st.Message())
			}
		}
		return fmt.Errorf("connection error: %w", err)
	}

	return nil
}

// Close закрывает соединение с gRPC сервером
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
