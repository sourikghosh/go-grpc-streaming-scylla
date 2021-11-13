package main

import (
	"apex/internal/pb"
	"apex/internal/upload"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("✋🏾 logger init failed %v", err.Error())
	}

	defer logger.Sync()
	address := fmt.Sprintf("0.0.0.0:%d", 1500)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("cannot start server ", zap.Error(err))
	}

	fileStore := upload.NewDiskFileStore("files")
	uploadSrv := upload.NewServer(fileStore, logger)
	grpcSrv := grpc.NewServer()

	pb.RegisterUploadServiceServer(grpcSrv, uploadSrv)
	err = grpcSrv.Serve(listener)
	if err != nil {
		logger.Error("cannot start gRPC server", zap.Error(err))
	}
}
