package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"regexp"
	"time"

	"github.com/google/uuid"
	serviceentity "github.com/kiribu/jwt-practice/services/auth/entity/service"
	"github.com/kiribu/jwt-practice/services/auth/repository"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	store repository.Repository
	redis *redis.Client
}

func NewAuthService(store repository.Repository, redisClient *redis.Client) *AuthService {
	return &AuthService{
		store: store,
		redis: redisClient,
	}
}

type UserResponse struct {
	ID        uuid.UUID
	Username  string
	CreatedAt time.Time
}
type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
}

type cachedUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,255}$`)
	passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!#$%*]{8,16}$`)

	// Password complexity checks
	hasLetterRegex  = regexp.MustCompile(`[a-zA-Z]`)
	hasDigitRegex   = regexp.MustCompile(`[0-9]`)
	hasSpecialRegex = regexp.MustCompile(`[!#$%*]`)
)

func (s *AuthService) Register(ctx context.Context, username, password string) (*UserResponse, error) {
	if err := s.validateCredentials(username, password); err != nil {
		return nil, err
	}

	user, err := s.store.CreateUser(ctx, username, password)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*TokenResponse, error) {
	user, err := s.store.ValidatePassword(ctx, username, password)
	if err != nil {
		return nil, err
	}

	accessToken, err := GenerateAccessToken(user.Username, user.ID.String())
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(RefreshTokenDuration)
	if err := s.store.SaveRefreshToken(ctx, refreshToken, user.ID, expiresAt); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	userID, err := s.store.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	accessToken, err := GenerateAccessToken(user.Username, user.ID.String())
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	s.store.DeleteRefreshToken(ctx, refreshToken)
	expiresAt := time.Now().Add(RefreshTokenDuration)
	if err := s.store.SaveRefreshToken(ctx, newRefreshToken, user.ID, expiresAt); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (string, uuid.UUID, error) {
	// Check Blacklist
	val, err := s.redis.Get(ctx, "blacklist:"+token).Result()
	if err == nil && val == "revoked" {
		slog.Warn("Blacklist hit for token", "token", token)
		return "", uuid.Nil, errors.New("token revoked")
	}

	claims, err := ValidateAccessToken(token)
	if err != nil {
		return "", uuid.Nil, err
	}

	// Check User Cache
	cacheKey := "user:" + claims.Username
	val, err = s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache Hit
		slog.Debug("Cache hit for user", "username", claims.Username)
		var user cachedUser
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			return user.Username, user.ID, nil
		}
	}

	// Cache Miss
	slog.Debug("Cache miss for user", "username", claims.Username)
	user, err := s.store.GetUserByUsername(ctx, claims.Username)
	if err != nil {
		return "", uuid.Nil, err
	}

	// Set Cache
	if userJSON, err := json.Marshal(toCachedUser(user)); err == nil {
		s.redis.Set(ctx, cacheKey, userJSON, AccessTokenDuration)
	}

	return claims.Username, user.ID, nil
}

func (s *AuthService) GetProfile(ctx context.Context, username string) (*UserResponse, error) {
	// Check Cache
	cacheKey := "user:" + username
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache Hit
		slog.Debug("Cache hit for user profile", "username", username)
		var user cachedUser
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			return &UserResponse{
				ID:        user.ID,
				Username:  user.Username,
				CreatedAt: user.CreatedAt,
			}, nil
		}
	}

	// Cache Miss
	slog.Debug("Cache miss for user profile", "username", username)
	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Set Cache
	if userJSON, err := json.Marshal(toCachedUser(user)); err == nil {
		s.redis.Set(ctx, cacheKey, userJSON, AccessTokenDuration)
	}

	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.redis.Set(ctx, "blacklist:"+token, "revoked", AccessTokenDuration).Err()
}

func (s *AuthService) validateCredentials(username, password string) error {
	if !usernameRegex.MatchString(username) {
		return errors.New("invalid username format: must be 3-255 alphanumeric characters or underscore")
	}

	if !passwordRegex.MatchString(password) {
		return errors.New("invalid password format: must be 8-16 characters and contain only alphanumeric or !#$%*")
	}

	if !hasLetterRegex.MatchString(password) || !hasDigitRegex.MatchString(password) || !hasSpecialRegex.MatchString(password) {
		return errors.New("password must contain at least one letter, one digit, and one special character (!#$%*)")
	}

	return nil
}

func toCachedUser(user *serviceentity.User) cachedUser {
	return cachedUser{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}
}
