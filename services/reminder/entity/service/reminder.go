package serviceentity

import (
	"time"

	"github.com/google/uuid"
)

type Reminder struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	RemindAt    time.Time
	IsSent      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
