package main

import (
	"context"
	"time"

	"github.com/jiajunhuang/natrp/dial"
	"github.com/jiajunhuang/natrp/pb/serverpb"
	"go.uber.org/zap"
)

var (
	serverAddr = "127.0.0.1:10022"
	token      = "balalaxiaomoxian"
	logger, _  = zap.NewProduction()
)

func main() {
	defer logger.Sync()

	ctx := context.Background()

	client, conn, err := dial.WithServer(ctx, serverAddr, false)
	if err != nil {
		logger.Error("failed to connect to server server", zap.Error(err))
		return
	}
	defer conn.Close()

	stream, err := client.Msg(ctx)
	if err != nil {
		logger.Error("failed to communicate with server", zap.Error(err))
		return
	}

	for {
		resp := &serverpb.MsgRequest{
			Type:  serverpb.MsgType_PingPong,
			Token: token,
			Data:  []byte("ping"),
		}
		if err := stream.Send(resp); err != nil {
			logger.Error("failed to ping the server", zap.Error(err))
			return
		}
		logger.Info("ping set", zap.Any("resp", resp))
		time.Sleep(time.Second * time.Duration(5))
	}
}
