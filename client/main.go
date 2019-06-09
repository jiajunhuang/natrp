package main

import (
	"context"
	"flag"
	"net"
	"time"

	"github.com/jiajunhuang/natrp/dial"
	"github.com/jiajunhuang/natrp/pb/serverpb"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	logger, _ = zap.NewProduction()

	localAddr  = flag.String("local", "127.0.0.1:80", "-local=<你本地需要转发的地址>")
	serverAddr = flag.String("server", "natrp.jiajunhuang.com:10022", "-server=<你的服务器地址>")
	token      = flag.String("token", "balalaxiaomoxian", "-token=<你的token>")
)

func main() {
	defer logger.Sync()

	flag.Parse()
	retryCount := 0

	for {
		func() {
			md := metadata.Pairs("natrp-token", *token)
			ctx := metadata.NewOutgoingContext(context.Background(), md)

			client, conn, err := dial.WithServer(ctx, *serverAddr, false)
			if err != nil {
				logger.Error("failed to connect to server server", zap.Error(err))
				return
			}
			defer conn.Close()

			logger.Info("try to connect to server", zap.String("addr", *serverAddr))

			stream, err := client.Msg(ctx)
			if err != nil {
				logger.Error("failed to communicate with server", zap.Error(err))
				return
			}

			logger.Info("success to connect to server", zap.String("addr", *serverAddr))
			retryCount = 0

			localConn, err := net.Dial("tcp", *localAddr)
			if err != nil {
				logger.Error("failed to communicate with local port", zap.Error(err))
				return
			}
			defer localConn.Close()

			logger.Info("start to build a brige between local and server", zap.String("server", *serverAddr), zap.String("local", *localAddr))

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

		if retryCount < 10 {
			time.Sleep(time.Microsecond * time.Duration(100*retryCount))
		} else if retryCount < 60 {
			time.Sleep(time.Second * time.Duration(1))
		} else if retryCount > 60 {
			time.Sleep(time.Second * time.Duration(10))
		}
		logger.Info("trying to reconnect", zap.String("addr", *serverAddr))
		retryCount++
	}
}
