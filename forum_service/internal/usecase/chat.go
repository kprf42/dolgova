package usecase

import (
	"context"
	"time"

	"github.com/kprf42/dolgova/forum_service/internal/entity"
	"github.com/kprf42/dolgova/forum_service/internal/repository"
	"github.com/kprf42/dolgova/pkg/logger"
)

type ChatUseCase struct {
	repo *repository.ChatRepository
	log  *logger.Logger
}

func NewChatUseCase(repo *repository.ChatRepository, log *logger.Logger) *ChatUseCase {
	return &ChatUseCase{
		repo: repo,
		log:  log,
	}
}

func (uc *ChatUseCase) SaveMessage(ctx context.Context, msg *entity.ChatMessage) error {
	uc.log.Info("Saving chat message",
		logger.String("message_id", msg.ID),
		logger.String("user_id", msg.UserID))

	if err := uc.repo.SaveMessage(ctx, msg); err != nil {
		uc.log.Error("Failed to save chat message",
			logger.String("message_id", msg.ID),
			logger.Error(err))
		return err
	}

	uc.log.Info("Successfully saved chat message",
		logger.String("message_id", msg.ID))

	return nil
}

func (uc *ChatUseCase) GetMessages(ctx context.Context, limit, offset int) ([]*entity.ChatMessage, error) {
	uc.log.Info("Getting chat messages",
		logger.Int("limit", limit),
		logger.Int("offset", offset))

	messages, err := uc.repo.GetMessages(ctx, limit, offset)
	if err != nil {
		uc.log.Error("Failed to get chat messages",
			logger.Error(err))
		return nil, err
	}

	uc.log.Info("Successfully got chat messages",
		logger.Int("count", len(messages)))

	return messages, nil
}

func (uc *ChatUseCase) CleanOldMessages(ctx context.Context, olderThan time.Duration) error {
	uc.log.Info("Cleaning old chat messages",
		logger.Float64("older_than_seconds", olderThan.Seconds()))

	if err := uc.repo.CleanOldMessages(ctx, olderThan); err != nil {
		uc.log.Error("Failed to clean old chat messages",
			logger.Float64("older_than_seconds", olderThan.Seconds()),
			logger.Error(err))
		return err
	}

	uc.log.Info("Successfully cleaned old chat messages")
	return nil
}
