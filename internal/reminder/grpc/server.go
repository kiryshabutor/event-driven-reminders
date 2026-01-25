package remindergrpc

import (
	"context"

	"github.com/kiribu/jwt-practice/internal/reminder/grpc/pb"
	"github.com/kiribu/jwt-practice/internal/reminder/service"
	"github.com/kiribu/jwt-practice/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReminderServer struct {
	pb.UnimplementedReminderServiceServer
	service *service.ReminderService
}

func NewReminderServer(svc *service.ReminderService) *ReminderServer {
	return &ReminderServer{service: svc}
}

func (s *ReminderServer) CreateReminder(ctx context.Context, req *pb.CreateReminderRequest) (*pb.ReminderResponse, error) {
	reminder, err := s.service.Create(req.UserId, req.Title, req.Description, req.RemindAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toProtoReminder(reminder), nil
}

func (s *ReminderServer) GetReminders(ctx context.Context, req *pb.GetRemindersRequest) (*pb.GetRemindersResponse, error) {
	reminders, err := s.service.GetByUserID(req.UserId, req.Status)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoReminders []*pb.ReminderResponse
	for _, r := range reminders {
		protoReminders = append(protoReminders, toProtoReminder(&r))
	}

	return &pb.GetRemindersResponse{Reminders: protoReminders}, nil
}

func (s *ReminderServer) GetReminder(ctx context.Context, req *pb.GetReminderRequest) (*pb.ReminderResponse, error) {
	reminder, err := s.service.GetByID(req.UserId, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return toProtoReminder(reminder), nil
}

func (s *ReminderServer) UpdateReminder(ctx context.Context, req *pb.UpdateReminderRequest) (*pb.ReminderResponse, error) {
	reminder, err := s.service.Update(req.UserId, req.Id, req.Title, req.Description, req.RemindAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toProtoReminder(reminder), nil
}

func (s *ReminderServer) DeleteReminder(ctx context.Context, req *pb.DeleteReminderRequest) (*pb.DeleteReminderResponse, error) {
	err := s.service.Delete(req.UserId, req.Id)
	if err != nil {
		return &pb.DeleteReminderResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.DeleteReminderResponse{
		Success: true,
		Message: "Reminder deleted successfully",
	}, nil
}

func toProtoReminder(r *models.Reminder) *pb.ReminderResponse {
	return &pb.ReminderResponse{
		Id:          r.ID,
		UserId:      r.UserID,
		Title:       r.Title,
		Description: r.Description,
		RemindAt:    r.RemindAt.Format("2006-01-02T15:04:05Z07:00"),
		IsSent:      r.IsSent,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
