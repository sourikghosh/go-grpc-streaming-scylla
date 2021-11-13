package main

import (
	"apex/internal/upload"
	"flag"
	"fmt"
	"log"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("✋🏾 logger init failed %v", err.Error())
	}

	defer logger.Sync()
	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	client := upload.NewUploadClient(conn, logger)
	client.UploadClient("tmp/song.mp3")
}
