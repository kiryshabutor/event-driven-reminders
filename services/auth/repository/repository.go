package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	serviceentity "github.com/kiribu/jwt-practice/services/auth/entity/service"
)

type Repository interface {
	CreateUser(ctx context.Context, username, password string) (*serviceentity.User, error)
	GetUserByUsername(ctx context.Context, username string) (*serviceentity.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*serviceentity.User, error)
	ValidatePassword(ctx context.Context, username, password string) (*serviceentity.User, error)
	SaveRefreshToken(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error
	ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}
