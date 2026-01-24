package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/config"
	remindergrpc "github.com/kiribu/jwt-practice/internal/reminder/grpc"
	"github.com/kiribu/jwt-practice/internal/reminder/grpc/pb"
	"github.com/kiribu/jwt-practice/internal/reminder/service"
	"github.com/kiribu/jwt-practice/internal/reminder/storage"
	"google.golang.org/grpc"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using system environment variables")
	}
}

func main() {
	dbConfig := config.LoadDatabaseConfig()

	db, err := config.ConnectDatabase(dbConfig)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	log.Println("Reminder Service: Successfully connected to PostgreSQL")

	store := storage.NewPostgresStorage(db)
	reminderService := service.NewReminderService(store)
	reminderServer := remindergrpc.NewReminderServer(reminderService)

	grpcServer := grpc.NewServer()
	pb.RegisterReminderServiceServer(grpcServer, reminderServer)

	port := getEnv("REMINDER_GRPC_PORT", "50052")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}

	log.Printf("Reminder Service (gRPC) started on port %s\n", port)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Reminder Service...")
	grpcServer.GracefulStop()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
