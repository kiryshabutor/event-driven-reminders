package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kiribu/jwt-practice/internal/reminder/storage"
	"github.com/kiribu/jwt-practice/models"
)

type ReminderService struct {
	storage storage.ReminderStorage
}

func NewReminderService(storage storage.ReminderStorage) *ReminderService {
	return &ReminderService{
		storage: storage,
	}
}

func (s *ReminderService) Create(ctx context.Context, userID uuid.UUID, title, description, remindAtStr string) (*models.Reminder, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	remindAt, err := time.Parse(time.RFC3339, remindAtStr)
	if err != nil {
		return nil, errors.New("invalid remind_at format, use RFC3339: 2026-01-25T10:00:00+03:00")
	}

	if remindAt.Before(time.Now()) {
		return nil, errors.New("remind_at must be in the future")
	}

	reminder, err := s.storage.Create(ctx, userID, title, description, remindAt)
	if err != nil {
		return nil, err
	}

	return reminder, nil
}

func (s *ReminderService) GetByUserID(ctx context.Context, userID uuid.UUID, status string) ([]models.Reminder, error) {
	return s.storage.GetByUserID(ctx, userID, status)
}

func (s *ReminderService) GetByID(ctx context.Context, userID, id uuid.UUID) (*models.Reminder, error) {
	return s.storage.GetByID(ctx, userID, id)
}

func (s *ReminderService) Update(ctx context.Context, userID, id uuid.UUID, title, description, remindAtStr string) (*models.Reminder, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	remindAt, err := time.Parse(time.RFC3339, remindAtStr)
	if err != nil {
		return nil, errors.New("invalid remind_at format, use RFC3339: 2026-01-25T10:00:00+03:00")
	}

	if remindAt.Before(time.Now()) {
		return nil, errors.New("remind_at must be in the future")
	}

	reminder, err := s.storage.Update(ctx, userID, id, title, description, remindAt)
	if err != nil {
		return nil, err
	}

	return reminder, nil
}

func (s *ReminderService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.storage.Delete(ctx, userID, id)
}
