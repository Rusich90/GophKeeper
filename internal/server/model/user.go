package model

// User представляет сущность пользователя.
// Login используется как уникальный идентификатор (Primary Key).
type User struct {
	Login        string
	PasswordHash string // Хеш пароля (bcrypt)
}
