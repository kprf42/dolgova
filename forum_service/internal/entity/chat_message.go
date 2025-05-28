package entity

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id" validate:"required,uuid4"`
	Text      string    `json:"text" db:"text" validate:"required,min=1,max=1000"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ChatMessageRequest struct {
	Text string `json:"text" validate:"required,min=1,max=1000"`
}

func NewChatMessage(req *ChatMessageRequest, userID string) *ChatMessage {
	return &ChatMessage{
		ID:        uuid.New().String(),
		UserID:    userID,
		Text:      req.Text,
		CreatedAt: time.Now().UTC(),
	}
}
