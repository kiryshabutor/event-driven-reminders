package grpc

import (
	"context"

	"github.com/kiribu/jwt-practice/services/auth/service"
	authpb "github.com/kiribu/jwt-practice/shared/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer
type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	service *service.AuthService
}

func NewAuthServer(svc *service.AuthService) *AuthServer {
	return &AuthServer{service: svc}
}

func (s *AuthServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	user, err := s.service.Register(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	return &authpb.RegisterResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	tokens, err := s.service.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	return &authpb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
	}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *authpb.RefreshRequest) (*authpb.RefreshResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	tokens, err := s.service.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	return &authpb.RefreshResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return &authpb.ValidateTokenResponse{
			Valid: false,
			Error: "access_token is required",
		}, nil
	}

	username, userID, err := s.service.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:    true,
		Username: username,
		UserId:   userID.String(),
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	if err := s.service.Logout(ctx, req.Token); err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &authpb.LogoutResponse{Success: true}, nil
}

func (s *AuthServer) GetProfile(ctx context.Context, req *authpb.GetProfileRequest) (*authpb.UserResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	user, err := s.service.GetProfile(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authpb.UserResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
