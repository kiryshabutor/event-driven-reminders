package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	repositoryentity "github.com/kiribu/jwt-practice/services/reminder/entity/repository"
	serviceentity "github.com/kiribu/jwt-practice/services/reminder/entity/service"
	"github.com/kiribu/jwt-practice/shared/consts"
	"github.com/kiribu/jwt-practice/shared/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

	outboxEvent := repositoryentity.OutboxEvent{
		ID:          uuid.Must(uuid.NewV7()),
		EventType:   eventType,
		AggregateID: aggregateID,
		UserID:      userID,
		Payload:     payloadJSON,
	}

	return tx.Create(&outboxEvent).Error
}

func (s *PostgresStorage) Create(ctx context.Context, userID uuid.UUID, title, description string, remindAt time.Time) (*serviceentity.Reminder, error) {
	var reminder repositoryentity.Reminder

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		reminder = repositoryentity.Reminder{
			ID:          uuid.Must(uuid.NewV7()),
			UserID:      userID,
			Title:       title,
			Description: description,
			RemindAt:    remindAt,
		}

		if err := tx.Create(&reminder).Error; err != nil {
			return fmt.Errorf("failed to insert reminder: %w", err)
		}

		event := types.LifecycleEvent{
			EventID:    uuid.Must(uuid.NewV7()),
			EventType:  consts.EventCreated,
			ReminderID: reminder.ID,
			UserID:     reminder.UserID,
			Timestamp:  time.Now(),
			Payload:    toWireReminderRow(reminder),
		}

		if err := s.createOutboxEvent(tx, consts.EventCreated, reminder.UserID, reminder.ID, event); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return toDomainReminder(reminder), nil
}

func (s *PostgresStorage) GetByUserID(ctx context.Context, userID uuid.UUID, status string) ([]serviceentity.Reminder, error) {
	var reminders []repositoryentity.Reminder
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

	result := make([]serviceentity.Reminder, len(reminders))
	for i := range reminders {
		result[i] = *toDomainReminder(reminders[i])
	}
	return result, nil
}

func (s *PostgresStorage) GetByID(ctx context.Context, userID, id uuid.UUID) (*serviceentity.Reminder, error) {
	var reminder repositoryentity.Reminder
	result := s.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&reminder)
	if result.Error != nil {
		return nil, errors.New("reminder not found")
	}
	return toDomainReminder(reminder), nil
}

func (s *PostgresStorage) Update(ctx context.Context, userID, id uuid.UUID, title, description string, remindAt time.Time) (*serviceentity.Reminder, error) {
	var reminder repositoryentity.Reminder

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

		event := types.LifecycleEvent{
			EventID:    uuid.Must(uuid.NewV7()),
			EventType:  consts.EventUpdated,
			ReminderID: reminder.ID,
			UserID:     reminder.UserID,
			Timestamp:  time.Now(),
			Payload:    toWireReminderRow(reminder),
		}

		if err := s.createOutboxEvent(tx, consts.EventUpdated, reminder.UserID, reminder.ID, event); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return toDomainReminder(reminder), nil
}

func (s *PostgresStorage) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND id = ? AND is_sent = ?", userID, id, false).Delete(&repositoryentity.Reminder{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("reminder not found or already sent")
		}

		event := types.LifecycleEvent{
			EventID:    uuid.Must(uuid.NewV7()),
			EventType:  consts.EventDeleted,
			ReminderID: id,
			UserID:     userID,
			Timestamp:  time.Now(),
			Payload:    nil,
		}

		if err := s.createOutboxEvent(tx, consts.EventDeleted, userID, id, event); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})
}

func (s *PostgresStorage) GetPending(ctx context.Context) ([]serviceentity.Reminder, error) {
	var reminders []repositoryentity.Reminder
	err := s.db.WithContext(ctx).Where("is_sent = ? AND remind_at <= ?", false, time.Now()).Find(&reminders).Error
	if err != nil {
		return nil, err
	}

	result := make([]serviceentity.Reminder, len(reminders))
	for i := range reminders {
		result[i] = *toDomainReminder(reminders[i])
	}
	return result, nil
}

func (s *PostgresStorage) MarkAsSent(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&repositoryentity.Reminder{}).Where("id = ?", id).Update("is_sent", true).Error
}

