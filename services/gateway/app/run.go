package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/services/gateway/config"
	client "github.com/kiribu/jwt-practice/services/gateway/gateway/grpc"
	"github.com/kiribu/jwt-practice/services/gateway/handler/http"
	customMiddleware "github.com/kiribu/jwt-practice/services/gateway/handler/http/middleware"
	"github.com/kiribu/jwt-practice/shared/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Run() {
	_ = godotenv.Load(".env")
	logger.Setup(config.GetEnv("APP_ENV", "local"))

	authClient, err := client.NewAuthClient(config.GetEnv("AUTH_SERVICE_ADDR", "auth-service:50051"))
	if err != nil {
		slog.Error("Failed to connect to Auth Service", "error", err)
		os.Exit(1)
	}
	defer authClient.Close()

	reminderClient, err := client.NewReminderClient(config.GetEnv("REMINDER_SERVICE_ADDR", "reminder-service:50052"))
	if err != nil {
		slog.Error("Failed to connect to Reminder Service", "error", err)
		os.Exit(1)
	}
	defer reminderClient.Close()

	analyticsClient, err := client.NewAnalyticsClient(config.GetEnv("ANALYTICS_SERVICE_ADDR", "analytics-service:50053"))
	if err != nil {
		slog.Error("Failed to connect to Analytics Service", "error", err)
		os.Exit(1)
	}
	defer analyticsClient.Close()

	authHandler := handlers.NewAuthHandler(authClient)
	reminderHandler := handlers.NewReminderHandler(reminderClient)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsClient)

	e := echo.New()
	e.HideBanner = true
	e.Use(customMiddleware.SlogLogger)
	e.Use(middleware.Recover())

	e.POST("/auth/register", authHandler.Register)
	e.POST("/auth/login", authHandler.Login)
	e.POST("/auth/refresh", authHandler.Refresh)

	protected := e.Group("")
	protected.Use(authHandler.AuthMiddleware)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.GET("/auth/profile", authHandler.Profile)
	protected.POST("/reminders", reminderHandler.Create)
	protected.GET("/reminders", reminderHandler.List)
	protected.GET("/reminders/:id", reminderHandler.Get)
	protected.PUT("/reminders/:id", reminderHandler.Update)
	protected.DELETE("/reminders/:id", reminderHandler.Delete)
	protected.GET("/analytics/me", analyticsHandler.GetStats)

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	port := config.GetEnv("HTTP_PORT", "8080")
	go func() {
		if err := e.Start(":" + port); err != nil {
			slog.Info("HTTP server stopped", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down API Gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = e.Shutdown(ctx)
}
