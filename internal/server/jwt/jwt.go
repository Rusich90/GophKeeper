package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager определяет интерфейс для управления JWT токенами
type TokenManager interface {
	GenerateToken(login string) (string, error)
	ValidateToken(tokenString string) (string, error)
}

type Manager struct {
	secret string
	ttl    time.Duration
}

func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{
		secret: secret,
		ttl:    ttl,
	}
}

func (m *Manager) GenerateToken(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(m.ttl).Unix(),
	})

	tokenString, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (m *Manager) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secret), nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	login, ok := claims["login"].(string)
	if !ok {
		return "", fmt.Errorf("login not found in token")
	}

	return login, nil
}
