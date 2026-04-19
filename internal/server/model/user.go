package model

// User представляет сущность пользователя.
// Login используется как уникальный идентификатор (Primary Key).
type User struct {
	Login         string
	PasswordHash  string // Хеш пароля (bcrypt)
	EncryptedData []byte // Зашифрованные данные пользователя
	UpdatedAt     int64  // Время последнего обновления данных (Unix timestamp)
}
