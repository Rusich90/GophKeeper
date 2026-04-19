package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Rusich90/GophKeeper/internal/client/crypto"
	"github.com/Rusich90/GophKeeper/internal/client/grpc"
	"github.com/Rusich90/GophKeeper/internal/client/storage"
)

// SyncService определяет интерфейс для работы с синхронизацией
type SyncService interface {
	Sync(ctx context.Context, session *storage.Session) error
}

// syncService реализация SyncService
type syncService struct {
	grpcClient     grpc.ClientInterface
	sessionStorage *storage.SessionStorage
	dataFilePath   string
}

// NewSyncService создает новый экземпляр сервиса синхронизации
func NewSyncService(grpcClient grpc.ClientInterface, sessionStorage *storage.SessionStorage) SyncService {
	configDir, err := os.UserHomeDir()
	if err != nil {
		configDir = ".gophkeeper"
	}

	return &syncService{
		grpcClient:     grpcClient,
		sessionStorage: sessionStorage,
		dataFilePath:   filepath.Join(configDir, ".gophkeeper", "data.enc"),
	}
}

// Sync выполняет полную синхронизацию секретов с сервером
func (s *syncService) Sync(ctx context.Context, session *storage.Session) error {
	// Шаг 1: Инициализация - загружаем локальные секреты
	localSecrets, err := s.loadLocalSecrets(session.Key)
	if err != nil {
		return fmt.Errorf("ошибка загрузки локальных секретов: %w", err)
	}

	// Шаг 2: Pull - получаем секреты с сервера
	serverData, serverTime, err := s.grpcClient.Pull(ctx, session.Token)
	if err != nil {
		return fmt.Errorf("ошибка получения секретов с сервера: %w", err)
	}

	// Шаг 3: Merge - объединяем секреты
	mergedSecrets, err := s.mergeSecrets(localSecrets, serverData, session.Key, session.LastSyncTS, serverTime)
	if err != nil {
		return fmt.Errorf("ошибка объединения секретов: %w", err)
	}

	// Шаг 4: Push - отправляем секреты на сервер
	mergedJSON, err := json.Marshal(mergedSecrets)
	if err != nil {
		return fmt.Errorf("ошибка сериализации секретов: %w", err)
	}

	encryptedData, err := crypto.Encrypt(mergedJSON, session.Key)
	if err != nil {
		return fmt.Errorf("ошибка шифрования секретов: %w", err)
	}

	newServerTime, err := s.grpcClient.Push(ctx, session.Token, encryptedData)
	if err != nil {
		return fmt.Errorf("ошибка отправки секретов на сервер: %w", err)
	}

	// Шаг 5: Финализация - сохраняем локальные секреты
	if err := s.saveLocalSecrets(mergedSecrets, session.Key); err != nil {
		return fmt.Errorf("ошибка сохранения локальных секретов: %w", err)
	}

	// Обновляем last_sync_ts в сессии
	session.LastSyncTS = newServerTime
	if err := s.sessionStorage.SaveSession(session); err != nil {
		return fmt.Errorf("ошибка обновления сессии: %w", err)
	}

	return nil
}

// loadLocalSecrets загружает и расшифровывает локальные секреты
func (s *syncService) loadLocalSecrets(key string) (*storage.Storage, error) {
	data, err := os.ReadFile(s.dataFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Если файл не существует, возвращаем пустое хранилище
			return storage.NewStorage(), nil
		}
		return nil, fmt.Errorf("ошибка чтения файла секретов: %w", err)
	}

	decryptedData, err := crypto.Decrypt(data, key)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки секретов: %w", err)
	}

	var store storage.Storage
	if err := json.Unmarshal(decryptedData, &store); err != nil {
		return nil, fmt.Errorf("ошибка парсинга секретов: %w", err)
	}

	return &store, nil
}

// saveLocalSecrets шифрует и сохраняет локальные секреты
func (s *syncService) saveLocalSecrets(store *storage.Storage, key string) error {
	data, err := json.Marshal(store)
	if err != nil {
		return fmt.Errorf("ошибка сериализации секретов: %w", err)
	}

	encryptedData, err := crypto.Encrypt(data, key)
	if err != nil {
		return fmt.Errorf("ошибка шифрования секретов: %w", err)
	}

	// Убеждаемся, что директория существует
	dataDir := filepath.Dir(s.dataFilePath)
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("ошибка создания директории: %w", err)
	}

	if err := os.WriteFile(s.dataFilePath, encryptedData, 0600); err != nil {
		return fmt.Errorf("ошибка записи файла секретов: %w", err)
	}

	return nil
}

// mergeSecrets объединяет локальные и серверные секреты
func (s *syncService) mergeSecrets(local *storage.Storage, serverData []byte, key string, lastSyncTS, serverTime int64) (*storage.Storage, error) {
	// Если секреты на сервере не новее, возвращаем локальные секреты
	if lastSyncTS >= serverTime {
		return local, nil
	}

	// Расшифровываем секреты с сервера
	decryptedServerData, err := crypto.Decrypt(serverData, key)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки серверных секретов: %w", err)
	}

	var server storage.Storage
	if err := json.Unmarshal(decryptedServerData, &server); err != nil {
		return nil, fmt.Errorf("ошибка парсинга серверных секретов: %w", err)
	}

	// Создаем мапу для быстрого поиска по ID
	localMap := make(map[string]storage.Item)
	for _, item := range local.Items {
		localMap[item.ID] = item
	}

	serverMap := make(map[string]storage.Item)
	for _, item := range server.Items {
		serverMap[item.ID] = item
	}

	// Объединяем секреты - оптимизированный алгоритм O(n)
	mergedMap := make(map[string]storage.Item)

	// Добавляем все локальные записи в мапу результата
	for id, localItem := range localMap {
		mergedMap[id] = localItem
	}

	// Обновляем или добавляем записи с сервера
	for id, serverItem := range serverMap {
		if localItem, exists := localMap[id]; exists {
			// Запись существует локально и на сервере - выбираем более новую
			localTime, err := time.Parse(time.RFC3339, localItem.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("ошибка парсинга времени локальной записи: %w", err)
			}

			serverTimeParsed, err := time.Parse(time.RFC3339, serverItem.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("ошибка парсинга времени серверной записи: %w", err)
			}

			// Выбираем более новую версию записи
			if serverTimeParsed.After(localTime) {
				mergedMap[id] = serverItem
			}
			// Иначе оставляем локальную версию (уже в мапе)
		} else {
			// Запись существует только на сервере - добавляем
			mergedMap[id] = serverItem
		}
	}

	// Конвертируем мапу в слайс для результата
	merged := storage.NewStorage()
	merged.Items = make([]storage.Item, 0, len(mergedMap))
	for _, item := range mergedMap {
		merged.Items = append(merged.Items, item)
	}

	// Обновляем время последнего изменения
	merged.LastModified = time.Now().UTC().Format(time.RFC3339)

	return merged, nil
}
