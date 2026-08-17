package app

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	dbpg "github.com/kiribu/jwt-practice/services/analytics/app/db/pg"
	grpcapp "github.com/kiribu/jwt-practice/services/analytics/app/grpc"
	"github.com/kiribu/jwt-practice/services/analytics/config"
	"github.com/kiribu/jwt-practice/services/analytics/handler/kafka"
	pgrepository "github.com/kiribu/jwt-practice/services/analytics/repository/pg"
	"github.com/kiribu/jwt-practice/services/analytics/service"
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
	analyticsService := service.NewAnalyticsService(store)

	brokers := strings.Split(config.GetEnv("KAFKA_BROKERS", "kafka:9092"), ",")
	consumer := kafka.NewConsumer(brokers, config.GetEnv("KAFKA_TOPIC_LIFECYCLE", "reminder_lifecycle"), analyticsService)
	go consumer.Start()
	defer consumer.Close()

	port := config.GetEnv("ANALYTICS_GRPC_PORT", "50053")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("Failed to start listener", "error", err)
		os.Exit(1)
	}

	grpcServer := grpcapp.NewServer(analyticsService)
	slog.Info("Analytics Service (gRPC) started", "port", port)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			slog.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down Analytics Service...")
	grpcServer.GracefulStop()
}
