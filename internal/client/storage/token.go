package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TokenStorage управляет сохранением и загрузкой токена авторизации
type TokenStorage struct {
	tokenFile string
}

// NewTokenStorage создает новое хранилище токенов
func NewTokenStorage() *TokenStorage {
	configDir, err := getConfigDir()
	if err != nil {
		// Если не можем создать директорию, используем текущую
		configDir = ".gophkeeper"
		if err := os.MkdirAll(configDir, 0700); err != nil {
			// Если и это не удалось, используем временный файл
			return &TokenStorage{tokenFile: ".token.json"}
		}
	}

	return &TokenStorage{
		tokenFile: filepath.Join(configDir, "token.json"),
	}
}

// getConfigDir возвращает директорию для хранения конфигурации
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".gophkeeper")

	// Создаем директорию с правами только для владельца
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

// SaveToken сохраняет токен в файл
func (ts *TokenStorage) SaveToken(token string) error {
	data := map[string]string{
		"token": token,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(ts.tokenFile, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// LoadToken загружает токен из файла
func (ts *TokenStorage) LoadToken() (string, error) {
	data, err := os.ReadFile(ts.tokenFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("token file not found")
		}
		return "", fmt.Errorf("failed to read token file: %w", err)
	}

	var tokenData map[string]string
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return "", fmt.Errorf("failed to unmarshal token: %w", err)
	}

	token, ok := tokenData["token"]
	if !ok {
		return "", fmt.Errorf("token not found in file")
	}

	return token, nil
}

// DeleteToken удаляет файл с токеном
func (ts *TokenStorage) DeleteToken() error {
	if err := os.Remove(ts.tokenFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}
	return nil
}
