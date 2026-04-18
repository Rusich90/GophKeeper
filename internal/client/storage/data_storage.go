package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Rusich90/GophKeeper/internal/client/crypto"
)

// DataStorage управляет сохранением и загрузкой зашифрованных данных
type DataStorage struct {
	dataFile string
}

// NewDataStorage создает новое хранилище данных
func NewDataStorage() *DataStorage {
	configDir, err := getConfigDir()
	if err != nil {
		// Если не можем создать директорию, используем текущую
		configDir = ".gophkeeper"
		if err := os.MkdirAll(configDir, 0700); err != nil {
			// Если и это не удалось, используем временный файл
			return &DataStorage{dataFile: ".data.enc"}
		}
	}

	return &DataStorage{
		dataFile: filepath.Join(configDir, "data.enc"),
	}
}

// Load загружает и расшифровывает данные из файла
func (ds *DataStorage) Load(key string) (*Storage, error) {
	// Проверяем существование файла
	if _, err := os.Stat(ds.dataFile); os.IsNotExist(err) {
		// Файл не существует, возвращаем пустое хранилище
		return NewStorage(), nil
	}

	// Читаем зашифрованные данные
	ciphertext, err := os.ReadFile(ds.dataFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла данных: %w", err)
	}

	// Расшифровываем данные
	plaintext, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки данных: %w", err)
	}

	// Десериализуем JSON
	var storage Storage
	if err := json.Unmarshal(plaintext, &storage); err != nil {
		return nil, fmt.Errorf("ошибка десериализации данных: %w", err)
	}

	return &storage, nil
}

// Save шифрует и сохраняет данные в файл
func (ds *DataStorage) Save(storage *Storage, key string) error {
	// Обновляем время последнего изменения
	storage.LastModified = time.Now().UTC().Format(time.RFC3339)

	// Сериализуем в JSON
	plaintext, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	// Шифруем данные
	ciphertext, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		return fmt.Errorf("ошибка шифрования данных: %w", err)
	}

	// Записываем в файл
	if err := os.WriteFile(ds.dataFile, ciphertext, 0600); err != nil {
		return fmt.Errorf("ошибка записи файла данных: %w", err)
	}

	return nil
}

// Delete удаляет файл данных
func (ds *DataStorage) Delete() error {
	if err := os.Remove(ds.dataFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ошибка удаления файла данных: %w", err)
	}
	return nil
}