package grpcdel

import (
	"context"
	"time"

	"github.com/kprf42/dolgova/forum_service/internal/entity"
	chat "github.com/kprf42/dolgova/forum_service/internal/usecase"
	comment "github.com/kprf42/dolgova/forum_service/internal/usecase"
	post "github.com/kprf42/dolgova/forum_service/internal/usecase"
	"github.com/kprf42/dolgova/proto/forum"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ForumServer struct {
	forum.UnimplementedForumServiceServer
	postUC    *post.PostUseCase
	commentUC *comment.CommentUseCase
	chatUC    *chat.ChatUseCase
}

func NewForumServer(
	postUC *post.PostUseCase,
	commentUC *comment.CommentUseCase,
	chatUC *chat.ChatUseCase,
) *ForumServer {
	return &ForumServer{
		postUC:    postUC,
		commentUC: commentUC,
		chatUC:    chatUC,
	}
}

func (s *ForumServer) CreatePost(ctx context.Context, req *forum.CreatePostRequest) (*forum.PostResponse, error) {
	postReq := &entity.PostRequest{
		Title:      req.Title,
		Content:    req.Content,
		CategoryID: req.CategoryId,
	}

	response, err := s.postUC.Create(ctx, postReq, req.AuthorId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create post: %v", err)
	}

	return &forum.PostResponse{
		Id:         response.ID,
		Title:      response.Title,
		Content:    response.Content,
		AuthorId:   response.AuthorID,
		CategoryId: response.CategoryID,
		CreatedAt:  response.CreatedAt.Format(time.RFC3339),
		IsPinned:   response.IsPinned,
	}, nil
}

func (s *ForumServer) GetPost(ctx context.Context, req *forum.GetPostRequest) (*forum.PostResponse, error) {
	post, err := s.postUC.GetByID(ctx, req.PostId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "post not found: %v", err)
	}

	return &forum.PostResponse{
		Id:         post.ID,
		Title:      post.Title,
		Content:    post.Content,
		AuthorId:   post.AuthorID,
		CategoryId: post.CategoryID,
		CreatedAt:  post.CreatedAt.Format(time.RFC3339),
		IsPinned:   post.IsPinned,
	}, nil
}

func (s *ForumServer) GetPosts(ctx context.Context, req *forum.GetPostsRequest) (*forum.GetPostsResponse, error) {
	posts, total, err := s.postUC.GetAll(ctx, int(req.Limit), int(req.Offset), req.CategoryId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get posts: %v", err)
	}

	var responses []*forum.PostResponse
	for _, post := range posts {
		responses = append(responses, &forum.PostResponse{
			Id:         post.ID,
			Title:      post.Title,
			Content:    post.Content,
			AuthorId:   post.AuthorID,
			CategoryId: post.CategoryID,
			CreatedAt:  post.CreatedAt.Format(time.RFC3339),
			IsPinned:   post.IsPinned,
		})
	}

	return &forum.GetPostsResponse{
		Posts: responses,
		Total: int32(total),
	}, nil
}

func (s *ForumServer) CreateComment(ctx context.Context, req *forum.CreateCommentRequest) (*forum.CommentResponse, error) {
	commentReq := &entity.CommentRequest{
		Content: req.Content,
		PostID:  req.PostId,
	}

	comment, err := s.commentUC.Create(ctx, commentReq, req.AuthorId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create comment: %v", err)
	}

	return &forum.CommentResponse{
		Id:        comment.ID,
		Content:   comment.Content,
		PostId:    comment.PostID,
		AuthorId:  comment.AuthorID,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ForumServer) GetComments(ctx context.Context, req *forum.GetCommentsRequest) (*forum.GetCommentsResponse, error) {
	comments, total, err := s.commentUC.GetByPostID(ctx, req.PostId, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get comments: %v", err)
	}

	var responses []*forum.CommentResponse
	for _, comment := range comments {
		responses = append(responses, &forum.CommentResponse{
			Id:        comment.ID,
			Content:   comment.Content,
			PostId:    comment.PostID,
			AuthorId:  comment.AuthorID,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		})
	}

	return &forum.GetCommentsResponse{
		Comments: responses,
		Total:    int32(total),
	}, nil
}

func (s *ForumServer) GetChatMessages(ctx context.Context, req *forum.GetChatMessagesRequest) (*forum.GetChatMessagesResponse, error) {
	messages, err := s.chatUC.GetMessages(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get chat messages: %v", err)
	}

	var responses []*forum.ChatMessage
	for _, msg := range messages {
		responses = append(responses, &forum.ChatMessage{
			Id:        msg.ID,
			UserId:    msg.UserID,
			Text:      msg.Text,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
		})
	}

	return &forum.GetChatMessagesResponse{
		Messages: responses,
		Total:    int32(len(responses)),
	}, nil
}
