package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateEncryptionKey(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantLen  int
	}{
		{
			name:     "обычный пароль",
			password: "mySecurePassword123",
			wantLen:  44, // base64 encoded 32 bytes
		},
		{
			name:     "короткий пароль",
			password: "abc",
			wantLen:  44,
		},
		{
			name:     "пустой пароль",
			password: "",
			wantLen:  44,
		},
		{
			name:     "пароль с спецсимволами",
			password: "p@$$w0rd!#$%",
			wantLen:  44,
		},
		{
			name:     "длинный пароль",
			password: strings.Repeat("a", 1000),
			wantLen:  44,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateEncryptionKey(tt.password)
			if len(got) != tt.wantLen {
				t.Errorf("GenerateEncryptionKey() длина = %v, хотим %v", len(got), tt.wantLen)
			}

			// Проверяем, что результат валидный base64
			_, err := base64.StdEncoding.DecodeString(got)
			if err != nil {
				t.Errorf("GenerateEncryptionKey() вернул невалидный base64: %v", err)
			}
		})
	}
}

func TestGenerateEncryptionKey_Deterministic(t *testing.T) {
	password := "testPassword123"
	key1 := GenerateEncryptionKey(password)
	key2 := GenerateEncryptionKey(password)

	if key1 != key2 {
		t.Errorf("GenerateEncryptionKey() должен быть детерминированным для одного пароля")
	}
}

func TestEncrypt(t *testing.T) {
	tests := []struct {
		name    string
		plain   []byte
		key     string
		wantErr bool
	}{
		{
			name:    "успешное шифрование",
			plain:   []byte("секретные данные"),
			key:     GenerateEncryptionKey("password123"),
			wantErr: false,
		},
		{
			name:    "пустые данные",
			plain:   []byte(""),
			key:     GenerateEncryptionKey("password123"),
			wantErr: false,
		},
		{
			name:    "большие данные",
			plain:   []byte(strings.Repeat("a", 10000)),
			key:     GenerateEncryptionKey("password123"),
			wantErr: false,
		},
		{
			name:    "невалидный ключ (не base64)",
			plain:   []byte("данные"),
			key:     "not-valid-base64!!!",
			wantErr: true,
		},
		{
			name:    "ключ неправильной длины",
			plain:   []byte("данные"),
			key:     base64.StdEncoding.EncodeToString([]byte("short")),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encrypt(tt.plain, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() ошибка = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) == 0 {
					t.Errorf("Encrypt() вернул пустой результат")
				}

				// Зашифрованные данные должны отличаться от исходных
				if string(got) == string(tt.plain) {
					t.Errorf("Encrypt() данные не зашифрованы")
				}
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	key := GenerateEncryptionKey("password123")
	plaintext := []byte("секретные данные")

	// Сначала шифруем данные
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("не удалось зашифровать данные для теста: %v", err)
	}

	tests := []struct {
		name    string
		cipher  []byte
		key     string
		want    []byte
		wantErr bool
	}{
		{
			name:    "успешная расшифровка",
			cipher:  ciphertext,
			key:     key,
			want:    plaintext,
			wantErr: false,
		},
		{
			name:    "невалидный ключ",
			cipher:  ciphertext,
			key:     GenerateEncryptionKey("wrongpassword"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "невалидный ключ (не base64)",
			cipher:  ciphertext,
			key:     "not-valid-base64!!!",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "пустые зашифрованные данные",
			cipher:  []byte(""),
			key:     key,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "слишком короткие зашифрованные данные",
			cipher:  []byte("short"),
			key:     key,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "поврежденные зашифрованные данные",
			cipher:  []byte("corrupted data"),
			key:     key,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decrypt(tt.cipher, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() ошибка = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if string(got) != string(tt.want) {
					t.Errorf("Decrypt() = %v, хотим %v", string(got), string(tt.want))
				}
			}
		})
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "текст на русском",
			data: []byte("Привет, мир! Это секретные данные."),
		},
		{
			name: "текст на английском",
			data: []byte("Hello, World! This is secret data."),
		},
		{
			name: "спецсимволы",
			data: []byte("!@#$%^&*()_+-=[]{}|;':\",./<>?"),
		},
		{
			name: "числа",
			data: []byte("1234567890"),
		},
		{
			name: "пустые данные",
			data: []byte(""),
		},
		{
			name: "большие данные",
			data: []byte(strings.Repeat("test data ", 1000)),
		},
		{
			name: "двоичные данные",
			data: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GenerateEncryptionKey("testPassword123")

			// Шифруем
			encrypted, err := Encrypt(tt.data, key)
			if err != nil {
				t.Fatalf("Encrypt() ошибка = %v", err)
			}

			// Расшифровываем
			decrypted, err := Decrypt(encrypted, key)
			if err != nil {
				t.Fatalf("Decrypt() ошибка = %v", err)
			}

			// Проверяем, что данные совпадают
			if string(decrypted) != string(tt.data) {
				t.Errorf("данные не совпадают после шифрования/расшифровки: got %v, want %v", string(decrypted), string(tt.data))
			}
		})
	}
}

func TestEncrypt_DifferentNonces(t *testing.T) {
	key := GenerateEncryptionKey("password123")
	plaintext := []byte("одинаковые данные")

	// Шифруем одни и те же данные дважды
	ciphertext1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("первое шифрование не удалось: %v", err)
	}

	ciphertext2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("второе шифрование не удалось: %v", err)
	}

	// Зашифрованные данные должны отличаться из-за разных nonce
	if string(ciphertext1) == string(ciphertext2) {
		t.Errorf("зашифрованные данные должны отличаться из-за разных nonce")
	}

	// Но оба должны расшифровываться в одинаковые данные
	decrypted1, err := Decrypt(ciphertext1, key)
	if err != nil {
		t.Fatalf("первая расшифровка не удалась: %v", err)
	}

	decrypted2, err := Decrypt(ciphertext2, key)
	if err != nil {
		t.Fatalf("вторая расшифровка не удалась: %v", err)
	}

	if string(decrypted1) != string(decrypted2) {
		t.Errorf("расшифрованные данные должны совпадать")
	}
}