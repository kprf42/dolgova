package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kprf42/dolgova/forum_service/internal/entity"
	"github.com/kprf42/dolgova/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
)

type ChatRepository struct {
	db  *sql.DB
	log *logger.Logger
}

func NewChatRepository(db *sql.DB, log *logger.Logger) *ChatRepository {
	return &ChatRepository{
		db:  db,
		log: log,
	}
}

func (r *ChatRepository) SaveMessage(ctx context.Context, msg *entity.ChatMessage) error {
	r.log.Info("Saving chat message",
		logger.String("message_id", msg.ID),
		logger.String("user_id", msg.UserID))

	query := `INSERT INTO chat_messages (id, user_id, text, created_at) VALUES (?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, msg.ID, msg.UserID, msg.Text, msg.CreatedAt.Format(time.RFC3339))
	if err != nil {
		r.log.Error("Failed to save chat message",
			logger.String("message_id", msg.ID),
			logger.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected",
			logger.String("message_id", msg.ID),
			logger.Error(err))
		return err
	}

	if rows == 0 {
		r.log.Error("No rows affected when saving chat message",
			logger.String("message_id", msg.ID))
		return fmt.Errorf("no rows affected when saving chat message")
	}

	r.log.Info("Successfully saved chat message",
		logger.String("message_id", msg.ID))
	return nil
}

func (r *ChatRepository) GetMessages(ctx context.Context, limit, offset int) ([]*entity.ChatMessage, error) {
	r.log.Info("Getting chat messages",
		logger.Int("limit", limit),
		logger.Int("offset", offset))

	query := `SELECT id, user_id, text, created_at FROM chat_messages 
	          ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.log.Error("Failed to get chat messages",
			logger.Int("limit", limit),
			logger.Int("offset", offset),
			logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var messages []*entity.ChatMessage
	for rows.Next() {
		var msg entity.ChatMessage
		var createdAt string

		if err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.Text,
			&createdAt,
		); err != nil {
			r.log.Error("Failed to scan chat message row",
				logger.Error(err))
			return nil, err
		}

		msg.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			r.log.Error("Failed to parse created_at",
				logger.String("created_at", createdAt),
				logger.Error(err))
			return nil, err
		}

		messages = append(messages, &msg)
	}

	r.log.Info("Successfully got chat messages",
		logger.Int("count", len(messages)))
	return messages, nil
}

func (r *ChatRepository) CleanOldMessages(ctx context.Context, olderThan time.Duration) error {
	r.log.Info("Cleaning old chat messages",
		logger.Float64("older_than_seconds", olderThan.Seconds()))

	result, err := r.db.ExecContext(ctx,
		`DELETE FROM chat_messages WHERE created_at < datetime('now', ?)`,
		fmt.Sprintf("-%d seconds", int(olderThan.Seconds())))
	if err != nil {
		r.log.Error("Failed to clean old chat messages",
			logger.Float64("older_than_seconds", olderThan.Seconds()),
			logger.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected",
			logger.Error(err))
		return err
	}

	r.log.Info("Successfully cleaned old chat messages",
		logger.Int64("deleted_count", rows))
	return nil
}
