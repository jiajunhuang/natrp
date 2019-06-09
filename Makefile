default:
	protoc -I pb/ serverpb/server.proto --go_out=plugins=grpc:pb/
	go build -o bin/client client/main.go
	go build -o bin/server server/main.go server/server.go
