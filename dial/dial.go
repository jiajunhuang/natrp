package dial

import (
	"context"

	"github.com/jiajunhuang/natrp/errors"
	"github.com/jiajunhuang/natrp/pb/serverpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	logger, _ = zap.NewProduction()
)

// WithServer dial with server
func WithServer(ctx context.Context, addr string, tls bool) (serverpb.ServerServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		logger.Error("failed to connect to server server", zap.Error(err))
		return nil, nil, err
	}

	select {
	case <-ctx.Done():
		logger.Error("ctx had been done")
		return nil, nil, errors.ErrCanceled
	default:
	}

	client := serverpb.NewServerServiceClient(conn)
	return client, conn, nil
}
