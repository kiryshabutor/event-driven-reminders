package grpc

import (
	"context"
	"log/slog"
	"time"

	appgrpc "github.com/kiribu/jwt-practice/services/gateway/app/grpc"
	reminderpb "github.com/kiribu/jwt-practice/shared/proto/reminder"
	gogrpc "google.golang.org/grpc"
)

type ReminderClient struct {
	conn   *gogrpc.ClientConn
	client reminderpb.ReminderServiceClient
}

func NewReminderClient(addr string) (*ReminderClient, error) {
	slog.Info("Connecting to Reminder Service", "addr", addr)
	conn, err := appgrpc.Dial(addr, 30*time.Second)
	if err != nil {
		return nil, err
	}
	slog.Info("Connected to Reminder Service", "addr", addr)

	return &ReminderClient{
		conn:   conn,
		client: reminderpb.NewReminderServiceClient(conn),
	}, nil
}

func (c *ReminderClient) Close() error {
	return c.conn.Close()
}

func (c *ReminderClient) Create(ctx context.Context, userID string, title, description, remindAt string) (*reminderpb.ReminderResponse, error) {
	return c.client.CreateReminder(ctx, &reminderpb.CreateReminderRequest{
		UserId:      userID,
		Title:       title,
		Description: description,
		RemindAt:    remindAt,
	})
}

func (c *ReminderClient) GetAll(ctx context.Context, userID string, status string) (*reminderpb.GetRemindersResponse, error) {
	return c.client.GetReminders(ctx, &reminderpb.GetRemindersRequest{
		UserId: userID,
		Status: status,
	})
}

func (c *ReminderClient) GetByID(ctx context.Context, userID, id string) (*reminderpb.ReminderResponse, error) {
	return c.client.GetReminder(ctx, &reminderpb.GetReminderRequest{
		UserId: userID,
		Id:     id,
	})
}

func (c *ReminderClient) Update(ctx context.Context, userID, id string, title, description, remindAt string) (*reminderpb.ReminderResponse, error) {
	return c.client.UpdateReminder(ctx, &reminderpb.UpdateReminderRequest{
		UserId:      userID,
		Id:          id,
		Title:       title,
		Description: description,
		RemindAt:    remindAt,
	})
}

func (c *ReminderClient) Delete(ctx context.Context, userID, id string) (*reminderpb.DeleteReminderResponse, error) {
	return c.client.DeleteReminder(ctx, &reminderpb.DeleteReminderRequest{
		UserId: userID,
		Id:     id,
	})
}
