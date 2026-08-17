package app

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	dbpg "github.com/kiribu/jwt-practice/services/reminder/app/db/pg"
	grpcapp "github.com/kiribu/jwt-practice/services/reminder/app/grpc"
	"github.com/kiribu/jwt-practice/services/reminder/config"
	"github.com/kiribu/jwt-practice/services/reminder/gateway/kafka"
	pgrepository "github.com/kiribu/jwt-practice/services/reminder/repository/pg"
	"github.com/kiribu/jwt-practice/services/reminder/service"
	"github.com/kiribu/jwt-practice/services/reminder/worker"
	"github.com/kiribu/jwt-practice/shared/logger"
)

func Run() {
	_ = godotenv.Load(".env")
	logger.Setup(config.GetEnv("APP_ENV", "local"))

	db, err := dbpg.Connect(config.LoadDatabaseConfig())
	if err != nil {
		slog.Error("DB connection error", "error", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	store := pgrepository.NewPostgresStorage(db)
	brokers := strings.Split(config.GetEnv("KAFKA_BROKERS", "kafka:9092"), ",")
	notificationProducer := kafka.NewProducer(brokers, config.GetEnv("KAFKA_TOPIC_NOTIFICATIONS", "notifications"))
	defer notificationProducer.Close()
	lifecycleProducer := kafka.NewProducer(brokers, config.GetEnv("KAFKA_TOPIC_LIFECYCLE", "reminder_lifecycle"))
	defer lifecycleProducer.Close()

	reminderService := service.NewReminderService(store)
	grpcServer := grpcapp.NewServer(reminderService)

	port := config.GetEnv("REMINDER_GRPC_PORT", "50052")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("Failed to start listener", "error", err)
		os.Exit(1)
	}

	interval, err := time.ParseDuration(config.GetEnv("WORKER_INTERVAL", "5s"))
	if err != nil {
		slog.Error("Invalid WORKER_INTERVAL", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.NewNotificationWorker(store, interval).Start(ctx)
	go worker.NewOutboxWorker(store, lifecycleProducer, notificationProducer, 500*time.Millisecond).Start(ctx)

	slog.Info("Reminder Service (gRPC) started", "port", port)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			slog.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down Reminder Service...")
	cancel()
	grpcServer.GracefulStop()
}
