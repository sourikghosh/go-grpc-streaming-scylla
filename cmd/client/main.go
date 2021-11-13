package main

import (
	"apex/internal/pb"
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func uploadImage(uploadClient pb.UploadServiceClient, imagePath string) {
	fileInfo, err := os.Lstat(imagePath)
	if err != nil {
		log.Fatal("cannot get file info", err)
	}

	if fileInfo.IsDir() {
		log.Fatal("expected file found dir")
	}

	if float64(fileInfo.Size()) > 5e6 {
		log.Fatalf("size reached %.1fmb\n", float64(fileInfo.Size())/1e6)
	}

	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := uploadClient.UploadFile(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}

	req := &pb.UploadRequest{
		Data: &pb.UploadRequest_Info{
			Info: &pb.FileInfo{
				FileName: filepath.Base(imagePath),
				Type:     filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info to server: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &pb.UploadRequest{
			Data: &pb.UploadRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image uploaded with id: %s, size: %d", res.GetId(), res.GetTotalSize())
}

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
	uploadImage(uploadClient, "tmp/game.png")
}
