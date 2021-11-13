package main

import (
	"apex/internal/pb"
	"apex/internal/upload"
	"flag"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 1500, "the server port")
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("✋🏾 logger init failed %v", err.Error())
	}

	defer logger.Sync()
	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("cannot start server ", zap.Error(err))
	}

	fileStore := upload.NewDiskFileStore("files")
	uploadSrv := upload.NewServer(fileStore, logger)
	grpcSrv := grpc.NewServer()

	logger.Info("gRPC server binding", zap.String("protocol", "tcp"), zap.String("addr", address))

	pb.RegisterUploadServiceServer(grpcSrv, uploadSrv)
	err = grpcSrv.Serve(listener)
	if err != nil {
		logger.Error("cannot start gRPC server", zap.Error(err))
	}
}
