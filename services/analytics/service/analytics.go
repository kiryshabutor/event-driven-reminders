package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	serviceentity "github.com/kiribu/jwt-practice/services/analytics/entity/service"
	"github.com/kiribu/jwt-practice/services/analytics/repository"
	"github.com/kiribu/jwt-practice/shared/consts"
	"github.com/kiribu/jwt-practice/shared/types"
	"gorm.io/gorm"
)

type AnalyticsService struct {
	storage repository.Repository
}

func NewAnalyticsService(storage repository.Repository) *AnalyticsService {
	return &AnalyticsService{storage: storage}
}

func (s *AnalyticsService) ProcessEvent(ctx context.Context, event types.LifecycleEvent) error {
	slog.Info("Processing event", "event_id", event.EventID, "type", event.EventType, "user_id", event.UserID)

	tx := s.storage.BeginTx(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	processed, err := s.storage.IsEventProcessed(ctx, tx, event.EventID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if processed {
		slog.Info("Event already processed, skipping", "event_id", event.EventID)
		tx.Rollback()
		return nil
	}

	switch event.EventType {
	case consts.EventCreated:
		err = s.storage.IncrementCreated(ctx, tx, event.UserID, event.Timestamp)
	case consts.EventUpdated:
		err = nil // No-op for updated
	case consts.EventNotificationSent:
		err = s.storage.IncrementCompleted(ctx, tx, event.UserID, event.Timestamp)
	case consts.EventDeleted:
		err = s.storage.IncrementDeleted(ctx, tx, event.UserID, event.Timestamp)
	default:
		slog.Warn("Unknown event type", "type", event.EventType)
		err = nil
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	if err := s.storage.MarkEventProcessed(ctx, tx, event.EventID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *AnalyticsService) GetUserStats(ctx context.Context, userID uuid.UUID) (*serviceentity.UserStatistics, error) {
	stats, err := s.storage.GetUserStats(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &serviceentity.UserStatistics{UserID: userID}, nil
		}
		return nil, err
	}
	return stats, nil
}
