package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kiribu/jwt-practice/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReminderStorage interface {
	Create(ctx context.Context, userID uuid.UUID, title, description string, remindAt time.Time) (*models.Reminder, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, status string) ([]models.Reminder, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*models.Reminder, error)
	Update(ctx context.Context, userID, id uuid.UUID, title, description string, remindAt time.Time) (*models.Reminder, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	GetPending(ctx context.Context) ([]models.Reminder, error)
	MarkAsSent(ctx context.Context, id uuid.UUID) error
	// Outbox methods
	GetPendingOutboxEvents(ctx context.Context, limit int) ([]models.OutboxEvent, error)
	MarkOutboxEventAsSent(ctx context.Context, id uuid.UUID) error
	IncrementOutboxRetryCount(ctx context.Context, id uuid.UUID, errMsg string) error
	CreateNotificationEventsAndMarkSent(ctx context.Context, reminder models.Reminder) error
}

type PostgresStorage struct {
	db *gorm.DB
}

func NewPostgresStorage(db *gorm.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) createOutboxEvent(tx *gorm.DB, eventType string, userID, aggregateID uuid.UUID, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal outbox payload: %w", err)
	}

	outboxEvent := models.OutboxEvent{
		ID:          uuid.Must(uuid.NewV7()),
		EventType:   eventType,
		AggregateID: aggregateID,
		UserID:      userID,
		Payload:     payloadJSON,
	}

	return tx.Create(&outboxEvent).Error
}

func (s *PostgresStorage) Create(ctx context.Context, userID uuid.UUID, title, description string, remindAt time.Time) (*models.Reminder, error) {
	var reminder models.Reminder

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		reminder = models.Reminder{
			ID:          uuid.Must(uuid.NewV7()),
			UserID:      userID,
			Title:       title,
			Description: description,
			RemindAt:    remindAt,
		}

		if err := tx.Create(&reminder).Error; err != nil {
			return fmt.Errorf("failed to insert reminder: %w", err)
		}

		event := models.LifecycleEvent{
			EventID:    uuid.Must(uuid.NewV7()),
			EventType:  "created",
			ReminderID: reminder.ID,
			UserID:     reminder.UserID,
			Timestamp:  time.Now(),
			Payload:    reminder,
		}

		if err := s.createOutboxEvent(tx, "created", reminder.UserID, reminder.ID, event); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &reminder, nil
}

func (s *PostgresStorage) GetByUserID(ctx context.Context, userID uuid.UUID, status string) ([]models.Reminder, error) {
	var reminders []models.Reminder
	query := s.db.WithContext(ctx).Where("user_id = ?", userID)

	switch status {
	case "pending":
		query = query.Where("is_sent = ?", false).Order("remind_at ASC")
	case "sent":
		query = query.Where("is_sent = ?", true).Order("remind_at DESC")
	default:
		query = query.Order("remind_at ASC")
	}

	if err := query.Find(&reminders).Error; err != nil {
		return nil, err
	}

	return reminders, nil
}

func (s *PostgresStorage) GetByID(ctx context.Context, userID, id uuid.UUID) (*models.Reminder, error) {
	var reminder models.Reminder
	result := s.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&reminder)
	if result.Error != nil {
		return nil, errors.New("reminder not found")
	}
	return &reminder, nil
}

func (s *PostgresStorage) Update(ctx context.Context, userID, id uuid.UUID, title, description string, remindAt time.Time) (*models.Reminder, error) {
	var reminder models.Reminder

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND id = ? AND is_sent = ?", userID, id, false).First(&reminder)
		if result.Error != nil {
			return errors.New("reminder not found or already sent")
		}

		reminder.Title = title
		reminder.Description = description
		reminder.RemindAt = remindAt

		if err := tx.Save(&reminder).Error; err != nil {
			return fmt.Errorf("failed to update reminder: %w", err)
		}

		event := models.LifecycleEvent{
			EventID:    uuid.Must(uuid.NewV7()),
			EventType:  "updated",
			ReminderID: reminder.ID,
			UserID:     reminder.UserID,
			Timestamp:  time.Now(),
			Payload:    reminder,
		}

		if err := s.createOutboxEvent(tx, "updated", reminder.UserID, reminder.ID, event); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &reminder, nil
}

