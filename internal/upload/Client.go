package upload

import (
	"apex/internal/pb"
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func UploadClient(uploadClient pb.UploadServiceClient, filePath string) {
	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		log.Fatal("cannot get file info", err)
	}

	if fileInfo.IsDir() {
		log.Fatal("expected file : found dir")
	}

	if float64(fileInfo.Size()) > 5e6 {
		log.Fatalf("size reached %.1fmb\n", float64(fileInfo.Size())/1e6)
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("cannot open file file: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := uploadClient.UploadFile(ctx)
	if err != nil {
		log.Fatal("cannot upload file: ", err)
	}

	req := &pb.UploadRequest{
		Data: &pb.UploadRequest_Info{
			Info: &pb.FileInfo{
				FileName: filepath.Base(filePath),
				Type:     filepath.Ext(filePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send file info to server: ", err, stream.RecvMsg(nil))
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

	log.Printf("file uploaded with id: %s, size: %d", res.GetId(), res.GetTotalSize())
}
