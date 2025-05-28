package entity

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        string    `json:"id"`
	Content   string    `json:"content" validate:"required,min=3,max=500"`
	PostID    string    `json:"post_id" validate:"required,uuid4"`
	AuthorID  string    `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentRequest struct {
	Content string `json:"content" validate:"required,min=3,max=500"`
	PostID  string `json:"post_id" validate:"required,uuid4"`
}

type CommentResponse struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	PostID    string    `json:"post_id"`
	AuthorID  string    `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewComment(req *CommentRequest, authorID string) *Comment {
	return &Comment{
		ID:        uuid.New().String(),
		Content:   req.Content,
		PostID:    req.PostID,
		AuthorID:  authorID,
		CreatedAt: time.Now().UTC(),
	}
}
