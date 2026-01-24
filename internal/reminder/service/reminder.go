package service

import (
	"errors"
	"time"

	"github.com/kiribu/jwt-practice/internal/reminder/storage"
	"github.com/kiribu/jwt-practice/models"
)

type ReminderService struct {
	storage storage.ReminderStorage
}

func NewReminderService(storage storage.ReminderStorage) *ReminderService {
	return &ReminderService{storage: storage}
}

func (s *ReminderService) Create(userID int64, title, description, remindAtStr string) (*models.Reminder, error) {
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

	return s.storage.Create(userID, title, description, remindAt)
}

func (s *ReminderService) GetByUserID(userID int64) ([]models.Reminder, error) {
	return s.storage.GetByUserID(userID)
}

func (s *ReminderService) GetByID(userID, id int64) (*models.Reminder, error) {
	return s.storage.GetByID(userID, id)
}

func (s *ReminderService) Update(userID, id int64, title, description, remindAtStr string) (*models.Reminder, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	remindAt, err := time.Parse(time.RFC3339, remindAtStr)
	if err != nil {
		return nil, errors.New("invalid remind_at format, use RFC3339: 2026-01-25T10:00:00+03:00")
	}

	return s.storage.Update(userID, id, title, description, remindAt)
}

func (s *ReminderService) Delete(userID, id int64) error {
	return s.storage.Delete(userID, id)
}
