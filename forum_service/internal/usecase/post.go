package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kprf42/dolgova/forum_service/internal/entity"
	"github.com/kprf42/dolgova/forum_service/internal/repository"
	"github.com/kprf42/dolgova/pkg/logger"
)

type PostUseCase struct {
	postRepo *repository.PostRepository
	log      *logger.Logger
}

func NewPostUseCase(postRepo *repository.PostRepository, log *logger.Logger) *PostUseCase {
	return &PostUseCase{
		postRepo: postRepo,
		log:      log,
	}
}

func (uc *PostUseCase) Create(ctx context.Context, req *entity.PostRequest, authorID string) (*entity.PostResponse, error) {
	uc.log.Info("Creating new post",
		logger.String("title", req.Title),
		logger.String("author_id", authorID),
		logger.String("category_id", req.CategoryID))

	post := &entity.Post{
		ID:         uuid.New().String(),
		Title:      req.Title,
		Content:    req.Content,
		AuthorID:   authorID,
		CategoryID: req.CategoryID,
		IsPinned:   false,
		CreatedAt:  time.Now(),
	}

	uc.log.Debug("Generated post details",
		logger.String("post_id", post.ID),
		logger.String("title", post.Title))

	if err := uc.postRepo.Create(ctx, post); err != nil {
		uc.log.Error("Failed to create post",
			logger.String("post_id", post.ID),
			logger.Error(err))
		return nil, err
	}

	uc.log.Info("Successfully created post",
		logger.String("post_id", post.ID))

	return &entity.PostResponse{
		ID:         post.ID,
		Title:      post.Title,
		Content:    post.Content,
		AuthorID:   post.AuthorID,
		CategoryID: post.CategoryID,
		IsPinned:   post.IsPinned,
		CreatedAt:  post.CreatedAt,
	}, nil
}

func (uc *PostUseCase) GetByID(ctx context.Context, id string) (*entity.PostResponse, error) {
	uc.log.Info("Getting post by ID",
		logger.String("post_id", id))

	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		uc.log.Error("Failed to get post",
			logger.String("post_id", id),
			logger.Error(err))
		return nil, err
	}

	uc.log.Info("Successfully got post",
		logger.String("post_id", id))

	return &entity.PostResponse{
		ID:         post.ID,
		Title:      post.Title,
		Content:    post.Content,
		AuthorID:   post.AuthorID,
		CategoryID: post.CategoryID,
		IsPinned:   post.IsPinned,
		CreatedAt:  post.CreatedAt,
	}, nil
}

func (uc *PostUseCase) GetAll(ctx context.Context, limit, offset int, categoryID string) ([]*entity.PostResponse, int, error) {
	uc.log.Info("Getting all posts",
		logger.Int("limit", limit),
		logger.Int("offset", offset),
		logger.String("category_id", categoryID))

	posts, err := uc.postRepo.GetAll(ctx, limit, offset, categoryID)
	if err != nil {
		uc.log.Error("Failed to get posts",
			logger.Error(err))
		return nil, 0, err
	}

	total, err := uc.postRepo.Count(ctx, categoryID)
	if err != nil {
		uc.log.Error("Failed to count posts",
			logger.Error(err))
		return nil, 0, err
	}

	var responses []*entity.PostResponse
	for _, post := range posts {
		responses = append(responses, &entity.PostResponse{
			ID:         post.ID,
			Title:      post.Title,
			Content:    post.Content,
			AuthorID:   post.AuthorID,
			CategoryID: post.CategoryID,
			IsPinned:   post.IsPinned,
			CreatedAt:  post.CreatedAt,
		})
	}

	uc.log.Info("Successfully got posts",
		logger.Int("count", len(responses)),
		logger.Int("total", total))

	return responses, total, nil
}

func (uc *PostUseCase) Update(ctx context.Context, id string, req *entity.PostUpdate, authorID string) (*entity.PostResponse, error) {
	uc.log.Info("Updating post",
		logger.String("post_id", id),
		logger.String("author_id", authorID))

	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		uc.log.Error("Failed to get post for update",
			logger.String("post_id", id),
			logger.Error(err))
		return nil, err
	}

	if post.AuthorID != authorID {
		uc.log.Warn("Unauthorized post update attempt",
			logger.String("post_id", id),
			logger.String("author_id", authorID),
			logger.String("post_author_id", post.AuthorID))
		return nil, errors.New("unauthorized")
	}

	if err := uc.postRepo.Update(ctx, id, req); err != nil {
		uc.log.Error("Failed to update post",
			logger.String("post_id", id),
			logger.Error(err))
		return nil, err
	}

	updatedPost, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		uc.log.Error("Failed to get updated post",
			logger.String("post_id", id),
			logger.Error(err))
		return nil, err
	}

	uc.log.Info("Successfully updated post",
		logger.String("post_id", id))

	return &entity.PostResponse{
		ID:         updatedPost.ID,
		Title:      updatedPost.Title,
		Content:    updatedPost.Content,
		AuthorID:   updatedPost.AuthorID,
		CategoryID: updatedPost.CategoryID,
		IsPinned:   updatedPost.IsPinned,
		CreatedAt:  updatedPost.CreatedAt,
	}, nil
}

func (uc *PostUseCase) Delete(ctx context.Context, id string, authorID string) error {
	uc.log.Info("Deleting post",
		logger.String("post_id", id),
		logger.String("author_id", authorID))

	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		uc.log.Error("Failed to get post for deletion",
			logger.String("post_id", id),
			logger.Error(err))
		return err
	}

	if post.AuthorID != authorID {
		uc.log.Warn("Unauthorized post deletion attempt",
			logger.String("post_id", id),
			logger.String("author_id", authorID),
			logger.String("post_author_id", post.AuthorID))
		return errors.New("unauthorized")
	}

	if err := uc.postRepo.Delete(ctx, id); err != nil {
		uc.log.Error("Failed to delete post",
			logger.String("post_id", id),
			logger.Error(err))
		return err
	}

	uc.log.Info("Successfully deleted post",
		logger.String("post_id", id))

	return nil
}
