package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kprf42/dolgova/forum_service/internal/entity"
	"github.com/kprf42/dolgova/pkg/logger"
)

type CommentRepository struct {
	db  *sql.DB
	log *logger.Logger
}

func NewCommentRepository(db *sql.DB, log *logger.Logger) *CommentRepository {
	return &CommentRepository{
		db:  db,
		log: log,
	}
}

func (r *CommentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	r.log.Info("Creating new comment",
		logger.String("comment_id", comment.ID),
		logger.String("post_id", comment.PostID),
		logger.String("author_id", comment.AuthorID))

	query := `INSERT INTO comments (id, content, post_id, author_id, created_at) 
	          VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query,
		comment.ID,
		comment.Content,
		comment.PostID,
		comment.AuthorID,
		comment.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		r.log.Error("Failed to create comment",
			logger.String("comment_id", comment.ID),
			logger.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected",
			logger.String("comment_id", comment.ID),
			logger.Error(err))
		return err
	}

	if rows == 0 {
		r.log.Error("No rows affected when creating comment",
			logger.String("comment_id", comment.ID))
		return fmt.Errorf("no rows affected when creating comment")
	}

	r.log.Info("Successfully created comment",
		logger.String("comment_id", comment.ID))
	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id string) (*entity.Comment, error) {
	r.log.Info("Getting comment by ID",
		logger.String("comment_id", id))

	query := `SELECT id, content, post_id, author_id, created_at 
	          FROM comments WHERE id = ?`

	var comment entity.Comment
	var createdAt string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.Content,
		&comment.PostID,
		&comment.AuthorID,
		&createdAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		r.log.Warn("Comment not found",
			logger.String("comment_id", id))
		return nil, fmt.Errorf("comment not found")
	}
	if err != nil {
		r.log.Error("Failed to get comment",
			logger.String("comment_id", id),
			logger.Error(err))
		return nil, err
	}

	comment.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		r.log.Error("Failed to parse created_at",
			logger.String("comment_id", id),
			logger.String("created_at", createdAt),
			logger.Error(err))
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	r.log.Info("Successfully got comment",
		logger.String("comment_id", id))
	return &comment, nil
}

func (r *CommentRepository) GetByPostID(ctx context.Context, postID string, limit, offset int) ([]*entity.Comment, error) {
	r.log.Info("Getting comments by post ID",
		logger.String("post_id", postID),
		logger.Int("limit", limit),
		logger.Int("offset", offset))

	query := `SELECT id, content, post_id, author_id, created_at 
	          FROM comments WHERE post_id = ? 
	          ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		r.log.Error("Failed to get comments",
			logger.String("post_id", postID),
			logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var comments []*entity.Comment
	for rows.Next() {
		var comment entity.Comment
		var createdAt string

		if err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.PostID,
			&comment.AuthorID,
			&createdAt,
		); err != nil {
			r.log.Error("Failed to scan comment row",
				logger.Error(err))
			return nil, err
		}

		comment.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			r.log.Error("Failed to parse created_at",
				logger.String("created_at", createdAt),
				logger.Error(err))
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		comments = append(comments, &comment)
	}

	r.log.Info("Successfully got comments",
		logger.String("post_id", postID),
		logger.Int("count", len(comments)))
	return comments, nil
}

func (r *CommentRepository) Update(ctx context.Context, id string, content string) error {
	r.log.Info("Updating comment",
		logger.String("comment_id", id))

	query := `UPDATE comments SET content = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, content, id)
	if err != nil {
		r.log.Error("Failed to update comment",
			logger.String("comment_id", id),
			logger.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected",
			logger.String("comment_id", id),
			logger.Error(err))
		return err
	}

	if rows == 0 {
		r.log.Warn("No rows affected when updating comment",
			logger.String("comment_id", id))
	} else {
		r.log.Info("Successfully updated comment",
			logger.String("comment_id", id))
	}

	return nil
}

func (r *CommentRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("Deleting comment",
		logger.String("comment_id", id))

	query := `DELETE FROM comments WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete comment",
			logger.String("comment_id", id),
			logger.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected",
			logger.String("comment_id", id),
			logger.Error(err))
		return err
	}

	if rows == 0 {
		r.log.Warn("No rows affected when deleting comment",
			logger.String("comment_id", id))
	} else {
		r.log.Info("Successfully deleted comment",
			logger.String("comment_id", id))
	}

	return nil
}

func (r *CommentRepository) CountByPostID(ctx context.Context, postID string) (int, error) {
	r.log.Info("Counting comments by post ID",
		logger.String("post_id", postID))

	query := `SELECT COUNT(*) FROM comments WHERE post_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, postID).Scan(&count)
	if err != nil {
		r.log.Error("Failed to count comments",
			logger.String("post_id", postID),
			logger.Error(err))
		return 0, err
	}

	r.log.Info("Successfully counted comments",
		logger.String("post_id", postID),
		logger.Int("count", count))
	return count, nil
}
