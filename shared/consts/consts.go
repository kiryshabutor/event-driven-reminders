package consts

const (
	EventCreated          = "created"
	EventUpdated          = "updated"
	EventDeleted          = "deleted"
	EventNotificationSent = "notification_sent"
	EventNotification     = "notification_trigger"

	OutboxPending = "PENDING"
	OutboxSent    = "SENT"
	OutboxFailed  = "FAILED"
)
