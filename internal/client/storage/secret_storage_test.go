package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Rusich90/GophKeeper/internal/client/crypto"
)

func TestNewSecretStorage(t *testing.T) {
	ss := NewSecretStorage()
	if ss == nil {
		t.Fatal("NewSecretStorage() вернул nil")
	}

	if ss.dataFile == "" {
		t.Error("NewSecretStorage() dataFile пустой")
	}
}

func TestSecretStorage_SaveAndLoad(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.enc")

	ss := &SecretStorage{dataFile: testFile}
	key := crypto.GenerateEncryptionKey("testPassword123")

	// Создаем тестовое хранилище
	storage := NewStorage()
	item := NewItem("login", "Тестовая запись")
	item.Username = "testuser"
	item.Password = "testpass"
	storage.Items = append(storage.Items, item)

	// Сохраняем
	err := ss.Save(storage, key)
	if err != nil {
		t.Fatalf("Save() ошибка = %v", err)
	}

	// Проверяем, что файл создан
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Save() не создал файл")
	}

	// Загружаем
	loaded, err := ss.Load(key)
	if err != nil {
		t.Fatalf("Load() ошибка = %v", err)
	}

	// Проверяем данные
	if len(loaded.Items) != 1 {
		t.Errorf("Load() вернул %d элементов, хотим 1", len(loaded.Items))
	}

	if loaded.Items[0].Title != "Тестовая запись" {
		t.Errorf("Load() Title = %v, хотим 'Тестовая запись'", loaded.Items[0].Title)
	}

	if loaded.Items[0].Username != "testuser" {
		t.Errorf("Load() Username = %v, хотим 'testuser'", loaded.Items[0].Username)
	}

	if loaded.Items[0].Password != "testpass" {
		t.Errorf("Load() Password = %v, хотим 'testpass'", loaded.Items[0].Password)
	}
}

func TestSecretStorage_Load_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "nonexistent.enc")

	ss := &SecretStorage{dataFile: testFile}
	key := crypto.GenerateEncryptionKey("testPassword123")

	// Загружаем несуществующий файл
	loaded, err := ss.Load(key)
	if err != nil {
		t.Fatalf("Load() ошибка = %v", err)
	}

	// Должен вернуть пустое хранилище
	if loaded == nil {
		t.Fatal("Load() вернул nil вместо пустого хранилища")
	}

	if len(loaded.Items) != 0 {
		t.Errorf("Load() вернул %d элементов, хотим 0", len(loaded.Items))
	}
}

func TestSecretStorage_Load_WrongKey(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.enc")

	ss := &SecretStorage{dataFile: testFile}
	correctKey := crypto.GenerateEncryptionKey("correctPassword")
	wrongKey := crypto.GenerateEncryptionKey("wrongPassword")

	// Создаем и сохраняем хранилище
	storage := NewStorage()
	item := NewItem("text", "Секретный текст")
	item.Content = "секретные данные"
	storage.Items = append(storage.Items, item)

	err := ss.Save(storage, correctKey)
	if err != nil {
		t.Fatalf("Save() ошибка = %v", err)
	}

	// Пытаемся загрузить с неправильным ключом
	_, err = ss.Load(wrongKey)
	if err == nil {
		t.Error("Load() с неправильным ключом должен возвращать ошибку")
	}
}

func TestSecretStorage_Save_MultipleItems(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.enc")

	ss := &SecretStorage{dataFile: testFile}
	key := crypto.GenerateEncryptionKey("testPassword123")

	// Создаем хранилище с несколькими элементами разных типов
	storage := NewStorage()

	// Логин
	loginItem := NewItem("login", "Google")
	loginItem.Username = "user@gmail.com"
	loginItem.Password = "googlepass123"
	storage.Items = append(storage.Items, loginItem)

	// Текст
	textItem := NewItem("text", "Заметки")
	textItem.Content = "Важные заметки"
	storage.Items = append(storage.Items, textItem)

	// Карта
	cardItem := NewItem("card", "Visa")
	cardItem.Number = "4111111111111111"
	cardItem.ExpiryMonth = "12"
	cardItem.ExpiryYear = "25"
	cardItem.CVV = "123"
	storage.Items = append(storage.Items, cardItem)

	// Сохраняем
	err := ss.Save(storage, key)
	if err != nil {
		t.Fatalf("Save() ошибка = %v", err)
	}

	// Загружаем
	loaded, err := ss.Load(key)
	if err != nil {
		t.Fatalf("Load() ошибка = %v", err)
	}

	// Проверяем количество элементов
	if len(loaded.Items) != 3 {
		t.Fatalf("Load() вернул %d элементов, хотим 3", len(loaded.Items))
	}

	// Проверяем логин
	if loaded.Items[0].Type != "login" {
		t.Errorf("Items[0].Type = %v, хотим 'login'", loaded.Items[0].Type)
	}
	if loaded.Items[0].Username != "user@gmail.com" {
		t.Errorf("Items[0].Username = %v, хотим 'user@gmail.com'", loaded.Items[0].Username)
	}

	// Проверяем текст
	if loaded.Items[1].Type != "text" {
		t.Errorf("Items[1].Type = %v, хотим 'text'", loaded.Items[1].Type)
	}
	if loaded.Items[1].Content != "Важные заметки" {
		t.Errorf("Items[1].Content = %v, хотим 'Важные заметки'", loaded.Items[1].Content)
	}

	// Проверяем карту
	if loaded.Items[2].Type != "card" {
		t.Errorf("Items[2].Type = %v, хотим 'card'", loaded.Items[2].Type)
	}
	if loaded.Items[2].Number != "4111111111111111" {
		t.Errorf("Items[2].Number = %v, хотим '4111111111111111'", loaded.Items[2].Number)
	}
}

