package grpc

import (
	authhandler "github.com/kiribu/jwt-practice/services/auth/handler/grpc"
	"github.com/kiribu/jwt-practice/services/auth/service"
	authpb "github.com/kiribu/jwt-practice/shared/proto/auth"
	gogrpc "google.golang.org/grpc"
)

func NewServer(authService *service.AuthService) *gogrpc.Server {
	server := gogrpc.NewServer()
	authpb.RegisterAuthServiceServer(server, authhandler.NewAuthServer(authService))
	return server
}
