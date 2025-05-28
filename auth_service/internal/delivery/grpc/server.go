package auth

import (
	"context"
	"errors"

	"github.com/kprf42/dolgova/auth_service/internal/entity"
	"github.com/kprf42/dolgova/auth_service/internal/usecase/auth"
	"github.com/kprf42/dolgova/auth_service/internal/usecase/jwt"
	proto "github.com/kprf42/dolgova/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	proto.UnimplementedAuthServiceServer
	authUC *auth.AuthUseCase
	jwtUC  jwt.JWTUseCase
}

func NewAuthServer(authUC *auth.AuthUseCase, jwtUC jwt.JWTUseCase) *AuthServer {
	return &AuthServer{authUC: authUC, jwtUC: jwtUC}
}

func (s *AuthServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	// Валидация запроса
	if req.GetUsername() == "" || req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "username, email and password are required")
	}

	// Вызов use case
	user, err := s.authUC.Register(ctx, req.GetUsername(), req.GetEmail(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrUserAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
		case errors.Is(err, entity.ErrInvalidEmail):
			return nil, status.Error(codes.InvalidArgument, "invalid email format")
		case errors.Is(err, entity.ErrWeakPassword):
			return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters")
		default:
			return nil, status.Error(codes.Internal, "failed to register user")
		}
	}

	return &proto.RegisterResponse{
		UserId: user.ID,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	// Валидация запроса
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	// Вызов use case
	tokens, err := s.authUC.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		// Для безопасности возвращаем одинаковую ошибку
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	return &proto.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.AtExpires,
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	claims, err := s.jwtUC.ValidateToken(req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &proto.ValidateTokenResponse{
		UserId: claims.UserID,
		Valid:  true,
	}, nil
}
