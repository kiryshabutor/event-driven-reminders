package repositoryentity

import (
	"time"

	"github.com/google/uuid"
)

type Reminder struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	RemindAt    time.Time `gorm:"type:timestamptz;not null"`
	IsSent      bool      `gorm:"default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (Reminder) TableName() string {
	return "reminders"
}
