package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	secret := "test-secret"
	ttl := time.Hour

	mgr := NewManager(secret, ttl)

	assert.NotNil(t, mgr)
	assert.Equal(t, secret, mgr.secret)
	assert.Equal(t, ttl, mgr.ttl)
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		ttl     time.Duration
		login   string
		wantErr bool
	}{
		{
			name:    "successful token generation",
			secret:  "test-secret",
			ttl:     time.Hour,
			login:   "testuser",
			wantErr: false,
		},
		{
			name:    "token with short ttl",
			secret:  "test-secret",
			ttl:     time.Minute,
			login:   "testuser",
			wantErr: false,
		},
		{
			name:    "token with long ttl",
			secret:  "test-secret",
			ttl:     24 * time.Hour,
			login:   "testuser",
			wantErr: false,
		},
		{
			name:    "token with empty login",
			secret:  "test-secret",
			ttl:     time.Hour,
			login:   "",
			wantErr: false,
		},
		{
			name:    "token with special characters in login",
			secret:  "test-secret",
			ttl:     time.Hour,
			login:   "user@example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.secret, tt.ttl)

			token, err := mgr.GenerateToken(tt.login)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name        string
		secret      string
		ttl         time.Duration
		login       string
		tokenString string
		wantLogin   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid token",
			secret:    "test-secret",
			ttl:       time.Hour,
			login:     "testuser",
			wantLogin: "testuser",
			wantErr:   false,
		},
		{
			name:      "valid token with different login",
			secret:    "test-secret",
			ttl:       time.Hour,
			login:     "anotheruser",
			wantLogin: "anotheruser",
			wantErr:   false,
		},
		{
			name:        "invalid token format",
			secret:      "test-secret",
			ttl:         time.Hour,
			login:       "testuser",
			tokenString: "invalid.token.string",
			wantLogin:   "",
			wantErr:     true,
			errContains: "invalid token",
		},
		{
			name:        "empty token",
			secret:      "test-secret",
			ttl:         time.Hour,
			login:       "testuser",
			tokenString: "",
			wantLogin:   "",
			wantErr:     true,
			errContains: "invalid token",
		},
		{
			name:        "token with wrong secret",
			secret:      "test-secret",
			ttl:         time.Hour,
			login:       "testuser",
			tokenString: generateTokenWithSecret("wrong-secret", time.Hour, "testuser"),
			wantLogin:   "",
			wantErr:     true,
			errContains: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.secret, tt.ttl)

			token := tt.tokenString
			// Генерируем токен только если он не задан явно
			if token == "" && !tt.wantErr {
				var err error
				token, err = mgr.GenerateToken(tt.login)
				require.NoError(t, err)
			}

			login, err := mgr.ValidateToken(token)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Empty(t, login)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLogin, login)
			}
		})
	}
}

func TestValidateToken_Expired(t *testing.T) {
	secret := "test-secret"
	mgr := NewManager(secret, time.Nanosecond) // Очень короткий TTL

	token, err := mgr.GenerateToken("testuser")
	require.NoError(t, err)

	// Ждем истечения токена
	time.Sleep(10 * time.Millisecond)

	login, err := mgr.ValidateToken(token)
	assert.Error(t, err)
	assert.Empty(t, login)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret"
	ttl := time.Hour
	login := "testuser"

	mgr := NewManager(secret, ttl)

	// Генерируем токен
	token, err := mgr.GenerateToken(login)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Валидируем токен
	validatedLogin, err := mgr.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, login, validatedLogin)
}

func TestTokenStructure(t *testing.T) {
	secret := "test-secret"
	ttl := time.Hour
	login := "testuser"

	mgr := NewManager(secret, ttl)

	token, err := mgr.GenerateToken(login)
	require.NoError(t, err)

	// Проверяем, что токен имеет три части (header.payload.signature)
	parts := len(splitToken(token))
	assert.Equal(t, 3, parts, "Token should have 3 parts separated by dots")
}

func TestMultipleTokens(t *testing.T) {
	secret := "test-secret"
	ttl := time.Hour
	mgr := NewManager(secret, ttl)

	logins := []string{"user1", "user2", "user3"}
	tokens := make(map[string]string)

	// Генерируем токены для разных пользователей
	for _, login := range logins {
		token, err := mgr.GenerateToken(login)
		require.NoError(t, err)
		tokens[login] = token
	}

	// Валидируем все токены
	for _, login := range logins {
		validatedLogin, err := mgr.ValidateToken(tokens[login])
		require.NoError(t, err)
		assert.Equal(t, login, validatedLogin)
	}
}

// Вспомогательная функция для генерации токена с другим секретом
func generateTokenWithSecret(secret string, ttl time.Duration, login string) string {
	mgr := NewManager(secret, ttl)
	token, err := mgr.GenerateToken(login)
	if err != nil {
		return ""
	}
	return token
}

// Вспомогательная функция для разделения токена
func splitToken(token string) []string {
	parts := make([]string, 0)
	current := ""
	for _, char := range token {
		if char == '.' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
