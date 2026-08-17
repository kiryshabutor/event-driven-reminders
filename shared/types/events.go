package types

import (
	"time"

	"github.com/google/uuid"
)

type LifecycleEvent struct {
	EventID    uuid.UUID   `json:"event_id"`   // Unique ID for idempotency
	EventType  string      `json:"event_type"` // "created", "updated", "deleted", "notification_sent"
	ReminderID uuid.UUID   `json:"reminder_id"`
	UserID     uuid.UUID   `json:"user_id"`
	Timestamp  time.Time   `json:"timestamp"`
	Payload    interface{} `json:"payload,omitempty"` // Reminder snapshot or nil
}

// Reminder is the stable JSON shape exchanged through the notification topic.
// It intentionally lives outside reminder's domain package.
type Reminder struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	RemindAt    time.Time `json:"remind_at"`
	IsSent      bool      `json:"is_sent"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
