package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	repositoryentity "github.com/kiribu/jwt-practice/services/reminder/entity/repository"
	serviceentity "github.com/kiribu/jwt-practice/services/reminder/entity/service"
)

type Repository interface {
	Create(ctx context.Context, userID uuid.UUID, title, description string, remindAt time.Time) (*serviceentity.Reminder, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, status string) ([]serviceentity.Reminder, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*serviceentity.Reminder, error)
	Update(ctx context.Context, userID, id uuid.UUID, title, description string, remindAt time.Time) (*serviceentity.Reminder, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	GetPending(ctx context.Context) ([]serviceentity.Reminder, error)
	MarkAsSent(ctx context.Context, id uuid.UUID) error
	GetPendingOutboxEvents(ctx context.Context, limit int) ([]repositoryentity.OutboxEvent, error)
	MarkOutboxEventAsSent(ctx context.Context, id uuid.UUID) error
	IncrementOutboxRetryCount(ctx context.Context, id uuid.UUID, errMsg string) error
	CreateNotificationEventsAndMarkSent(ctx context.Context, reminder serviceentity.Reminder) error
}
