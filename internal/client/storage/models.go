package storage

import (
	"time"

	"github.com/google/uuid"
)

// Storage представляет корневую структуру данных
type Storage struct {
	LastModified string `json:"last_modified"` // Метка времени последнего изменения всего файла
	Items        []Item `json:"items"`
}

// Item представляет отдельную запись
type Item struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"` // "login", "text", "card"
	Title     string            `json:"title"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"` // Метка времени для синхронизации конкретной записи
	Metadata  map[string]string `json:"metadata"`
	// Поля для типа "login"
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	// Поля для типа "text"
	Content string `json:"content,omitempty"`
	// Поля для типа "card"
	Number      string `json:"number,omitempty"`
	ExpiryMonth string `json:"expiry_month,omitempty"`
	ExpiryYear  string `json:"expiry_year,omitempty"`
	CVV         string `json:"cvv,omitempty"`
}

// Session представляет сессию пользователя с токеном и ключом шифрования
type Session struct {
	Token      string `json:"token"`
	Key        string `json:"key"`
	LastSyncTS int64  `json:"last_sync_ts"` // Время последней синхронизации (Unix timestamp)
}

// NewItem создает новую запись с текущим временем (UTC)
func NewItem(itemType, title string) Item {
	now := time.Now().UTC().Format(time.RFC3339)
	return Item{
		ID:        uuid.New().String(),
		Type:      itemType,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]string),
	}
}

// NewStorage создает новое пустое хранилище
func NewStorage() *Storage {
	return &Storage{
		LastModified: time.Now().UTC().Format(time.RFC3339),
		Items:        []Item{},
	}
}
