package grpc

import (
	"context"
	"time"

	appgrpc "github.com/kiribu/jwt-practice/services/gateway/app/grpc"
	authpb "github.com/kiribu/jwt-practice/shared/proto/auth"
	gogrpc "google.golang.org/grpc"
)

// AuthClient
type AuthClient struct {
	conn   *gogrpc.ClientConn
	client authpb.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := appgrpc.Dial(addr, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		conn:   conn,
		client: authpb.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}
func (c *AuthClient) Register(ctx context.Context, username, password string) (*authpb.RegisterResponse, error) {
	return c.client.Register(ctx, &authpb.RegisterRequest{
		Username: username,
		Password: password,
	})
}

func (c *AuthClient) Login(ctx context.Context, username, password string) (*authpb.LoginResponse, error) {
	return c.client.Login(ctx, &authpb.LoginRequest{
		Username: username,
		Password: password,
	})
}

func (c *AuthClient) Refresh(ctx context.Context, refreshToken string) (*authpb.RefreshResponse, error) {
	return c.client.Refresh(ctx, &authpb.RefreshRequest{
		RefreshToken: refreshToken,
	})
}

func (c *AuthClient) ValidateToken(ctx context.Context, accessToken string) (*authpb.ValidateTokenResponse, error) {
	return c.client.ValidateToken(ctx, &authpb.ValidateTokenRequest{
		AccessToken: accessToken,
	})
}

func (c *AuthClient) Logout(ctx context.Context, token string) (*authpb.LogoutResponse, error) {
	return c.client.Logout(ctx, &authpb.LogoutRequest{
		Token: token,
	})
}

func (c *AuthClient) GetProfile(ctx context.Context, username string) (*authpb.UserResponse, error) {
	return c.client.GetProfile(ctx, &authpb.GetProfileRequest{
		Username: username,
	})
}
