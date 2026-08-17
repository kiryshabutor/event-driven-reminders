package grpc

import (
	reminderhandler "github.com/kiribu/jwt-practice/services/reminder/handler/grpc"
	"github.com/kiribu/jwt-practice/services/reminder/service"
	reminderpb "github.com/kiribu/jwt-practice/shared/proto/reminder"
	gogrpc "google.golang.org/grpc"
)

func NewServer(reminderService *service.ReminderService) *gogrpc.Server {
	server := gogrpc.NewServer()
	reminderpb.RegisterReminderServiceServer(server, reminderhandler.NewReminderServer(reminderService))
	return server
}
