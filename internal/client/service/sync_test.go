package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Rusich90/GophKeeper/internal/client/crypto"
	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/Rusich90/GophKeeper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewSyncService(t *testing.T) {
	mockGrpcClient := mocks.NewClientInterface(t)
	mockSessionStorage := storage.NewSessionStorage()

	// Создаем сервис с моками
	service := NewSyncService(mockGrpcClient, mockSessionStorage)

	// Проверяем, что сервис создан
	assert.NotNil(t, service)

	// Преобразуем к конкретному типу для проверки полей
	syncService, ok := service.(*syncService)
	require.True(t, ok)

	assert.Equal(t, mockGrpcClient, syncService.grpcClient)
	assert.Equal(t, mockSessionStorage, syncService.sessionStorage)
	assert.NotEmpty(t, syncService.dataFilePath)
}

func TestSyncService_Sync_Success(t *testing.T) {
	// Создаем временные файлы для теста
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "data.enc")

	// Создаем моки
	mockGrpcClient := mocks.NewClientInterface(t)
	mockSessionStorage := storage.NewSessionStorage()

	// Создаем сервис с моками
	service := &syncService{
		grpcClient:     mockGrpcClient,
		sessionStorage: mockSessionStorage,
		dataFilePath:   dataFile,
	}

	// Создаем тестовую сессию
	session := &storage.Session{
		Token:      "test-token",
		Key:        crypto.GenerateEncryptionKey("test-password"),
		LastSyncTS: time.Now().Unix() - 1000, // Старая синхронизация
	}

	// Создаем тестовые данные
	localStorage := storage.NewStorage()
	localItem := storage.NewItem("login", "Local Item")
	localStorage.Items = append(localStorage.Items, localItem)

	// Сериализуем и шифруем локальные данные
	localData, err := json.Marshal(localStorage)
	require.NoError(t, err)
	_, err = crypto.Encrypt(localData, session.Key)
	require.NoError(t, err)

	// Создаем серверные данные
	serverStorage := storage.NewStorage()
	serverItem := storage.NewItem("text", "Server Item")
	serverItem.UpdatedAt = time.Now().Add(time.Hour).UTC().Format(time.RFC3339) // Более новая дата
	serverStorage.Items = append(serverStorage.Items, serverItem)

	// Сериализуем и шифруем серверные данные
	serverData, err := json.Marshal(serverStorage)
	require.NoError(t, err)
	encryptedServerData, err := crypto.Encrypt(serverData, session.Key)
	require.NoError(t, err)

	// Настраиваем моки
	mockGrpcClient.On("Pull", mock.Anything, session.Token).Return(encryptedServerData, time.Now().Unix(), nil)
	
	// Используем mock.Anything для данных, так как шифрование AES-GCM генерирует разные данные каждый раз
	mockGrpcClient.On("Push", mock.Anything, session.Token, mock.AnythingOfType("[]uint8")).Return(time.Now().Unix(), nil)

	// Выполняем тест
	err = service.Sync(context.Background(), session)
	assert.NoError(t, err)

	// Проверяем, что моки были вызваны
	mockGrpcClient.AssertExpectations(t)

	// Проверяем, что файл данных был создан
	_, err = os.Stat(dataFile)
	assert.NoError(t, err)

	// Проверяем, что сессия была обновлена
	assert.Greater(t, session.LastSyncTS, int64(0))
}

func TestSyncService_Sync_PullError(t *testing.T) {
	// Создаем временные файлы для теста
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "data.enc")

	// Создаем моки
	mockGrpcClient := mocks.NewClientInterface(t)
	mockSessionStorage := storage.NewSessionStorage()

	// Создаем сервис с моками
	service := &syncService{
		grpcClient:     mockGrpcClient,
		sessionStorage: mockSessionStorage,
		dataFilePath:   dataFile,
	}

	// Создаем тестовую сессию
	session := &storage.Session{
		Token: "test-token",
		Key:   crypto.GenerateEncryptionKey("test-password"),
	}

	// Настраиваем моки
	pullErr := errors.New("pull failed")
	mockGrpcClient.On("Pull", mock.Anything, session.Token).Return([]byte{}, int64(0), pullErr)

	// Выполняем тест
	err := service.Sync(context.Background(), session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка получения секретов с сервера")
	assert.Contains(t, err.Error(), "pull failed")

	// Проверяем, что моки были вызваны
	mockGrpcClient.AssertExpectations(t)
	// Убедимся, что Push не был вызван
	mockGrpcClient.AssertNotCalled(t, "Push")
}

