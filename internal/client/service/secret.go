package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
)

// SecretService определяет интерфейс для работы с секретами
type SecretService interface {
	// FindSecretByIDPrefix находит секрет по префиксу ID
	FindSecretByIDPrefix(ctx context.Context, storageData *storage.Storage, idPrefix string) (*storage.Item, []storage.Item, error)

	// AddSecret добавляет новый секрет в хранилище
	AddSecret(ctx context.Context, storageData *storage.Storage, secret *storage.Item) error

	// UpdateSecret обновляет существующий секрет
	UpdateSecret(ctx context.Context, storageData *storage.Storage, secret *storage.Item) error

	// DeleteSecret удаляет секрет по ID
	DeleteSecret(ctx context.Context, storageData *storage.Storage, secretID string) error

	// FilterSecretsByType фильтрует секреты по типу
	FilterSecretsByType(ctx context.Context, storageData *storage.Storage, secretType string) ([]storage.Item, error)

	// UpdateStorage обновляет время модификации хранилища
	UpdateStorage(ctx context.Context, storageData *storage.Storage) error
}

// secretService реализация SecretService
type secretService struct{}

// NewSecretService создает новый экземпляр сервиса секретов
func NewSecretService() SecretService {
	return &secretService{}
}

// FindSecretByIDPrefix находит секрет по префиксу ID
func (s *secretService) FindSecretByIDPrefix(ctx context.Context, storageData *storage.Storage, idPrefix string) (*storage.Item, []storage.Item, error) {
	if idPrefix == "" {
		return nil, nil, fmt.Errorf("префикс ID не может быть пустым")
	}

	var foundItems []storage.Item
	var foundIndex = -1

	for i, item := range storageData.Items {
		if strings.HasPrefix(item.ID, idPrefix) {
			foundItems = append(foundItems, item)
			if foundIndex == -1 {
				foundIndex = i
			}
		}
	}

	switch len(foundItems) {
	case 0:
		return nil, nil, fmt.Errorf("секрет с ID '%s' не найден", idPrefix)
	case 1:
		return &storageData.Items[foundIndex], foundItems, nil
	default:
		return nil, foundItems, nil
	}
}

// AddSecret добавляет новый секрет
func (s *secretService) AddSecret(ctx context.Context, storageData *storage.Storage, secret *storage.Item) error {
	// Валидация
	if secret.Title == "" {
		return fmt.Errorf("название не может быть пустым")
	}

	// Добавляем секрет
	storageData.Items = append(storageData.Items, *secret)

	// Обновляем время последнего изменения
	storageData.LastModified = time.Now().UTC().Format(time.RFC3339)

	return nil
}

// UpdateSecret обновляет существующий секрет
func (s *secretService) UpdateSecret(ctx context.Context, storageData *storage.Storage, secret *storage.Item) error {
	// Валидация
	if secret.Title == "" {
		return fmt.Errorf("название не может быть пустым")
	}

	// Ищем секрет по ID и обновляем его
	for i, item := range storageData.Items {
		if item.ID == secret.ID {
			storageData.Items[i] = *secret
			
			// Обновляем время последнего изменения
			storageData.LastModified = time.Now().UTC().Format(time.RFC3339)
			
			return nil
		}
	}

	return fmt.Errorf("секрет с ID '%s' не найден", secret.ID)
}

// DeleteSecret удаляет секрет
func (s *secretService) DeleteSecret(ctx context.Context, storageData *storage.Storage, secretID string) error {
	for i, item := range storageData.Items {
		if item.ID == secretID {
			// Удаляем секрет
			storageData.Items = append(storageData.Items[:i], storageData.Items[i+1:]...)

			// Обновляем время последнего изменения
			storageData.LastModified = time.Now().UTC().Format(time.RFC3339)

			return nil
		}
	}

	return fmt.Errorf("секрет с ID '%s' не найден", secretID)
}

// FilterSecretsByType фильтрует секреты по типу
func (s *secretService) FilterSecretsByType(ctx context.Context, storageData *storage.Storage, secretType string) ([]storage.Item, error) {
	// Валидация типа
	validTypes := map[string]bool{
		"login": true,
		"text":  true,
		"card":  true,
	}

	if !validTypes[secretType] {
		return nil, fmt.Errorf("неизвестный тип данных: %s", secretType)
	}

	// Фильтруем секреты
	var filteredItems []storage.Item
	for _, item := range storageData.Items {
		if item.Type == secretType {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems, nil
}

// UpdateStorage обновляет время модификации хранилища
func (s *secretService) UpdateStorage(ctx context.Context, storageData *storage.Storage) error {
	storageData.LastModified = time.Now().UTC().Format(time.RFC3339)
	return nil
}
