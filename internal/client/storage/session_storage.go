package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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
