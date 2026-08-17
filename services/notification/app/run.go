package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/services/notification/config"
	"github.com/kiribu/jwt-practice/services/notification/handler/kafka"
	"github.com/kiribu/jwt-practice/shared/logger"
)

func Run() {
	_ = godotenv.Load(".env")
	logger.Setup(config.GetEnv("APP_ENV", "local"))

	brokers := strings.Split(config.GetEnv("KAFKA_BROKERS", "kafka:9092"), ",")
	consumer := kafka.NewConsumer(
		brokers,
		config.GetEnv("KAFKA_TOPIC", "notifications"),
		config.GetEnv("KAFKA_GROUP_ID", "notification-workers"),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go consumer.Start(ctx)

	slog.Info("Notification Service started")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down Notification Service...")
	cancel()
}
