package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/kiribu/jwt-practice/services/gateway/gateway/grpc"
	"github.com/labstack/echo/v4"
)

type AnalyticsHandler struct {
	analyticsClient *grpc.AnalyticsClient
}

func NewAnalyticsHandler(analyticsClient *grpc.AnalyticsClient) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsClient: analyticsClient}
}

func (h *AnalyticsHandler) GetStats(c echo.Context) error {
	userID := c.Get("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	resp, err := h.analyticsClient.GetUserStats(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
