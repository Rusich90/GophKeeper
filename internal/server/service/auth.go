package service

import (
	"context"
	"errors"
	"fmt"

	"log/slog"

	"golang.org/x/crypto/bcrypt"

	"github.com/Rusich90/GophKeeper/internal/server/jwt"
	"github.com/Rusich90/GophKeeper/internal/server/model"
	"github.com/Rusich90/GophKeeper/internal/server/storage/token"
	"github.com/Rusich90/GophKeeper/internal/server/storage/user"
	"github.com/Rusich90/GophKeeper/pkg/validator"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid login or password")
)

type AuthService struct {
	userRepo  user.UserRepository
	tokenRepo token.TokenRepository
	jwtMgr    jwt.TokenManager
	log       *slog.Logger
	validator validator.CredentialsValidator
}

func NewAuthService(userRepo user.UserRepository, tokenRepo token.TokenRepository, jwtMgr jwt.TokenManager, log *slog.Logger) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtMgr:    jwtMgr,
		log:       log,
		validator: validator.NewCredentialsValidator(),
	}
}

func (s *AuthService) Register(ctx context.Context, login, password string) error {
	s.log.Debug("Register attempt", "login", login)

	if err := s.validator.ValidateCredentials(login, password); err != nil {
		s.log.Warn("Invalid credentials", "login", login, "error", err)
		return err
	}

	existingUser, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			s.log.Debug("User not found, proceeding with registration", "login", login)
		} else {
			s.log.Error("Failed to check user existence", "login", login, "error", err)
			return fmt.Errorf("failed to check user existence: %w", err)
		}
	}
	if existingUser != nil {
		s.log.Warn("User already exists", "login", login)
		return ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("Failed to hash password", "login", login, "error", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &model.User{
		Login:        login,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		s.log.Error("Failed to create user", "login", login, "error", err)
		return err
	}

	s.log.Info("User registered successfully", "login", login)
	return nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	s.log.Debug("Login attempt", "login", login)

	if err := s.validator.ValidateCredentials(login, password); err != nil {
		s.log.Warn("Invalid credentials", "login", login, "error", err)
		return "", err
	}

	u, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			s.log.Warn("User not found", "login", login)
			return "", ErrInvalidCredentials
		}
		s.log.Error("Failed to find user", "login", login, "error", err)
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		s.log.Warn("Invalid password", "login", login)
		return "", ErrInvalidCredentials
	}

	// Создаем JWT токен
	tokenString, err := s.jwtMgr.GenerateToken(u.Login)
	if err != nil {
		s.log.Error("Failed to generate token", "login", login, "error", err)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Сохраняем токен в Redis с ключом = login
	if err := s.tokenRepo.StoreToken(ctx, u.Login, tokenString); err != nil {
		s.log.Error("Failed to store token", "login", login, "error", err)
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	s.log.Info("User logged in successfully", "login", login)
	return tokenString, nil
}

func (s *AuthService) Logout(ctx context.Context, login string) error {
	s.log.Debug("Logout attempt", "login", login)

	if err := s.tokenRepo.DeleteToken(ctx, login); err != nil {
		s.log.Error("Failed to delete token", "error", err)
		return fmt.Errorf("failed to delete token: %w", err)
	}

	s.log.Info("User logged out successfully", "login", login)
	return nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	s.log.Debug("Validating token")

	// Проверяем подпись JWT токена и получаем login
	login, err := s.jwtMgr.ValidateToken(tokenString)
	if err != nil {
		s.log.Warn("Failed to validate JWT token", "error", err)
		return "", fmt.Errorf("invalid token: %w", err)
	}

	// Получаем актуальный токен из Redis по login
	storedToken, err := s.tokenRepo.GetToken(ctx, login)
	if err != nil {
		s.log.Warn("Token not found in storage", "login", login, "error", err)
		return "", fmt.Errorf("token not found or expired")
	}

	// Сравниваем токен из запроса с актуальным токеном
	if storedToken != tokenString {
		s.log.Warn("Token mismatch", "login", login)
		return "", fmt.Errorf("invalid token")
	}

	s.log.Debug("Token validated successfully", "login", login)
	return login, nil
}
