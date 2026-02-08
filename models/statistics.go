package models

import (
	"time"

	"github.com/google/uuid"
)

type UserStatistics struct {
	UserID                  uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	TotalRemindersCreated   int64      `gorm:"default:0" json:"total_reminders_created"`
	TotalRemindersCompleted int64      `gorm:"default:0" json:"total_reminders_completed"`
	TotalRemindersDeleted   int64      `gorm:"default:0" json:"total_reminders_deleted"`
	ActiveReminders         int64      `gorm:"default:0" json:"active_reminders"`
	CompletionRate          float64    `gorm:"type:decimal(5,2);default:0" json:"completion_rate"`
	FirstReminderAt         *time.Time `json:"first_reminder_at"`
	LastActivityAt          *time.Time `json:"last_activity_at"`
	CreatedAt               time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserStatistics) TableName() string {
	return "analytics.user_statistics"
}

type ProcessedEvent struct {
	EventID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProcessedAt time.Time `gorm:"autoCreateTime"`
}

func (ProcessedEvent) TableName() string {
	return "analytics.processed_events"
}
