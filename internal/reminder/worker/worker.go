package worker

import (
	"context"
	"log"
	"time"

	"github.com/kiribu/jwt-practice/internal/reminder/kafka"
	"github.com/kiribu/jwt-practice/internal/reminder/storage"
)

type Worker struct {
	storage  storage.ReminderStorage
	producer *kafka.Producer
	interval time.Duration
}

func NewWorker(storage storage.ReminderStorage, producer *kafka.Producer, interval time.Duration) *Worker {
	return &Worker{
		storage:  storage,
		producer: producer,
		interval: interval,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("Reminder worker started with interval %v", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping reminder worker...")
			return
		case <-ticker.C:
			w.processPending()
		}
	}
}

func (w *Worker) processPending() {
	reminders, err := w.storage.GetPending()
	if err != nil {
		log.Printf("Error fetching pending reminders: %v", err)
		return
	}

	if len(reminders) > 0 {
		log.Printf("Found %d pending reminders", len(reminders))
	}

	for _, reminder := range reminders {
		err := w.producer.SendNotification(reminder)
		if err != nil {
			log.Printf("Error sending reminder %d to Kafka: %v", reminder.ID, err)
			continue
		}

		err = w.storage.MarkAsSent(reminder.ID)
		if err != nil {
			log.Printf("Error marking reminder %d as sent: %v", reminder.ID, err)
		}
	}
}
