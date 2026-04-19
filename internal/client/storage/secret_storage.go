package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Rusich90/GophKeeper/internal/client/crypto"
)

// SecretStorage управляет сохранением и загрузкой зашифрованных секретов
type SecretStorage struct {
	dataFile string
}

// NewSecretStorage создает новое хранилище секретов
func NewSecretStorage() *SecretStorage {
	configDir, err := getConfigDir()
	if err != nil {
		// Если не можем создать директорию, используем текущую
		configDir = ".gophkeeper"
		if err := os.MkdirAll(configDir, 0700); err != nil {
			// Если и это не удалось, используем временный файл
			return &SecretStorage{dataFile: ".data.enc"}
		}
	}

	return &SecretStorage{
		dataFile: filepath.Join(configDir, "data.enc"),
	}
}

// Load загружает и расшифровывает секреты из файла
func (ss *SecretStorage) Load(key string) (*Storage, error) {
	// Проверяем существование файла
	if _, err := os.Stat(ss.dataFile); os.IsNotExist(err) {
		// Файл не существует, возвращаем пустое хранилище
		return NewStorage(), nil
	}

	// Читаем зашифрованные секреты
	ciphertext, err := os.ReadFile(ss.dataFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла секретов: %w", err)
	}

	// Расшифровываем секреты
	plaintext, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки секретов: %w", err)
	}

	// Десериализуем JSON
	var storage Storage
	if err := json.Unmarshal(plaintext, &storage); err != nil {
		return nil, fmt.Errorf("ошибка десериализации секретов: %w", err)
	}

	return &storage, nil
}

// Save шифрует и сохраняет секреты в файл
func (ss *SecretStorage) Save(storage *Storage, key string) error {
	// Обновляем время последнего изменения
	storage.LastModified = time.Now().UTC().Format(time.RFC3339)

	// Сериализуем в JSON
	plaintext, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации секретов: %w", err)
	}

	// Шифруем секреты
	ciphertext, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		return fmt.Errorf("ошибка шифрования секретов: %w", err)
	}

	// Записываем в файл
	if err := os.WriteFile(ss.dataFile, ciphertext, 0600); err != nil {
		return fmt.Errorf("ошибка записи файла секретов: %w", err)
	}

	return nil
}

// Delete удаляет файл секретов
func (ss *SecretStorage) Delete() error {
	if err := os.Remove(ss.dataFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ошибка удаления файла секретов: %w", err)
	}
	return nil
}
