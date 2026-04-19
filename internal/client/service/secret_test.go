package service

import (
	"context"
	"testing"
	"time"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecretService(t *testing.T) {
	service := NewSecretService()
	assert.NotNil(t, service)
	
	// Проверяем, что это правильная реализация
	_, ok := service.(*secretService)
	require.True(t, ok)
}

func TestSecretService_FindSecretByIDPrefix_Success(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Добавляем тестовые данные
	item1 := storage.NewItem("login", "Test Login 1")
	item2 := storage.NewItem("text", "Test Text 2")
	item3 := storage.NewItem("card", "Test Card 3")
	
	// Модифицируем ID для тестирования префиксов
	item1.ID = "abc-def-123"
	item2.ID = "abc-xyz-456"
	item3.ID = "def-ghi-789"
	
	store.Items = append(store.Items, item1, item2, item3)
	
	// Тест 1: Находим уникальный элемент по префиксу
	foundItem, foundItems, err := service.FindSecretByIDPrefix(context.Background(), store, "abc-def")
	assert.NoError(t, err)
	assert.NotNil(t, foundItem)
	assert.Equal(t, "abc-def-123", foundItem.ID)
	assert.Len(t, foundItems, 1)
	assert.Equal(t, item1.ID, foundItems[0].ID)
	
	// Тест 2: Находим несколько элементов по общему префиксу
	foundItem, foundItems, err = service.FindSecretByIDPrefix(context.Background(), store, "abc")
	assert.NoError(t, err)
	assert.Nil(t, foundItem) // Не должно быть уникального элемента
	assert.Len(t, foundItems, 2)
	
	// Проверяем, что найдены правильные элементы
	foundIDs := []string{foundItems[0].ID, foundItems[1].ID}
	assert.Contains(t, foundIDs, "abc-def-123")
	assert.Contains(t, foundIDs, "abc-xyz-456")
	
	// Тест 3: Находим единственный элемент по точному ID
	foundItem, foundItems, err = service.FindSecretByIDPrefix(context.Background(), store, "def-ghi-789")
	assert.NoError(t, err)
	assert.NotNil(t, foundItem)
	assert.Equal(t, "def-ghi-789", foundItem.ID)
	assert.Len(t, foundItems, 1)
	assert.Equal(t, item3.ID, foundItems[0].ID)
}

func TestSecretService_FindSecretByIDPrefix_NotFound(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Добавляем тестовые данные
	item1 := storage.NewItem("login", "Test Login 1")
	item1.ID = "abc-def-123"
	store.Items = append(store.Items, item1)
	
	// Пытаемся найти несуществующий элемент
	foundItem, foundItems, err := service.FindSecretByIDPrefix(context.Background(), store, "xyz")
	assert.Error(t, err)
	assert.Nil(t, foundItem)
	assert.Nil(t, foundItems)
	assert.Contains(t, err.Error(), "не найден")
}

func TestSecretService_FindSecretByIDPrefix_EmptyPrefix(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Пытаемся найти элемент с пустым префиксом
	foundItem, foundItems, err := service.FindSecretByIDPrefix(context.Background(), store, "")
	assert.Error(t, err)
	assert.Nil(t, foundItem)
	assert.Nil(t, foundItems)
	assert.Contains(t, err.Error(), "не может быть пустым")
}

func TestSecretService_AddSecret_Success(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	initialCount := len(store.Items)
	
	// Создаем тестовый элемент
	item := storage.NewItem("login", "Test Login")
	
	// Добавляем элемент
	err := service.AddSecret(context.Background(), store, &item)
	assert.NoError(t, err)
	
	// Проверяем, что элемент добавлен
	assert.Len(t, store.Items, initialCount+1)
	assert.Equal(t, item.ID, store.Items[len(store.Items)-1].ID)
	assert.Equal(t, item.Title, store.Items[len(store.Items)-1].Title)
	
	// Проверяем, что время обновления изменено
	assert.NotEqual(t, store.LastModified, time.Time{}.Format(time.RFC3339))
}

func TestSecretService_AddSecret_EmptyTitle(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	initialCount := len(store.Items)
	
	// Создаем тестовый элемент с пустым заголовком
	item := storage.NewItem("login", "")
	
	// Пытаемся добавить элемент
	err := service.AddSecret(context.Background(), store, &item)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не может быть пустым")
	
	// Проверяем, что элемент не добавлен
	assert.Len(t, store.Items, initialCount)
}

func TestSecretService_UpdateSecret_Success(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Добавляем тестовый элемент
	item := storage.NewItem("login", "Test Login")
	store.Items = append(store.Items, item)
	
	// Модифицируем элемент
	updatedItem := item
	updatedItem.Title = "Updated Title"
	updatedItem.Username = "newuser"
	
	// Обновляем элемент
	err := service.UpdateSecret(context.Background(), store, &updatedItem)
	assert.NoError(t, err)
	
	// Проверяем, что элемент обновлен
	assert.Len(t, store.Items, 1)
	assert.Equal(t, "Updated Title", store.Items[0].Title)
	assert.Equal(t, "newuser", store.Items[0].Username)
	
	// Проверяем, что время обновления изменено
	assert.NotEqual(t, store.LastModified, time.Time{}.Format(time.RFC3339))
}

func TestSecretService_UpdateSecret_EmptyTitle(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Добавляем тестовый элемент
	item := storage.NewItem("login", "Test Login")
	store.Items = append(store.Items, item)
	
	// Пытаемся обновить элемент с пустым заголовком
	updatedItem := item
	updatedItem.Title = ""
	
	// Обновляем элемент
	err := service.UpdateSecret(context.Background(), store, &updatedItem)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не может быть пустым")
	
	// Проверяем, что элемент не изменился
	assert.Len(t, store.Items, 1)
	assert.Equal(t, "Test Login", store.Items[0].Title)
}

func TestSecretService_UpdateSecret_NotFound(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Добавляем тестовый элемент
	item := storage.NewItem("login", "Test Login")
	store.Items = append(store.Items, item)
	
	// Пытаемся обновить несуществующий элемент
	nonExistentItem := storage.NewItem("login", "Non Existent")
	nonExistentItem.ID = uuid.New().String()
	
	err := service.UpdateSecret(context.Background(), store, &nonExistentItem)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
	
	// Проверяем, что исходный элемент не изменился
	assert.Len(t, store.Items, 1)
	assert.Equal(t, "Test Login", store.Items[0].Title)
}

func TestSecretService_DeleteSecret_Success(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Добавляем тестовые элементы
	item1 := storage.NewItem("login", "Test Login 1")
	item2 := storage.NewItem("text", "Test Text 2")
	store.Items = append(store.Items, item1, item2)
	
	// Удаляем первый элемент
	err := service.DeleteSecret(context.Background(), store, item1.ID)
	assert.NoError(t, err)
	
	// Проверяем, что элемент удален
	assert.Len(t, store.Items, 1)
	assert.Equal(t, item2.ID, store.Items[0].ID)
	
	// Проверяем, что время обновления изменено
	assert.NotEqual(t, store.LastModified, time.Time{}.Format(time.RFC3339))
}

func TestSecretService_DeleteSecret_NotFound(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Добавляем тестовый элемент
	item := storage.NewItem("login", "Test Login")
	store.Items = append(store.Items, item)
	
	// Пытаемся удалить несуществующий элемент
	nonExistentID := uuid.New().String()
	err := service.DeleteSecret(context.Background(), store, nonExistentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
	
	// Проверяем, что исходный элемент не удален
	assert.Len(t, store.Items, 1)
	assert.Equal(t, item.ID, store.Items[0].ID)
}

func TestSecretService_FilterSecretsByType_Success(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Добавляем тестовые элементы разных типов
	loginItem := storage.NewItem("login", "Test Login")
	textItem := storage.NewItem("text", "Test Text")
	cardItem := storage.NewItem("card", "Test Card")
	anotherLoginItem := storage.NewItem("login", "Another Login")
	
	store.Items = append(store.Items, loginItem, textItem, cardItem, anotherLoginItem)
	
	// Фильтруем по типу "login"
	filtered, err := service.FilterSecretsByType(context.Background(), store, "login")
	assert.NoError(t, err)
	assert.Len(t, filtered, 2)
	
	// Проверяем, что найдены правильные элементы
	foundTitles := []string{filtered[0].Title, filtered[1].Title}
	assert.Contains(t, foundTitles, "Test Login")
	assert.Contains(t, foundTitles, "Another Login")
	
	// Фильтруем по типу "text"
	filtered, err = service.FilterSecretsByType(context.Background(), store, "text")
	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Test Text", filtered[0].Title)
	
	// Фильтруем по типу "card"
	filtered, err = service.FilterSecretsByType(context.Background(), store, "card")
	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Test Card", filtered[0].Title)
}

func TestSecretService_FilterSecretsByType_InvalidType(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Пытаемся отфильтровать по невалидному типу
	filtered, err := service.FilterSecretsByType(context.Background(), store, "invalid")
	assert.Error(t, err)
	assert.Nil(t, filtered)
	assert.Contains(t, err.Error(), "неизвестный тип данных")
}

func TestSecretService_UpdateStorage(t *testing.T) {
	service := NewSecretService()
	
	// Создаем тестовое хранилище
	store := storage.NewStorage()
	
	// Обновляем хранилище
	err := service.UpdateStorage(context.Background(), store)
	assert.NoError(t, err)
	
	// Проверяем, что время обновления установлено
	assert.NotEmpty(t, store.LastModified)
	
	// Проверяем формат времени (RFC3339)
	_, err = time.Parse(time.RFC3339, store.LastModified)
	assert.NoError(t, err)
}