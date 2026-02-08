package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kiribu/jwt-practice/models"
	"gorm.io/gorm"
)

type AnalyticsStorage interface {
	GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStatistics, error)

	BeginTx(ctx context.Context) *gorm.DB
	IsEventProcessed(ctx context.Context, tx *gorm.DB, eventID uuid.UUID) (bool, error)
	MarkEventProcessed(ctx context.Context, tx *gorm.DB, eventID uuid.UUID) error
	IncrementCreated(ctx context.Context, tx *gorm.DB, userID uuid.UUID, timestamp time.Time) error
	IncrementCompleted(ctx context.Context, tx *gorm.DB, userID uuid.UUID, timestamp time.Time) error
	IncrementDeleted(ctx context.Context, tx *gorm.DB, userID uuid.UUID, timestamp time.Time) error
}

type PostgresStorage struct {
	db *gorm.DB
}

func NewPostgresStorage(db *gorm.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStatistics, error) {
	var stats models.UserStatistics
	result := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&stats)
	if result.Error != nil {
		return nil, result.Error
	}
	return &stats, nil
}

func (s *PostgresStorage) BeginTx(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx).Begin()
}

func (s *PostgresStorage) IsEventProcessed(ctx context.Context, tx *gorm.DB, eventID uuid.UUID) (bool, error) {
	var count int64
	result := tx.Model(&models.ProcessedEvent{}).Where("event_id = ?", eventID).Count(&count)
	return count > 0, result.Error
}

func (s *PostgresStorage) MarkEventProcessed(ctx context.Context, tx *gorm.DB, eventID uuid.UUID) error {
	event := models.ProcessedEvent{
		EventID: eventID,
	}
	return tx.Create(&event).Error
}

func (s *PostgresStorage) IncrementCreated(ctx context.Context, tx *gorm.DB, userID uuid.UUID, timestamp time.Time) error {
	query := `
		INSERT INTO analytics.user_statistics (user_id, total_reminders_created, active_reminders, first_reminder_at, last_activity_at)
		VALUES ($1, 1, 1, $2, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			total_reminders_created = user_statistics.total_reminders_created + 1,
			active_reminders = user_statistics.active_reminders + 1,
			last_activity_at = $2,
			completion_rate = CASE 
				WHEN user_statistics.total_reminders_created + 1 > 0 
				THEN ROUND((user_statistics.total_reminders_completed::DECIMAL / (user_statistics.total_reminders_created + 1)) * 100, 2)
				ELSE 0 
			END,
			updated_at = NOW()
	`
	return tx.Exec(query, userID, timestamp).Error
}

func (s *PostgresStorage) IncrementCompleted(ctx context.Context, tx *gorm.DB, userID uuid.UUID, timestamp time.Time) error {
	query := `
		UPDATE analytics.user_statistics SET
			total_reminders_completed = total_reminders_completed + 1,
			active_reminders = GREATEST(active_reminders - 1, 0),
			last_activity_at = $2,
			completion_rate = CASE 
				WHEN total_reminders_created > 0 
				THEN ROUND(((total_reminders_completed + 1)::DECIMAL / total_reminders_created) * 100, 2)
				ELSE 0 
			END,
			updated_at = NOW()
		WHERE user_id = $1
	`
	return tx.Exec(query, userID, timestamp).Error
}

func (s *PostgresStorage) IncrementDeleted(ctx context.Context, tx *gorm.DB, userID uuid.UUID, timestamp time.Time) error {
	query := `
		UPDATE analytics.user_statistics SET
			total_reminders_deleted = total_reminders_deleted + 1,
			active_reminders = GREATEST(active_reminders - 1, 0),
			last_activity_at = $2,
			updated_at = NOW()
		WHERE user_id = $1
	`
	return tx.Exec(query, userID, timestamp).Error
}
