package grpc

import (
	"context"

	"github.com/google/uuid"
	serviceentity "github.com/kiribu/jwt-practice/services/reminder/entity/service"
	"github.com/kiribu/jwt-practice/services/reminder/service"
	reminderpb "github.com/kiribu/jwt-practice/shared/proto/reminder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReminderServer struct {
	reminderpb.UnimplementedReminderServiceServer
	service *service.ReminderService
}

func NewReminderServer(svc *service.ReminderService) *ReminderServer {
	return &ReminderServer{service: svc}
}

func (s *ReminderServer) CreateReminder(ctx context.Context, req *reminderpb.CreateReminderRequest) (*reminderpb.ReminderResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	reminder, err := s.service.Create(ctx, userID, req.Title, req.Description, req.RemindAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toProtoReminder(reminder), nil
}

func (s *ReminderServer) GetReminders(ctx context.Context, req *reminderpb.GetRemindersRequest) (*reminderpb.GetRemindersResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	reminders, err := s.service.GetByUserID(ctx, userID, req.Status)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoReminders []*reminderpb.ReminderResponse
	for _, r := range reminders {
		protoReminders = append(protoReminders, toProtoReminder(&r))
	}

	return &reminderpb.GetRemindersResponse{Reminders: protoReminders}, nil
}

func (s *ReminderServer) GetReminder(ctx context.Context, req *reminderpb.GetReminderRequest) (*reminderpb.ReminderResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	reminder, err := s.service.GetByID(ctx, userID, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return toProtoReminder(reminder), nil
}

func (s *ReminderServer) UpdateReminder(ctx context.Context, req *reminderpb.UpdateReminderRequest) (*reminderpb.ReminderResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	reminder, err := s.service.Update(ctx, userID, id, req.Title, req.Description, req.RemindAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toProtoReminder(reminder), nil
}

func (s *ReminderServer) DeleteReminder(ctx context.Context, req *reminderpb.DeleteReminderRequest) (*reminderpb.DeleteReminderResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return &reminderpb.DeleteReminderResponse{
			Success: false,
			Message: "invalid user_id: " + err.Error(),
		}, nil
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return &reminderpb.DeleteReminderResponse{
			Success: false,
			Message: "invalid id: " + err.Error(),
		}, nil
	}

	err = s.service.Delete(ctx, userID, id)
	if err != nil {
		return &reminderpb.DeleteReminderResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &reminderpb.DeleteReminderResponse{
		Success: true,
		Message: "Reminder deleted successfully",
	}, nil
}

func toProtoReminder(r *serviceentity.Reminder) *reminderpb.ReminderResponse {
	return &reminderpb.ReminderResponse{
		Id:          r.ID.String(),     // UUID to string
		UserId:      r.UserID.String(), // UUID to string
		Title:       r.Title,
		Description: r.Description,
		RemindAt:    r.RemindAt.Format("2006-01-02T15:04:05Z07:00"),
		IsSent:      r.IsSent,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
