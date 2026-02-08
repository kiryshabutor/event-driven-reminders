package models

import (
	"time"

	"github.com/google/uuid"
)

type Reminder struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	RemindAt    time.Time `gorm:"type:timestamptz;not null" json:"remind_at"`
	IsSent      bool      `gorm:"default:false" json:"is_sent"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Reminder) TableName() string {
	return "reminders"
}