func TestSyncService_Sync_PushError(t *testing.T) {
	// Создаем временные файлы для теста
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "data.enc")

	// Создаем моки
	mockGrpcClient := mocks.NewClientInterface(t)
	mockSessionStorage := storage.NewSessionStorage()

	// Создаем сервис с моками
	service := &syncService{
		grpcClient:     mockGrpcClient,
		sessionStorage: mockSessionStorage,
		dataFilePath:   dataFile,
	}

	// Создаем тестовую сессию
	session := &storage.Session{
		Token:      "test-token",
		Key:        crypto.GenerateEncryptionKey("test-password"),
		LastSyncTS: time.Now().Unix() - 1000, // Старая синхронизация
	}

	// Создаем серверные данные
	serverStorage := storage.NewStorage()
	serverItem := storage.NewItem("text", "Server Item")
	serverStorage.Items = append(serverStorage.Items, serverItem)

	// Сериализуем и шифруем серверные данные
	serverData, err := json.Marshal(serverStorage)
	require.NoError(t, err)
	encryptedServerData, err := crypto.Encrypt(serverData, session.Key)
	require.NoError(t, err)

	// Настраиваем моки
	mockGrpcClient.On("Pull", mock.Anything, session.Token).Return(encryptedServerData, time.Now().Unix(), nil)
	
	// Используем mock.Anything для данных, так как шифрование AES-GCM генерирует разные данные каждый раз
	pushErr := errors.New("push failed")
	mockGrpcClient.On("Push", mock.Anything, session.Token, mock.AnythingOfType("[]uint8")).Return(int64(0), pushErr)

	// Выполняем тест
	err = service.Sync(context.Background(), session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка отправки секретов на сервер")
	assert.Contains(t, err.Error(), "push failed")

	// Проверяем, что моки были вызваны
	mockGrpcClient.AssertExpectations(t)
}

func TestSyncService_LoadLocalSecrets_FileNotFound(t *testing.T) {
	// Создаем временные файлы для теста
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "non-existent", "data.enc")

	// Создаем сервис с несуществующим путем к файлу
	service := &syncService{
		dataFilePath: dataFile,
	}

	// Выполняем тест
	key := crypto.GenerateEncryptionKey("test-password")
	result, err := service.loadLocalSecrets(key)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Items)
}

func TestSyncService_LoadLocalSecrets_DecryptionError(t *testing.T) {
	// Создаем временные файлы для теста
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "data.enc")

	// Создаем сервис
	service := &syncService{
		dataFilePath: dataFile,
	}

	// Создаем файл с некорректными данными
	err := os.WriteFile(dataFile, []byte("invalid data"), 0600)
	require.NoError(t, err)

	// Выполняем тест
	key := crypto.GenerateEncryptionKey("test-password")
	result, err := service.loadLocalSecrets(key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка расшифровки секретов")
	assert.Nil(t, result)
}

func TestSyncService_SaveLocalSecrets_Success(t *testing.T) {
	// Создаем временные файлы для теста
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "data.enc")

	// Создаем сервис
	service := &syncService{
		dataFilePath: dataFile,
	}

	// Создаем тестовые данные
	store := storage.NewStorage()
	item := storage.NewItem("login", "Test Item")
	store.Items = append(store.Items, item)

	// Выполняем тест
	key := crypto.GenerateEncryptionKey("test-password")
	err := service.saveLocalSecrets(store, key)
	assert.NoError(t, err)

	// Проверяем, что файл был создан
	_, err = os.Stat(dataFile)
	assert.NoError(t, err)

	// Проверяем, что файл можно прочитать и расшифровать
	encryptedData, err := os.ReadFile(dataFile)
	require.NoError(t, err)

	decryptedData, err := crypto.Decrypt(encryptedData, key)
	require.NoError(t, err)

	var loadedStore storage.Storage
	err = json.Unmarshal(decryptedData, &loadedStore)
	assert.NoError(t, err)
	assert.Len(t, loadedStore.Items, 1)
	assert.Equal(t, item.Title, loadedStore.Items[0].Title)
}

