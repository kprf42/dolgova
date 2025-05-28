package entity

import "time"

type Post struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	AuthorID   string    `json:"author_id"`
	CategoryID string    `json:"category_id"`
	IsPinned   bool      `json:"is_pinned"`
	CreatedAt  time.Time `json:"created_at"`
}

type PostRequest struct {
	Title      string `json:"title" validate:"required,min=3,max=100"`
	Content    string `json:"content" validate:"required,min=10"`
	CategoryID string `json:"category_id" validate:"required"`
}

type PostUpdate struct {
	Title   string `json:"title" validate:"required,min=3,max=100"`
	Content string `json:"content" validate:"required,min=10"`
}

type PostResponse struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	AuthorID   string    `json:"author_id"`
	CategoryID string    `json:"category_id"`
	IsPinned   bool      `json:"is_pinned"`
	CreatedAt  time.Time `json:"created_at"`
}

type PostErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type Claims struct {
	UserID string `json:"user_id"`
}