func TestSecretStorage_Save_UpdatesLastModified(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.enc")

	ss := &SecretStorage{dataFile: testFile}
	key := crypto.GenerateEncryptionKey("testPassword123")

	// Создаем хранилище
	storage := NewStorage()
	item := NewItem("login", "Тест")
	storage.Items = append(storage.Items, item)

	// Сохраняем
	err := ss.Save(storage, key)
	if err != nil {
		t.Fatalf("Save() ошибка = %v", err)
	}

	// Загружаем
	loaded, err := ss.Load(key)
	if err != nil {
		t.Fatalf("Load() ошибка = %v", err)
	}

	// Проверяем, что LastModified установлен
	if loaded.LastModified == "" {
		t.Error("LastModified должен быть установлен")
	}

	// Проверяем, что время валидное
	_, err = time.Parse(time.RFC3339, loaded.LastModified)
	if err != nil {
		t.Errorf("LastModified имеет неверный формат: %v", err)
	}
}

func TestSecretStorage_Delete(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.enc")

	ss := &SecretStorage{dataFile: testFile}
	key := crypto.GenerateEncryptionKey("testPassword123")

	// Создаем и сохраняем хранилище
	storage := NewStorage()
	item := NewItem("text", "Тест")
	storage.Items = append(storage.Items, item)

	err := ss.Save(storage, key)
	if err != nil {
		t.Fatalf("Save() ошибка = %v", err)
	}

	// Проверяем, что файл существует
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("файл не был создан")
	}

	// Удаляем
	err = ss.Delete()
	if err != nil {
		t.Fatalf("Delete() ошибка = %v", err)
	}

	// Проверяем, что файл удален
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("Delete() не удалил файл")
	}
}

func TestSecretStorage_Delete_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "nonexistent.enc")

	ss := &SecretStorage{dataFile: testFile}

	// Удаляем несуществующий файл - не должно быть ошибки
	err := ss.Delete()
	if err != nil {
		t.Errorf("Delete() несуществующего файла вернул ошибку: %v", err)
	}
}

func TestSecretStorage_SaveAndLoad_Metadata(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.enc")

	ss := &SecretStorage{dataFile: testFile}
	key := crypto.GenerateEncryptionKey("testPassword123")

	// Создаем элемент с метаданными
	storage := NewStorage()
	item := NewItem("login", "Тест с метаданными")
	item.Username = "user"
	item.Password = "pass"
	item.Metadata = map[string]string{
		"url":      "https://example.com",
		"category": "social",
		"tags":     "important,work",
	}
	storage.Items = append(storage.Items, item)

	// Сохраняем
	err := ss.Save(storage, key)
	if err != nil {
		t.Fatalf("Save() ошибка = %v", err)
	}

	// Загружаем
	loaded, err := ss.Load(key)
	if err != nil {
		t.Fatalf("Load() ошибка = %v", err)
	}

	// Проверяем метаданные
	if len(loaded.Items[0].Metadata) != 3 {
		t.Errorf("Metadata содержит %d элементов, хотим 3", len(loaded.Items[0].Metadata))
	}

	if loaded.Items[0].Metadata["url"] != "https://example.com" {
		t.Errorf("Metadata['url'] = %v, хотим 'https://example.com'", loaded.Items[0].Metadata["url"])
	}

	if loaded.Items[0].Metadata["category"] != "social" {
		t.Errorf("Metadata['category'] = %v, хотим 'social'", loaded.Items[0].Metadata["category"])
	}

	if loaded.Items[0].Metadata["tags"] != "important,work" {
		t.Errorf("Metadata['tags'] = %v, хотим 'important,work'", loaded.Items[0].Metadata["tags"])
	}
}

func TestSecretStorage_SaveAndLoad_EmptyStorage(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.enc")

	ss := &SecretStorage{dataFile: testFile}
	key := crypto.GenerateEncryptionKey("testPassword123")

	// Создаем пустое хранилище
	storage := NewStorage()

	// Сохраняем
	err := ss.Save(storage, key)
	if err != nil {
		t.Fatalf("Save() ошибка = %v", err)
	}

	// Загружаем
	loaded, err := ss.Load(key)
	if err != nil {
		t.Fatalf("Load() ошибка = %v", err)
	}

	// Проверяем, что хранилище пустое
	if len(loaded.Items) != 0 {
		t.Errorf("Load() вернул %d элементов, хотим 0", len(loaded.Items))
	}

	// Проверяем, что LastModified установлен
	if loaded.LastModified == "" {
		t.Error("LastModified не должен быть пустым")
	}
}

func TestSecretStorage_Load_CorruptedFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.enc")

	ss := &SecretStorage{dataFile: testFile}
	key := crypto.GenerateEncryptionKey("testPassword123")

	// Создаем поврежденный файл
	err := os.WriteFile(testFile, []byte("corrupted data"), 0600)
	if err != nil {
		t.Fatalf("не удалось создать поврежденный файл: %v", err)
	}

	// Пытаемся загрузить
	_, err = ss.Load(key)
	if err == nil {
		t.Error("Load() поврежденного файла должен возвращать ошибку")
	}
}