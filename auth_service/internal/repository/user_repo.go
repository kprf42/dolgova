package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/kprf42/dolgova/auth_service/internal/entity"
	"github.com/kprf42/dolgova/pkg/logger"
)

type UserRepository struct {
	db  *sql.DB
	log *logger.Logger
}

func NewUserRepository(db *sql.DB, log *logger.Logger) *UserRepository {
	return &UserRepository{
		db:  db,
		log: log,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	r.log.Info("Creating new user",
		logger.String("user_id", user.ID),
		logger.String("username", user.Username),
		logger.String("email", user.Email),
		logger.String("role", user.Role))

	query := `
		INSERT INTO users (id, username, email, password, role)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			r.log.Warn("Email already exists",
				logger.String("email", user.Email))
			return fmt.Errorf("email already exists")
		}
		r.log.Error("Failed to create user",
			logger.String("user_id", user.ID),
			logger.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected",
			logger.String("user_id", user.ID),
			logger.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		r.log.Error("No rows affected when creating user",
			logger.String("user_id", user.ID))
		return fmt.Errorf("no rows affected when creating user")
	}

	r.log.Info("Successfully created user",
		logger.String("user_id", user.ID))
	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	r.log.Info("Getting user by email",
		logger.String("email", email))

	query := `
		SELECT id, username, email, password, role
		FROM users
		WHERE email = ?
		LIMIT 1
	`

	var user entity.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.Warn("User not found",
				logger.String("email", email))
			return nil, nil
		}
		r.log.Error("Failed to get user",
			logger.String("email", email),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	r.log.Info("Successfully got user",
		logger.String("user_id", user.ID),
		logger.String("email", email))
	return &user, nil
}
