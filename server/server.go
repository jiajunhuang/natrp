package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/jiajunhuang/natrp/errors"
	"github.com/jiajunhuang/natrp/pb/serverpb"
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
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		logger.Error("failed to listen", zap.Error(err))
		return err
	}
	defer listener.Close()
	addrList := strings.Split(listener.Addr().String(), ":")
	addr := fmt.Sprintf("%s:%s", wanIP, addrList[len(addrList)-1])

	remoteReadChan := make(chan []byte, 1024)
	remoteWriteChan := make(chan []byte, 1024)
	go func() {
		defer close(remoteReadChan)
		defer close(remoteWriteChan)

		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Error("failed to receive", zap.Error(err))
				return
			}
			defer conn.Close()

			go func() {
				for {
					data := <-remoteWriteChan

					if _, err := conn.Write(data); err != nil {
						logger.Error("failed to write", zap.Error(err))
						return
					}
				}
			}()

			data := make([]byte, 1024)
			for {
				n, err := conn.Read(data)
				if err != nil {
					logger.Error("failed to receive", zap.Error(err))
					return
				}

				remoteReadChan <- data[:n]
			}
		}
	}()

	streamReqChan := make(chan *serverpb.MsgRequest, 1024)
	go func() {
		defer close(streamReqChan)
		for {
			req, err := stream.Recv()
			if err != nil {
				logger.Error("failed to receive request from stream", zap.Error(err))
				return
			}

			streamReqChan <- req
		}
	}()

	for {
		select {
		case data := <-remoteReadChan:
			go stream.Send(&serverpb.MsgResponse{Type: serverpb.MsgType_Proxy, Code: serverpb.Code_CodeSucceed, Data: data})
		case req := <-streamReqChan:
			switch req.Type {
			case serverpb.MsgType_Proxy:
				go func() { remoteWriteChan <- req.Data }()
			case serverpb.MsgType_Disconnect:
				return nil
			case serverpb.MsgType_PingPong:
				go func() {
					resp := &serverpb.MsgResponse{
						Code: serverpb.Code_CodeSucceed,
						Type: serverpb.MsgType_PingPong,
						Data: []byte(addr),
					}
					logger.Info("ping received", zap.String("token", req.Token))
					if err := stream.Send(resp); err != nil {
						logger.Error("failed to send pong", zap.Error(err))
						return
					}
				}()
			default:
				logger.Error("bad request type", zap.Any("req", req))
				return errors.ErrBadRequest
			}
		}
	}
}
