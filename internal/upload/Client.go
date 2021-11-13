package upload

import (
	"apex/internal/pb"
	"apex/pkg/config"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Client is a client to call upload service RPCs
type client struct {
	service pb.UploadServiceClient
	log     *zap.Logger
}

func NewUploadClient(cc *grpc.ClientConn, logger *zap.Logger) *client {
	return &client{
		service: pb.NewUploadServiceClient(cc),
		log:     logger,
	}
}

func (c *client) UploadClient(filePath string) {
	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		c.log.Fatal("cannot get file info", zap.Error(err))
	}

	if fileInfo.IsDir() {
		c.log.Fatal("expected file : found dir")
	}

	if float64(fileInfo.Size()) > config.MaxUploadFileSize {
		c.log.Fatal("size reached",
			zap.String("size",
				fmt.Sprintf("%.1fMB",
					float64(fileInfo.Size())/config.MiB1)))
	}

	file, err := os.Open(filePath)
	if err != nil {
		c.log.Fatal("cannot open file", zap.Error(err))
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := c.service.UploadFile(ctx)
	if err != nil {
		c.log.Fatal("cannot upload file", zap.Error(err))
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
		c.log.Fatal("cannot send file info to server", zap.Errors("errors", []error{err, stream.RecvMsg(nil)}))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			c.log.Fatal("cannot read chunk to buffer", zap.Error(err))
		}

		req := &pb.UploadRequest{
			Data: &pb.UploadRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			c.log.Fatal("cannot send chunk to server", zap.Errors("errors", []error{err, stream.RecvMsg(nil)}))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		c.log.Fatal("cannot receive response", zap.Error(err))
	}

	fileSizeMB := fmt.Sprintf("%.1fMB", float64(res.GetTotalSize())/config.MiB1)
	c.log.Info("file uploaded", zap.String("id", res.GetId()), zap.String("size", fileSizeMB))
}
