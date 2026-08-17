package grpc

import (
	analyticshandler "github.com/kiribu/jwt-practice/services/analytics/handler/grpc"
	"github.com/kiribu/jwt-practice/services/analytics/service"
	analyticspb "github.com/kiribu/jwt-practice/shared/proto/analytics"
	gogrpc "google.golang.org/grpc"
)

func NewServer(analyticsService *service.AnalyticsService) *gogrpc.Server {
	server := gogrpc.NewServer()
	analyticspb.RegisterAnalyticsServiceServer(server, analyticshandler.NewAnalyticsServer(analyticsService))
	return server
}
