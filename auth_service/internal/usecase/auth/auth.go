package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kprf42/dolgova/auth_service/internal/entity"
	"github.com/kprf42/dolgova/auth_service/internal/repository"
	"github.com/kprf42/dolgova/auth_service/internal/usecase/jwt"
	"github.com/kprf42/dolgova/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	repo repository.UserRepository
	jwt  *jwt.JWTService
	log  *logger.Logger
}

func NewAuthUseCase(repo repository.UserRepository, jwtSecret string, accessExpiry, refreshExpiry time.Duration, log *logger.Logger) *AuthUseCase {
	return &AuthUseCase{
		repo: repo,
		jwt:  jwt.NewJWTService(jwtSecret, accessExpiry, refreshExpiry),
		log:  log,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, username, email, password string) (*entity.User, error) {
	uc.log.Info("Starting user registration",
		logger.String("username", username),
		logger.String("email", email))

	// Валидация и нормализация ввода
	username = strings.TrimSpace(username)
	if username == "" {
		uc.log.Warn("Empty username provided")
		return nil, entity.ErrEmptyUsername
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if !isValidEmail(email) {
		uc.log.Warn("Invalid email provided",
			logger.String("email", email))
		return nil, entity.ErrInvalidEmail
	}

	if len(password) < 8 {
		uc.log.Warn("Weak password provided")
		return nil, entity.ErrWeakPassword
	}

	// Проверка существования пользователя
	existingUser, err := uc.repo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		uc.log.Error("Failed to check user existence",
			logger.String("email", email),
			logger.Error(err))
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if existingUser != nil {
		uc.log.Warn("User already exists",
			logger.String("email", email))
		return nil, entity.ErrUserAlreadyExists
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		uc.log.Error("Failed to hash password",
			logger.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создание пользователя
	user := &entity.User{
		ID:       uuid.New().String(),
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	uc.log.Debug("Created user object",
		logger.String("user_id", user.ID),
		logger.String("username", user.Username),
		logger.String("email", user.Email),
		logger.String("role", user.Role))

	if err := uc.repo.CreateUser(ctx, user); err != nil {
		uc.log.Error("Failed to create user",
			logger.String("user_id", user.ID),
			logger.Error(err))
		return nil, err
	}

	uc.log.Info("Successfully registered user",
		logger.String("user_id", user.ID),
		logger.String("username", user.Username),
		logger.String("email", user.Email))

	return user, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*entity.TokenDetails, error) {
	uc.log.Info("Attempting user login",
		logger.String("email", email))

	user, err := uc.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			uc.log.Warn("User not found during login",
				logger.String("email", email))
			return nil, fmt.Errorf("invalid credentials")
		}
		uc.log.Error("Failed to get user during login",
			logger.String("email", email),
			logger.Error(err))
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		uc.log.Warn("Invalid password during login",
			logger.String("user_id", user.ID))
		return nil, fmt.Errorf("invalid credentials")
	}

	tokens, err := uc.jwt.GenerateTokens(user.ID)
	if err != nil {
		uc.log.Error("Failed to generate tokens",
			logger.String("user_id", user.ID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	uc.log.Info("Successfully logged in user",
		logger.String("user_id", user.ID))

	return tokens, nil
}

func isValidEmail(email string) bool {
	// Простая проверка на наличие @ и домена
	return strings.Contains(email, "@") && strings.Contains(email[strings.Index(email, "@"):], ".")
}
