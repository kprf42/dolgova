package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kprf42/dolgova/forum_service/internal/entity"
	"github.com/kprf42/dolgova/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
)

type PostRepository struct {
	db  *sql.DB
	log *logger.Logger
}

func NewPostRepository(db *sql.DB, log *logger.Logger) *PostRepository {
	return &PostRepository{
		db:  db,
		log: log,
	}
}

func (r *PostRepository) Create(ctx context.Context, post *entity.Post) error {
	r.log.Info("Creating new post",
		logger.String("post_id", post.ID),
		logger.String("title", post.Title),
		logger.String("author_id", post.AuthorID),
		logger.String("category_id", post.CategoryID))

	query := `INSERT INTO posts (id, title, content, author_id, category_id, is_pinned, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		post.ID,
		post.Title,
		post.Content,
		post.AuthorID,
		post.CategoryID,
		post.IsPinned,
		post.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		r.log.Error("Failed to create post",
			logger.String("post_id", post.ID),
			logger.Error(err))
		return fmt.Errorf("failed to create post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected",
			logger.String("post_id", post.ID),
			logger.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		r.log.Error("No rows affected when creating post",
			logger.String("post_id", post.ID))
		return fmt.Errorf("no rows affected when creating post")
	}

	r.log.Info("Successfully created post",
		logger.String("post_id", post.ID))
	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (*entity.Post, error) {
	r.log.Info("Getting post by ID",
		logger.String("post_id", id))

	query := `SELECT id, title, content, author_id, category_id, is_pinned, created_at 
	          FROM posts WHERE id = ?`

	var post entity.Post
	var createdAt string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.CategoryID,
		&post.IsPinned,
		&createdAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		r.log.Warn("Post not found",
			logger.String("post_id", id))
		return nil, fmt.Errorf("post not found")
	}
	if err != nil {
		r.log.Error("Failed to get post",
			logger.String("post_id", id),
			logger.Error(err))
		return nil, err
	}

	post.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		r.log.Error("Failed to parse created_at",
			logger.String("post_id", id),
			logger.String("created_at", createdAt),
			logger.Error(err))
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	r.log.Info("Successfully got post",
		logger.String("post_id", id))
	return &post, nil
}

func (r *PostRepository) GetAll(ctx context.Context, limit, offset int, categoryID string) ([]*entity.Post, error) {
	r.log.Info("Getting all posts",
		logger.Int("limit", limit),
		logger.Int("offset", offset),
		logger.String("category_id", categoryID))

	var query string
	var args []interface{}

	if categoryID != "" {
		query = `SELECT id, title, content, author_id, category_id, is_pinned, created_at 
		         FROM posts WHERE category_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = []interface{}{categoryID, limit, offset}
	} else {
		query = `SELECT id, title, content, author_id, category_id, is_pinned, created_at 
		         FROM posts ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.log.Error("Failed to get posts",
			logger.Int("limit", limit),
			logger.Int("offset", offset),
			logger.String("category_id", categoryID),
			logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var posts []*entity.Post
	for rows.Next() {
		var post entity.Post
		var createdAt string

		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.CategoryID,
			&post.IsPinned,
			&createdAt,
		); err != nil {
			r.log.Error("Failed to scan post row",
				logger.Error(err))
			return nil, err
		}

		post.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			r.log.Error("Failed to parse created_at",
				logger.String("created_at", createdAt),
				logger.Error(err))
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		posts = append(posts, &post)
	}

	r.log.Info("Successfully got posts",
		logger.Int("count", len(posts)))
	return posts, nil
}

func (r *PostRepository) Update(ctx context.Context, id string, post *entity.PostUpdate) error {
	r.log.Info("Updating post",
		logger.String("post_id", id))

	query := `UPDATE posts SET title = ?, content = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, post.Title, post.Content, id)
	if err != nil {
		r.log.Error("Failed to update post",
			logger.String("post_id", id),
			logger.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected",
			logger.String("post_id", id),
			logger.Error(err))
		return err
	}

	if rows == 0 {
		r.log.Warn("No rows affected when updating post",
			logger.String("post_id", id))
	} else {
		r.log.Info("Successfully updated post",
			logger.String("post_id", id))
	}

	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("Deleting post",
		logger.String("post_id", id))

	query := `DELETE FROM posts WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete post",
			logger.String("post_id", id),
			logger.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected",
			logger.String("post_id", id),
			logger.Error(err))
		return err
	}

	if rows == 0 {
		r.log.Warn("No rows affected when deleting post",
			logger.String("post_id", id))
	} else {
		r.log.Info("Successfully deleted post",
			logger.String("post_id", id))
	}

	return nil
}

func (r *PostRepository) Count(ctx context.Context, categoryID string) (int, error) {
	r.log.Info("Counting posts",
		logger.String("category_id", categoryID))

	var query string
	var args []interface{}

	if categoryID != "" {
		query = `SELECT COUNT(*) FROM posts WHERE category_id = ?`
		args = []interface{}{categoryID}
	} else {
		query = `SELECT COUNT(*) FROM posts`
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		r.log.Error("Failed to count posts",
			logger.String("category_id", categoryID),
			logger.Error(err))
		return 0, err
	}

	r.log.Info("Successfully counted posts",
		logger.Int("count", count),
		logger.String("category_id", categoryID))
	return count, nil
}
