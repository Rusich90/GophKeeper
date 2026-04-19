package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

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
