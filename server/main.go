package main

import (
	"net"

	"github.com/jiajunhuang/natrp/pb/serverpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	addr      = "0.0.0.0:10022"
	logger, _ = zap.NewProduction()
)

func main() {
	defer logger.Sync()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("failed to listen", zap.String("addr", addr))
	}

	// register service
	svc := service{}
	server := grpc.NewServer()

	serverpb.RegisterServerServiceServer(server, &svc)
	logger.Info("server start to listen", zap.String("addr", addr))
	if err := server.Serve(listener); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}
}
