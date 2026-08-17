package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	serviceentity "github.com/kiribu/jwt-practice/services/analytics/entity/service"
	"gorm.io/gorm"
)

type Repository interface {
	GetUserStats(ctx context.Context, userID uuid.UUID) (*serviceentity.UserStatistics, error)
	BeginTx(ctx context.Context) *gorm.DB
	IsEventProcessed(ctx context.Context, tx *gorm.DB, eventID uuid.UUID) (bool, error)
	MarkEventProcessed(ctx context.Context, tx *gorm.DB, eventID uuid.UUID) error
	IncrementCreated(ctx context.Context, tx *gorm.DB, userID uuid.UUID, timestamp time.Time) error
	IncrementCompleted(ctx context.Context, tx *gorm.DB, userID uuid.UUID, timestamp time.Time) error
	IncrementDeleted(ctx context.Context, tx *gorm.DB, userID uuid.UUID, timestamp time.Time) error
}
