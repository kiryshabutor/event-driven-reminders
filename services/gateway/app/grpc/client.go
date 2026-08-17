package grpc

import (
	"context"
	"time"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Dial(addr string, timeout time.Duration) (*gogrpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return gogrpc.DialContext(ctx, addr,
		gogrpc.WithTransportCredentials(insecure.NewCredentials()),
		gogrpc.WithBlock(),
	)
}
