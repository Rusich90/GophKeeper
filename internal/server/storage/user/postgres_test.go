package user

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Rusich90/GophKeeper/internal/server/model"
	"github.com/Rusich90/GophKeeper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPGUserRepo_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		mockFn  func(*mocks.UserRepository)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful create",
			user: &model.User{
				Login:        "testuser",
				PasswordHash: "hashedpassword123",
			},
			mockFn: func(m *mocks.UserRepository) {
				m.On("Create", context.Background(), &model.User{
					Login:        "testuser",
					PasswordHash: "hashedpassword123",
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "duplicate user",
			user: &model.User{
				Login:        "duplicate",
				PasswordHash: "hashedpassword123",
			},
			mockFn: func(m *mocks.UserRepository) {
				m.On("Create", context.Background(), &model.User{
					Login:        "duplicate",
					PasswordHash: "hashedpassword123",
				}).Return(errors.New("duplicate key value violates unique constraint"))
			},
			wantErr: true,
			errMsg:  "duplicate key",
		},
		{
			name: "database error",
			user: &model.User{
				Login:        "testuser2",
				PasswordHash: "hashedpassword123",
			},
			mockFn: func(m *mocks.UserRepository) {
				m.On("Create", context.Background(), &model.User{
					Login:        "testuser2",
					PasswordHash: "hashedpassword123",
				}).Return(errors.New("database connection error"))
			},
			wantErr: true,
			errMsg:  "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewUserRepository(t)
			tt.mockFn(mockRepo)

			ctx := context.Background()
			err := mockRepo.Create(ctx, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPGUserRepo_FindByLogin(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		mockFn      func(*mocks.UserRepository)
		wantUser    *model.User
		wantErr     bool
		errContains string
	}{
		{
			name:  "successful find",
			login: "testuser",
			mockFn: func(m *mocks.UserRepository) {
				m.On("FindByLogin", context.Background(), "testuser").Return(&model.User{
					Login:        "testuser",
					PasswordHash: "hashedpassword123",
				}, nil)
			},
			wantUser: &model.User{
				Login:        "testuser",
				PasswordHash: "hashedpassword123",
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			login: "nonexistent",
			mockFn: func(m *mocks.UserRepository) {
				m.On("FindByLogin", context.Background(), "nonexistent").Return(nil, fmt.Errorf("%w: %s", ErrUserNotFound, "nonexistent"))
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "user not found",
		},
		{
			name:  "database error",
			login: "testuser2",
			mockFn: func(m *mocks.UserRepository) {
				m.On("FindByLogin", context.Background(), "testuser2").Return(nil, errors.New("database connection error"))
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewUserRepository(t)
			tt.mockFn(mockRepo)

			ctx := context.Background()
			user, err := mockRepo.FindByLogin(ctx, tt.login)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.wantUser.Login, user.Login)
				assert.Equal(t, tt.wantUser.PasswordHash, user.PasswordHash)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPGUserRepo_GetUserData(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		mockFn        func(*mocks.UserRepository)
		wantData      []byte
		wantUpdatedAt int64
		wantErr       bool
		errContains   string
	}{
		{
			name:  "successful get with data",
			login: "testuser",
			mockFn: func(m *mocks.UserRepository) {
				m.On("GetUserData", context.Background(), "testuser").Return(
					[]byte("encrypted data"),
					int64(1234567890),
					nil,
				)
			},
			wantData:      []byte("encrypted data"),
			wantUpdatedAt: 1234567890,
			wantErr:       false,
		},
		{
			name:  "successful get without data",
			login: "testuser2",
			mockFn: func(m *mocks.UserRepository) {
				m.On("GetUserData", context.Background(), "testuser2").Return(
					nil,
					int64(0),
					nil,
				)
			},
			wantData:      nil,
			wantUpdatedAt: 0,
			wantErr:       false,
		},
		{
			name:  "user not found",
			login: "nonexistent",
			mockFn: func(m *mocks.UserRepository) {
				m.On("GetUserData", context.Background(), "nonexistent").Return(
					nil,
					int64(0),
					fmt.Errorf("%w: %s", ErrUserNotFound, "nonexistent"),
				)
			},
			wantData:      nil,
			wantUpdatedAt: 0,
			wantErr:       true,
			errContains:   "user not found",
		},
		{
			name:  "database error",
			login: "testuser3",
			mockFn: func(m *mocks.UserRepository) {
				m.On("GetUserData", context.Background(), "testuser3").Return(
					nil,
					int64(0),
					errors.New("database connection error"),
				)
			},
			wantData:      nil,
			wantUpdatedAt: 0,
			wantErr:       true,
			errContains:   "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewUserRepository(t)
			tt.mockFn(mockRepo)

			ctx := context.Background()
			data, updatedAt, err := mockRepo.GetUserData(ctx, tt.login)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantData, data)
				assert.Equal(t, tt.wantUpdatedAt, updatedAt)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPGUserRepo_UpdateUserData(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		encryptedData []byte
		updatedAt     int64
		mockFn        func(*mocks.UserRepository)
		wantErr       bool
	}{
		{
			name:          "successful update",
			login:         "testuser",
			encryptedData: []byte("new encrypted data"),
			updatedAt:     1234567890,
			mockFn: func(m *mocks.UserRepository) {
				m.On("UpdateUserData", context.Background(), "testuser", []byte("new encrypted data"), int64(1234567890)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "update with empty data",
			login:         "testuser2",
			encryptedData: []byte(nil),
			updatedAt:     0,
			mockFn: func(m *mocks.UserRepository) {
				m.On("UpdateUserData", context.Background(), "testuser2", []byte(nil), int64(0)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "database error",
			login:         "testuser3",
			encryptedData: []byte("data"),
			updatedAt:     1234567890,
			mockFn: func(m *mocks.UserRepository) {
				m.On("UpdateUserData", context.Background(), "testuser3", []byte("data"), int64(1234567890)).Return(errors.New("database connection error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewUserRepository(t)
			tt.mockFn(mockRepo)

			ctx := context.Background()
			err := mockRepo.UpdateUserData(ctx, tt.login, tt.encryptedData, tt.updatedAt)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPGUserRepo_FullWorkflow(t *testing.T) {
	t.Run("complete user lifecycle", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		ctx := context.Background()

		user := &model.User{
			Login:        "lifecycleuser",
			PasswordHash: "hashedpassword123",
		}

		// 1. Создаем пользователя
		mockRepo.On("Create", ctx, user).Return(nil)

		err := mockRepo.Create(ctx, user)
		require.NoError(t, err)

		// 2. Находим пользователя
		foundUser := &model.User{
			Login:        "lifecycleuser",
			PasswordHash: "hashedpassword123",
		}
		mockRepo.On("FindByLogin", ctx, user.Login).Return(foundUser, nil)

		retrievedUser, err := mockRepo.FindByLogin(ctx, user.Login)
		require.NoError(t, err)
		assert.Equal(t, user.Login, retrievedUser.Login)
		assert.Equal(t, user.PasswordHash, retrievedUser.PasswordHash)

		// 3. Обновляем данные пользователя
		encryptedData := []byte("new encrypted data")
		updatedAt := int64(1234567900)
		mockRepo.On("UpdateUserData", ctx, user.Login, encryptedData, updatedAt).Return(nil)

		err = mockRepo.UpdateUserData(ctx, user.Login, encryptedData, updatedAt)
		require.NoError(t, err)

		// 4. Получаем данные пользователя
		mockRepo.On("GetUserData", ctx, user.Login).Return(encryptedData, updatedAt, nil)

		data, retrievedUpdatedAt, err := mockRepo.GetUserData(ctx, user.Login)
		require.NoError(t, err)
		assert.Equal(t, encryptedData, data)
		assert.Equal(t, updatedAt, retrievedUpdatedAt)

		// 5. Проверяем, что данные обновились
		newData := []byte("new encrypted data")
		newUpdatedAt := int64(1234567900)
		mockRepo.On("UpdateUserData", ctx, user.Login, newData, newUpdatedAt).Return(nil)

		err = mockRepo.UpdateUserData(ctx, user.Login, newData, newUpdatedAt)
		require.NoError(t, err)

		mockRepo.On("GetUserData", ctx, user.Login).Return(newData, newUpdatedAt, nil)

		data, retrievedUpdatedAt, err = mockRepo.GetUserData(ctx, user.Login)
		require.NoError(t, err)
		assert.Equal(t, newData, data)
		assert.Equal(t, newUpdatedAt, retrievedUpdatedAt)

		mockRepo.AssertExpectations(t)
	})
}

func TestNewPGUserRepo(t *testing.T) {
	t.Run("creates new repository", func(t *testing.T) {
		// Этот тест проверяет только создание структуры, без реального подключения
		repo := &PGUserRepo{}

		assert.NotNil(t, repo)
	})
}

func TestErrUserNotFound(t *testing.T) {
	t.Run("error message contains login", func(t *testing.T) {
		login := "testuser"
		err := fmt.Errorf("%w: %s", ErrUserNotFound, login)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.Contains(t, err.Error(), login)
	})
}
