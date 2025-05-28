package usecase

import (
	"context"
	"errors"

	"github.com/kprf42/dolgova/forum_service/internal/entity"
	"github.com/kprf42/dolgova/forum_service/internal/repository"
	"github.com/kprf42/dolgova/pkg/logger"
)

type CommentUseCase struct {
	repo *repository.CommentRepository
	log  *logger.Logger
}

func NewCommentUseCase(repo *repository.CommentRepository, log *logger.Logger) *CommentUseCase {
	return &CommentUseCase{
		repo: repo,
		log:  log,
	}
}

func (uc *CommentUseCase) Create(ctx context.Context, req *entity.CommentRequest, authorID string) (*entity.Comment, error) {
	uc.log.Info("Creating new comment",
		logger.String("post_id", req.PostID),
		logger.String("author_id", authorID))

	comment := entity.NewComment(req, authorID)

	uc.log.Debug("Generated comment details",
		logger.String("comment_id", comment.ID),
		logger.String("post_id", comment.PostID))

	if err := uc.repo.Create(ctx, comment); err != nil {
		uc.log.Error("Failed to create comment",
			logger.String("comment_id", comment.ID),
			logger.Error(err))
		return nil, err
	}

	uc.log.Info("Successfully created comment",
		logger.String("comment_id", comment.ID))

	return comment, nil
}

func (uc *CommentUseCase) GetByID(ctx context.Context, id string) (*entity.Comment, error) {
	uc.log.Info("Getting comment by ID",
		logger.String("comment_id", id))

	comment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.log.Error("Failed to get comment",
			logger.String("comment_id", id),
			logger.Error(err))
		return nil, err
	}

	uc.log.Info("Successfully got comment",
		logger.String("comment_id", id))

	return comment, nil
}

func (uc *CommentUseCase) GetByPostID(ctx context.Context, postID string, limit, offset int) ([]*entity.Comment, int, error) {
	uc.log.Info("Getting comments by post ID",
		logger.String("post_id", postID),
		logger.Int("limit", limit),
		logger.Int("offset", offset))

	comments, err := uc.repo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		uc.log.Error("Failed to get comments",
			logger.String("post_id", postID),
			logger.Error(err))
		return nil, 0, err
	}

	total, err := uc.repo.CountByPostID(ctx, postID)
	if err != nil {
		uc.log.Error("Failed to count comments",
			logger.String("post_id", postID),
			logger.Error(err))
		return nil, 0, err
	}

	uc.log.Info("Successfully got comments",
		logger.String("post_id", postID),
		logger.Int("count", len(comments)),
		logger.Int("total", total))

	return comments, total, nil
}

func (uc *CommentUseCase) Update(ctx context.Context, id string, content string, authorID string) (*entity.Comment, error) {
	uc.log.Info("Updating comment",
		logger.String("comment_id", id),
		logger.String("author_id", authorID))

	comment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.log.Error("Failed to get comment for update",
			logger.String("comment_id", id),
			logger.Error(err))
		return nil, err
	}

	if comment.AuthorID != authorID {
		uc.log.Warn("Unauthorized comment update attempt",
			logger.String("comment_id", id),
			logger.String("author_id", authorID),
			logger.String("comment_author_id", comment.AuthorID))
		return nil, errors.New("unauthorized")
	}

	if err := uc.repo.Update(ctx, id, content); err != nil {
		uc.log.Error("Failed to update comment",
			logger.String("comment_id", id),
			logger.Error(err))
		return nil, err
	}

	updatedComment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.log.Error("Failed to get updated comment",
			logger.String("comment_id", id),
			logger.Error(err))
		return nil, err
	}

	uc.log.Info("Successfully updated comment",
		logger.String("comment_id", id))

	return updatedComment, nil
}

func (uc *CommentUseCase) Delete(ctx context.Context, id string, authorID string) error {
	uc.log.Info("Deleting comment",
		logger.String("comment_id", id),
		logger.String("author_id", authorID))

	comment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.log.Error("Failed to get comment for deletion",
			logger.String("comment_id", id),
			logger.Error(err))
		return err
	}

	if comment.AuthorID != authorID {
		uc.log.Warn("Unauthorized comment deletion attempt",
			logger.String("comment_id", id),
			logger.String("author_id", authorID),
			logger.String("comment_author_id", comment.AuthorID))
		return errors.New("unauthorized")
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		uc.log.Error("Failed to delete comment",
			logger.String("comment_id", id),
			logger.Error(err))
		return err
	}

	uc.log.Info("Successfully deleted comment",
		logger.String("comment_id", id))

	return nil
}
