package main

import (
	"apex/internal/pb"
	"apex/internal/upload"
	"fmt"
	"log"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("✋🏾 logger init failed %v", err.Error())
	}

	defer logger.Sync()
	conn, err := grpc.Dial("0.0.0.0:1500", grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	uploadClient := pb.NewUploadServiceClient(conn)
	upload.UploadClient(uploadClient, "tmp/game.png")
}
