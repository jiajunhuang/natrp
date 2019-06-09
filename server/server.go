package main

import (
	"context"

	"github.com/jiajunhuang/natrp/pb/serverpb"
	"go.uber.org/zap"
)

type service struct{}

func (s *service) Register(ctx context.Context, req *serverpb.RegisterRequest) (*serverpb.RegisterResponse, error) {
	return nil, nil
}

func (s *service) Login(ctx context.Context, req *serverpb.LoginRequest) (*serverpb.LoginResponse, error) {
	return nil, nil
}

func (s *service) Msg(stream serverpb.ServerService_MsgServer) error {
	for {
		req, err := stream.Recv()
		logger.Info("request from client", zap.Any("req", req), zap.Error(err))

		resp := &serverpb.MsgResponse{
			Code: serverpb.Code_CodeSucceed,
			Type: serverpb.MsgType_PingPong,
			Data: []byte("pong"),
		}
		if err := stream.Send(resp); err != nil {
			logger.Info("pong sent")
			return err
		}
	}
}
