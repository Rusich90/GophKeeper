package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// Encrypt шифрует данные с использованием AES-GCM
func Encrypt(plaintext []byte, key string) ([]byte, error) {
	// Декодируем ключ из base64
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования ключа: %w", err)
	}

	// Создаем блок шифрования
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания блока шифрования: %w", err)
	}

	// Создаем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %w", err)
	}

	// Генерируем nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("ошибка генерации nonce: %w", err)
	}

	// Шифруем данные (nonce добавляется в начало зашифрованных данных)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// Decrypt расшифровывает данные с использованием AES-GCM
func Decrypt(ciphertext []byte, key string) ([]byte, error) {
	// Декодируем ключ из base64
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования ключа: %w", err)
	}

	// Создаем блок шифрования
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания блока шифрования: %w", err)
	}

	// Создаем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %w", err)
	}

	// Проверяем размер зашифрованных данных
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("зашифрованные данные слишком короткие")
	}

	// Извлекаем nonce и зашифрованные данные
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Расшифровываем данные
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки: %w", err)
	}

	return plaintext, nil
}

// GenerateEncryptionKey генерирует ключ шифрования из пароля
func GenerateEncryptionKey(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}