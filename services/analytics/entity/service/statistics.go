package serviceentity

import (
	"time"

	"github.com/google/uuid"
)

type UserStatistics struct {
	UserID                  uuid.UUID
	TotalRemindersCreated   int64
	TotalRemindersCompleted int64
	TotalRemindersDeleted   int64
	ActiveReminders         int64
	CompletionRate          float64
	FirstReminderAt         *time.Time
	LastActivityAt          *time.Time
}
