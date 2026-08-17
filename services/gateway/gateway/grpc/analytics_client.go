package grpc

import (
	"context"
	"log/slog"
	"time"

	appgrpc "github.com/kiribu/jwt-practice/services/gateway/app/grpc"
	analyticspb "github.com/kiribu/jwt-practice/shared/proto/analytics"
	gogrpc "google.golang.org/grpc"
)

type AnalyticsClient struct {
	conn   *gogrpc.ClientConn
	client analyticspb.AnalyticsServiceClient
}

func NewAnalyticsClient(addr string) (*AnalyticsClient, error) {
	slog.Info("Connecting to Analytics Service", "addr", addr)
	conn, err := appgrpc.Dial(addr, 30*time.Second)
	if err != nil {
		return nil, err
	}
	slog.Info("Connected to Analytics Service", "addr", addr)

	return &AnalyticsClient{conn: conn,
		client: analyticspb.NewAnalyticsServiceClient(conn),
	}, nil
}

func (c *AnalyticsClient) Close() error {
	return c.conn.Close()
}

func (c *AnalyticsClient) GetUserStats(ctx context.Context, userID string) (*analyticspb.UserStatsResponse, error) {
	req := &analyticspb.GetUserStatsRequest{
		UserId: userID,
	}
	return c.client.GetUserStats(ctx, req)
}
