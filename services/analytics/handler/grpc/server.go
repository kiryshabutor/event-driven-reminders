package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	serviceentity "github.com/kiribu/jwt-practice/services/analytics/entity/service"
	"github.com/kiribu/jwt-practice/services/analytics/service"
	analyticspb "github.com/kiribu/jwt-practice/shared/proto/analytics"
)

type AnalyticsServer struct {
	analyticspb.UnimplementedAnalyticsServiceServer
	service *service.AnalyticsService
}

func NewAnalyticsServer(service *service.AnalyticsService) *AnalyticsServer {
	return &AnalyticsServer{service: service}
}

func (s *AnalyticsServer) GetUserStats(ctx context.Context, req *analyticspb.GetUserStatsRequest) (*analyticspb.UserStatsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	stats, err := s.service.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	return convertToProto(stats), nil
}

func convertToProto(s *serviceentity.UserStatistics) *analyticspb.UserStatsResponse {
	resp := &analyticspb.UserStatsResponse{
		UserId:                  s.UserID.String(),
		TotalRemindersCreated:   s.TotalRemindersCreated,
		TotalRemindersCompleted: s.TotalRemindersCompleted,
		TotalRemindersDeleted:   s.TotalRemindersDeleted,
		ActiveReminders:         s.ActiveReminders,
		CompletionRate:          s.CompletionRate,
	}
	if s.FirstReminderAt != nil {
		resp.FirstReminderAt = s.FirstReminderAt.Format(time.RFC3339)
	}
	if s.LastActivityAt != nil {
		resp.LastActivityAt = s.LastActivityAt.Format(time.RFC3339)
	}
	return resp
}