func (s *PostgresStorage) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND id = ? AND is_sent = ?", userID, id, false).Delete(&models.Reminder{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("reminder not found or already sent")
		}

		event := models.LifecycleEvent{
			EventID:    uuid.Must(uuid.NewV7()),
			EventType:  "deleted",
			ReminderID: id,
			UserID:     userID,
			Timestamp:  time.Now(),
			Payload:    nil,
		}

		if err := s.createOutboxEvent(tx, "deleted", userID, id, event); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})
}

func (s *PostgresStorage) GetPending(ctx context.Context) ([]models.Reminder, error) {
	var reminders []models.Reminder
	err := s.db.WithContext(ctx).Where("is_sent = ? AND remind_at <= ?", false, time.Now()).Find(&reminders).Error
	if err != nil {
		return nil, err
	}
	return reminders, nil
}

func (s *PostgresStorage) MarkAsSent(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&models.Reminder{}).Where("id = ?", id).Update("is_sent", true).Error
}

func (s *PostgresStorage) GetPendingOutboxEvents(ctx context.Context, limit int) ([]models.OutboxEvent, error) {
	var events []models.OutboxEvent
	err := s.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ? AND retry_count < ?", "PENDING", 5).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (s *PostgresStorage) MarkOutboxEventAsSent(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "SENT",
			"processed_at": now,
		}).Error
}

func (s *PostgresStorage) IncrementOutboxRetryCount(ctx context.Context, id uuid.UUID, errMsg string) error {
	return s.db.WithContext(ctx).Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count":   gorm.Expr("retry_count + 1"),
			"error_message": errMsg,
			"status":        gorm.Expr("CASE WHEN retry_count + 1 >= 5 THEN 'FAILED' ELSE status END"),
		}).Error
}

func (s *PostgresStorage) CreateNotificationEventsAndMarkSent(ctx context.Context, reminder models.Reminder) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// notification_trigger: raw Reminder for notification-service
		reminderJSON, err := json.Marshal(reminder)
		if err != nil {
			return fmt.Errorf("failed to marshal reminder: %w", err)
		}

		notificationEvent := models.OutboxEvent{
			ID:          uuid.Must(uuid.NewV7()),
			EventType:   "notification_trigger",
			AggregateID: reminder.ID,
			UserID:      reminder.UserID,
			Payload:     reminderJSON,
		}

		if err := tx.Create(&notificationEvent).Error; err != nil {
			return fmt.Errorf("failed to create notification_trigger event: %w", err)
		}

		// notification_sent: LifecycleEvent for analytics-service
		lifecycleEvent := models.LifecycleEvent{
			EventID:    uuid.Must(uuid.NewV7()),
			EventType:  "notification_sent",
			ReminderID: reminder.ID,
			UserID:     reminder.UserID,
			Timestamp:  time.Now(),
			Payload:    reminder,
		}
		lifecycleEventJSON, err := json.Marshal(lifecycleEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal lifecycle event: %w", err)
		}

		lifecycleOutboxEvent := models.OutboxEvent{
			ID:          uuid.Must(uuid.NewV7()),
			EventType:   "notification_sent",
			AggregateID: reminder.ID,
			UserID:      reminder.UserID,
			Payload:     lifecycleEventJSON,
		}

		if err := tx.Create(&lifecycleOutboxEvent).Error; err != nil {
			return fmt.Errorf("failed to create notification_sent event: %w", err)
		}

		// Mark reminder as sent
		if err := tx.Model(&models.Reminder{}).Where("id = ?", reminder.ID).Update("is_sent", true).Error; err != nil {
			return fmt.Errorf("failed to mark reminder as sent: %w", err)
		}

		return nil
	})
}
