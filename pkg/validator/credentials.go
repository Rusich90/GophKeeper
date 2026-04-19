package validator

import (
	"fmt"
	"unicode"
)

const (
	MinLoginLength    = 3
	MaxLoginLength    = 50
	MinPasswordLength = 6
	MaxPasswordLength = 100
)

// CredentialsValidator определяет интерфейс для валидации учетных данных
type CredentialsValidator interface {
	ValidateLogin(login string) error
	ValidatePassword(password string) error
	ValidateCredentials(login, password string) error
}

// credentialsValidator реализует CredentialsValidator
type credentialsValidator struct{}

// NewCredentialsValidator создает новый валидатор учетных данных
func NewCredentialsValidator() CredentialsValidator {
	return &credentialsValidator{}
}

// ValidateLogin проверяет валидность логина
func (v *credentialsValidator) ValidateLogin(login string) error {
	if login == "" {
		return fmt.Errorf("логин не может быть пустым")
	}

	if len(login) < MinLoginLength {
		return fmt.Errorf("логин должен содержать минимум %d символа", MinLoginLength)
	}

	if len(login) > MaxLoginLength {
		return fmt.Errorf("логин не должен превышать %d символов", MaxLoginLength)
	}

	// Проверяем, что логин содержит только допустимые символы
	for _, r := range login {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return fmt.Errorf("логин может содержать только буквы, цифры, символы '_' и '-'")
		}
	}

	return nil
}

// ValidatePassword проверяет валидность пароля
func (v *credentialsValidator) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("пароль не может быть пустым")
	}

	if len(password) < MinPasswordLength {
		return fmt.Errorf("пароль должен содержать минимум %d символов", MinPasswordLength)
	}

	if len(password) > MaxPasswordLength {
		return fmt.Errorf("пароль не должен превышать %d символов", MaxPasswordLength)
	}

	// Проверяем сложность пароля
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	// Требуем все 4 критерия сложности
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return fmt.Errorf("пароль должен содержать: заглавные буквы, строчные буквы, цифры и специальные символы")
	}

	return nil
}

// ValidateCredentials проверяет валидность логина и пароля
func (v *credentialsValidator) ValidateCredentials(login, password string) error {
	if err := v.ValidateLogin(login); err != nil {
		return err
	}
	if err := v.ValidatePassword(password); err != nil {
		return err
	}
	return nil
}