func TestSyncService_MergeSecrets_ServerNewer(t *testing.T) {
	service := &syncService{}

	// Создаем локальные данные
	localStorage := storage.NewStorage()
	localItem := storage.NewItem("login", "Local Item")
	localItem.UpdatedAt = time.Now().Add(-time.Hour).UTC().Format(time.RFC3339) // Старая дата
	localStorage.Items = append(localStorage.Items, localItem)

	// Создаем серверные данные
	serverStorage := storage.NewStorage()
	serverItem := localItem // Тот же ID
	serverItem.Title = "Updated Server Item"
	serverItem.UpdatedAt = time.Now().Add(time.Hour).UTC().Format(time.RFC3339) // Более новая дата
	serverStorage.Items = append(serverStorage.Items, serverItem)

	// Сериализуем и шифруем серверные данные
	serverData, err := json.Marshal(serverStorage)
	require.NoError(t, err)
	key := crypto.GenerateEncryptionKey("test-password")
	encryptedServerData, err := crypto.Encrypt(serverData, key)
	require.NoError(t, err)

	// Выполняем тест
	lastSyncTS := time.Now().Unix() - 1000
	serverTime := time.Now().Unix()
	result, err := service.mergeSecrets(localStorage, encryptedServerData, key, lastSyncTS, serverTime)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "Updated Server Item", result.Items[0].Title)
}

func TestSyncService_MergeSecrets_LocalNewer(t *testing.T) {
	service := &syncService{}

	// Создаем локальные данные
	localStorage := storage.NewStorage()
	localItem := storage.NewItem("login", "Local Item")
	localItem.UpdatedAt = time.Now().Add(time.Hour).UTC().Format(time.RFC3339) // Более новая дата
	localStorage.Items = append(localStorage.Items, localItem)

	// Создаем серверные данные
	serverStorage := storage.NewStorage()
	serverItem := localItem // Тот же ID
	serverItem.Title = "Old Server Item"
	serverItem.UpdatedAt = time.Now().Add(-time.Hour).UTC().Format(time.RFC3339) // Старая дата
	serverStorage.Items = append(serverStorage.Items, serverItem)

	// Сериализуем и шифруем серверные данные
	serverData, err := json.Marshal(serverStorage)
	require.NoError(t, err)
	key := crypto.GenerateEncryptionKey("test-password")
	encryptedServerData, err := crypto.Encrypt(serverData, key)
	require.NoError(t, err)

	// Выполняем тест
	lastSyncTS := time.Now().Unix() - 1000
	serverTime := time.Now().Unix()
	result, err := service.mergeSecrets(localStorage, encryptedServerData, key, lastSyncTS, serverTime)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "Local Item", result.Items[0].Title) // Локальный элемент должен победить
}

func TestSyncService_MergeSecrets_NoUpdateNeeded(t *testing.T) {
	service := &syncService{}

	// Создаем локальные данные
	localStorage := storage.NewStorage()
	localItem := storage.NewItem("login", "Local Item")
	localStorage.Items = append(localStorage.Items, localItem)

	// Выполняем тест с ситуацией, когда серверные данные не новее
	lastSyncTS := time.Now().Unix()
	serverTime := time.Now().Unix() - 1000 // Серверное время старше
	result, err := service.mergeSecrets(localStorage, []byte{}, "", lastSyncTS, serverTime)
	assert.NoError(t, err)
	assert.Equal(t, localStorage, result) // Должны вернуть локальные данные без изменений
}