package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	EventType    string          `gorm:"type:varchar(50);not null" json:"event_type"`
	AggregateID  uuid.UUID       `gorm:"type:uuid" json:"aggregate_id"`
	UserID       uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`
	Payload      json.RawMessage `gorm:"type:jsonb;not null" json:"payload"`
	Status       string          `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	RetryCount   int             `gorm:"default:0" json:"retry_count"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
	ProcessedAt  *time.Time      `json:"processed_at"`
	ErrorMessage *string         `gorm:"type:text" json:"error_message"`
}

func (OutboxEvent) TableName() string {
	return "reminders_outbox"
}