func (s *PostgresStorage) GetPendingOutboxEvents(ctx context.Context, limit int) ([]repositoryentity.OutboxEvent, error) {
	var events []repositoryentity.OutboxEvent
	err := s.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ? AND retry_count < ?", consts.OutboxPending, 5).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (s *PostgresStorage) MarkOutboxEventAsSent(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&repositoryentity.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       consts.OutboxSent,
			"processed_at": now,
		}).Error
}

func (s *PostgresStorage) IncrementOutboxRetryCount(ctx context.Context, id uuid.UUID, errMsg string) error {
	return s.db.WithContext(ctx).Model(&repositoryentity.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count":   gorm.Expr("retry_count + 1"),
			"error_message": errMsg,
			"status":        gorm.Expr("CASE WHEN retry_count + 1 >= 5 THEN ? ELSE status END", consts.OutboxFailed),
		}).Error
}

func (s *PostgresStorage) CreateNotificationEventsAndMarkSent(ctx context.Context, reminder serviceentity.Reminder) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		reminderJSON, err := json.Marshal(toWireReminder(reminder))
		if err != nil {
			return fmt.Errorf("failed to marshal reminder: %w", err)
		}

		notificationEvent := repositoryentity.OutboxEvent{
			ID:          uuid.Must(uuid.NewV7()),
			EventType:   consts.EventNotification,
			AggregateID: reminder.ID,
			UserID:      reminder.UserID,
			Payload:     reminderJSON,
		}

		if err := tx.Create(&notificationEvent).Error; err != nil {
			return fmt.Errorf("failed to create notification_trigger event: %w", err)
		}

		lifecycleEvent := types.LifecycleEvent{
			EventID:    uuid.Must(uuid.NewV7()),
			EventType:  consts.EventNotificationSent,
			ReminderID: reminder.ID,
			UserID:     reminder.UserID,
			Timestamp:  time.Now(),
			Payload:    toWireReminder(reminder),
		}
		lifecycleEventJSON, err := json.Marshal(lifecycleEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal lifecycle event: %w", err)
		}

		lifecycleOutboxEvent := repositoryentity.OutboxEvent{
			ID:          uuid.Must(uuid.NewV7()),
			EventType:   consts.EventNotificationSent,
			AggregateID: reminder.ID,
			UserID:      reminder.UserID,
			Payload:     lifecycleEventJSON,
		}

		if err := tx.Create(&lifecycleOutboxEvent).Error; err != nil {
			return fmt.Errorf("failed to create notification_sent event: %w", err)
		}

		if err := tx.Model(&repositoryentity.Reminder{}).Where("id = ?", reminder.ID).Update("is_sent", true).Error; err != nil {
			return fmt.Errorf("failed to mark reminder as sent: %w", err)
		}

		return nil
	})
}

func toDomainReminder(reminder repositoryentity.Reminder) *serviceentity.Reminder {
	return &serviceentity.Reminder{
		ID:          reminder.ID,
		UserID:      reminder.UserID,
		Title:       reminder.Title,
		Description: reminder.Description,
		RemindAt:    reminder.RemindAt,
		IsSent:      reminder.IsSent,
		CreatedAt:   reminder.CreatedAt,
		UpdatedAt:   reminder.UpdatedAt,
	}
}

func toWireReminder(reminder serviceentity.Reminder) types.Reminder {
	return types.Reminder{
		ID:          reminder.ID,
		UserID:      reminder.UserID,
		Title:       reminder.Title,
		Description: reminder.Description,
		RemindAt:    reminder.RemindAt,
		IsSent:      reminder.IsSent,
		CreatedAt:   reminder.CreatedAt,
		UpdatedAt:   reminder.UpdatedAt,
	}
}

func toWireReminderRow(reminder repositoryentity.Reminder) types.Reminder {
	return toWireReminder(serviceentity.Reminder{
		ID:          reminder.ID,
		UserID:      reminder.UserID,
		Title:       reminder.Title,
		Description: reminder.Description,
		RemindAt:    reminder.RemindAt,
		IsSent:      reminder.IsSent,
		CreatedAt:   reminder.CreatedAt,
		UpdatedAt:   reminder.UpdatedAt,
	})
}
