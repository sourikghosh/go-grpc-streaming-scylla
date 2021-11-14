package main

import (
	"apex/internal/upload"
	"apex/pkg/config"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	dirName := flag.String("dir", ".", "Absolute Path to search for file/s to upload")
	maxWorkerCount := flag.Int("w", 6, "no of concurrent worker count to upload files")
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("✋🏾 logger init failed %v", err.Error())
	}
	defer logger.Sync()

	config.MaxWorkerCount = *maxWorkerCount
	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("cannot dial server: ", zap.Error(err))
	}

	if err = uploadFiles(context.Background(), conn, *dirName, logger); err != nil {
		logger.Error("error from server", zap.Error(err))
	}
}

func uploadFiles(ctx context.Context, cc *grpc.ClientConn, dir string, logger *zap.Logger) error {
	cli := upload.NewUploadClient(ctx, cc, logger, dir)
	var errorUploadbulk error

	// reading all the files in the dir
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logger.Fatal("unable to read", zap.String("path", dir), zap.Error(err))
	}
	defer cli.Stop()

	// putting the file to request channel if its not dir
	go func() {
		for _, file := range files {
			if !file.IsDir() {
				cli.Do(filepath.Join(dir, file.Name()))
			}
		}
	}()

	// ranging over each file to check the upload status
	for _, file := range files {
		if !file.IsDir() {
			select {
			case msg := <-cli.DoneRequest:
				logger.Info("file uploaded", zap.String("msg", msg))

			case req := <-cli.FailRequest:
				fmt.Println("failed to  send " + req)
				errorUploadbulk = errors.Wrapf(errorUploadbulk, " Failed to send %s", req)
			}
		}
	}

	fmt.Println("All done ")
	return errorUploadbulk
}
