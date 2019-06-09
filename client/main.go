package main

import (
	"context"
	"net"
	"time"

	"github.com/jiajunhuang/natrp/dial"
	"github.com/jiajunhuang/natrp/pb/serverpb"
	"go.uber.org/zap"
)

var (
	serverAddr = "127.0.0.1:10022"
	token      = "balalaxiaomoxian"
	logger, _  = zap.NewProduction()
	remoteAddr = ""
)

func main() {
	defer logger.Sync()

	for {
		func() {
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

			localConn, err := net.Dial("tcp", "127.0.0.1:22")
			if err != nil {
				logger.Error("failed to communicate with local port", zap.Error(err))
				return
			}
			defer localConn.Close()

			go func() {
				defer localConn.Close()

				for {
					req, err := stream.Recv()
					if err != nil {
						logger.Error("failed to read", zap.Error(err))
						return
					}

					if _, err := localConn.Write(req.Data); err != nil {
						logger.Error("failed to write", zap.Error(err))
						return
					}
				}
			}()

			data := make([]byte, 1024)
			for {
				defer localConn.Close()

				n, err := localConn.Read(data)
				if err != nil {
					logger.Error("failed to read", zap.Error(err))
					return
				}

				if err := stream.Send(&serverpb.MsgRequest{Data: data[:n]}); err != nil {
					logger.Error("failed to write", zap.Error(err))
					return
				}
			}
		}()

		time.Sleep(time.Second * time.Duration(1))
		logger.Info("trying to reconnect...")
	}
}
