package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jiajunhuang/natrp/pb/serverpb"
	reuse "github.com/libp2p/go-reuseport"
	"go.uber.org/zap"
)

var (
	wanIP = "127.0.0.1"
)

type service struct{}

func (s *service) Register(ctx context.Context, req *serverpb.RegisterRequest) (*serverpb.RegisterResponse, error) {
	return nil, nil
}

func (s *service) Login(ctx context.Context, req *serverpb.LoginRequest) (*serverpb.LoginResponse, error) {
	return nil, nil
}

func (s *service) Msg(stream serverpb.ServerService_MsgServer) error {
	listener, err := reuse.Listen("tcp", "0.0.0.0:10033")
	if err != nil {
		logger.Error("failed to listen", zap.Error(err))
		return err
	}
	defer listener.Close()
	addrList := strings.Split(listener.Addr().String(), ":")
	addr := fmt.Sprintf("%s:%s", wanIP, addrList[len(addrList)-1])
	logger.Info("server listen at", zap.String("addr", addr))

	conn, err := listener.Accept()
	if err != nil {
		logger.Error("failed to accept", zap.Error(err))
		return err
	}
	defer conn.Close()

	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				logger.Error("failed to read", zap.Error(err))
				return
			}

			if _, err := conn.Write(req.Data); err != nil {
				logger.Error("failed to write", zap.Error(err))
				return
			}
		}
	}()

	data := make([]byte, 1024)
	for {
		n, err := conn.Read(data)
		if err != nil {
			logger.Error("failed to read", zap.Error(err))
			return err
		}

		if err := stream.Send(&serverpb.MsgResponse{Data: data[:n]}); err != nil {
			logger.Error("failed to write", zap.Error(err))
			return err
		}
	}
}
