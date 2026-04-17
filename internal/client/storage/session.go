package storage

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Session представляет сессию пользователя с токеном и ключом шифрования
type Session struct {
	Token string `json:"token"`
	Key   string `json:"key"`
}

// SessionStorage управляет сохранением и загрузкой сессии пользователя
type SessionStorage struct {
	sessionFile string
}

// NewSessionStorage создает новое хранилище сессий
func NewSessionStorage() *SessionStorage {
	configDir, err := getConfigDir()
	if err != nil {
		// Если не можем создать директорию, используем текущую
		configDir = ".gophkeeper"
		if err := os.MkdirAll(configDir, 0700); err != nil {
			// Если и это не удалось, используем временный файл
			return &SessionStorage{sessionFile: ".session.json"}
		}
	}

	return &SessionStorage{
		sessionFile: filepath.Join(configDir, "session.json"),
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

// SaveSession сохраняет сессию в файл
func (ss *SessionStorage) SaveSession(session *Session) error {
	jsonData, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(ss.sessionFile, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession загружает сессию из файла
func (ss *SessionStorage) LoadSession() (*Session, error) {
	data, err := os.ReadFile(ss.sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session file not found")
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// DeleteSession удаляет файл с сессией
func (ss *SessionStorage) DeleteSession() error {
	if err := os.Remove(ss.sessionFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}
	return nil
}

// GenerateEncryptionKey генерирует ключ шифрования из пароля
func GenerateEncryptionKey(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}