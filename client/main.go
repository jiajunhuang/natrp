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

	// send a ping first
	if err := pingpong(stream); err != nil {
		logger.Error("failed to connect to server", zap.Error(err))
		return
	}

	// connect local port
	localConn, err := net.Dial("tcp", "127.0.0.1:22")
	if err != nil {
		logger.Error("failed to connect local port", zap.Error(err))
		return
	}
	defer localConn.Close()

	// ping pong timer
	pingPongTimer := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(5))
			pingPongTimer <- struct{}{}
		}
	}()

	// read from local port
	localDataChan := make(chan []byte, 1024)
	go func() {
		defer close(localDataChan)

		data := make([]byte, 1024)

		for {
			n, err := localConn.Read(data)
			if err != nil {
				logger.Error("failed to read data from local", zap.Error(err))
				return
			}
			localDataChan <- data[:n]
		}
	}()

	// read from stream
	streamMsgRespChan := make(chan *serverpb.MsgResponse, 1024)
	go func() {
		defer close(streamMsgRespChan)

		for {
			resp, err := stream.Recv()
			if err != nil {
				logger.Error("failed to read msg from stream", zap.Error(err))
				return
			}

			streamMsgRespChan <- resp
		}
	}()

	// loop
	for {
		select {
		case <-ctx.Done():
			logger.Info("ctx had been done")
			return
		case <-pingPongTimer:
			logger.Info("time to send ping pong")
			if err := pingpong(stream); err != nil {
				logger.Error("failed to connect server", zap.Error(err))
				return
			}
			continue
		case data, ok := <-localDataChan:
			if !ok {
				logger.Error("failed to read from channel", zap.Any("chan", localDataChan))
				return
			}

			logger.Info("received data from local", zap.ByteString("data", data))
			go stream.Send(&serverpb.MsgRequest{Type: serverpb.MsgType_Proxy, Token: token, Data: data})
		case resp, ok := <-streamMsgRespChan:
			if !ok {
				logger.Error("failed to read from channel", zap.Any("chan", streamMsgRespChan))
				return
			}

			logger.Info("received data from stream", zap.Any("resp", resp))
			if resp.Type == serverpb.MsgType_PingPong {
				logger.Info("pong received", zap.ByteString("remote adddr", resp.Data))
				remoteAddr = string(resp.Data)
				continue
			}
			go localConn.Write(resp.Data)
		}
	}
}

func pingpong(stream serverpb.ServerService_MsgClient) error {
	resp := &serverpb.MsgRequest{
		Type:  serverpb.MsgType_PingPong,
		Token: token,
		Data:  []byte("ping"),
	}
	if err := stream.Send(resp); err != nil {
		logger.Error("failed to ping the server", zap.Error(err))
		return err
	}

	logger.Info("connect to server succeed", zap.Any("resp", resp))
	return nil
}
