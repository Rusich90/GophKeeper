package service

import (
	"context"
	"errors"
	"fmt"

	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Rusich90/GophKeeper/internal/server/model"
	"github.com/Rusich90/GophKeeper/internal/server/storage/user"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid login or password")
)

type AuthService struct {
	userRepo user.UserRepository
	log      *slog.Logger
}

func NewAuthService(userRepo user.UserRepository, log *slog.Logger) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		log:      log,
	}
}

func (s *AuthService) Register(ctx context.Context, login, password string) error {
	s.log.Debug("Register attempt", "login", login)

	if err := validateCredentials(login, password); err != nil {
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

	if err := validateCredentials(login, password); err != nil {
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

	token := uuid.New().String()
	s.log.Info("User logged in successfully", "login", login)
	return token, nil
}

func validateCredentials(login, password string) error {
	if login == "" {
		return errors.New("login cannot be empty")
	}
	if len(login) < 3 {
		return errors.New("login must be at least 3 characters")
	}
	if len(login) > 50 {
		return errors.New("login must be at most 50 characters")
	}
	if password == "" {
		return errors.New("password cannot be empty")
	}
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if len(password) > 100 {
		return errors.New("password must be at most 100 characters")
	}
	return nil
}
