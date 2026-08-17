package app

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	dbpg "github.com/kiribu/jwt-practice/services/auth/app/db/pg"
	grpcapp "github.com/kiribu/jwt-practice/services/auth/app/grpc"
	"github.com/kiribu/jwt-practice/services/auth/app/redis"
	"github.com/kiribu/jwt-practice/services/auth/config"
	pgrepository "github.com/kiribu/jwt-practice/services/auth/repository/pg"
	"github.com/kiribu/jwt-practice/services/auth/service"
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

	redisClient, err := redis.NewRedisClient(
		config.GetEnv("REDIS_ADDR", "localhost:6379"),
		config.GetEnv("REDIS_PASSWORD", ""),
	)
	if err != nil {
		slog.Error("Redis connection error", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	store := pgrepository.NewPostgresStorage(db)
	authService := service.NewAuthService(store, redisClient)
	grpcServer := grpcapp.NewServer(authService)

	port := config.GetEnv("GRPC_PORT", "50051")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("Failed to start listener", "error", err)
		os.Exit(1)
	}

	slog.Info("Auth Service (gRPC) started", "port", port)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			slog.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down Auth Service...")
	grpcServer.GracefulStop()
}
