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
	"sync"
	"time"

	"github.com/k0kubun/go-ansi"
	bar "github.com/schollz/progressbar/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Client is a client to call upload service RPCs
type client struct {
	// dir is the DirectoryName where it finds all files and then uploads concurrently.
	dir string
	// uploadService has the upload RPCs
	service pb.UploadServiceClient
	log     *zap.Logger
	ctx     context.Context
	wg      sync.WaitGroup
	// each request is a filepath on client accessible to client
	requests    chan string
	DoneRequest chan string
	FailRequest chan string
}

func NewUploadClient(ctx context.Context, cc *grpc.ClientConn, logger *zap.Logger, dirname string) *client {
	uploadCli := &client{
		ctx:         ctx,
		dir:         dirname,
		service:     pb.NewUploadServiceClient(cc),
		log:         logger,
		requests:    make(chan string),
		DoneRequest: make(chan string),
		FailRequest: make(chan string),
	}

	// concurrency can be modified by setting `MaxWorkerCount`
	for i := 0; i < config.MaxWorkerCount; i++ {
		uploadCli.wg.Add(1)
		go uploadCli.worker(i + 1)
	}

	return uploadCli
}

func (c *client) Stop() {
	close(c.requests)
	c.wg.Wait()
}

func (c *client) Do(filepath string) {
	c.requests <- filepath
}

func (c *client) worker(workerID int) {
	defer func() {
		c.wg.Done()
	}()

	for request := range c.requests {
		fileInfo, err := os.Lstat(request)
		if err != nil {
			c.log.Fatal("cannot get file info", zap.Error(err))
		}

		if float64(fileInfo.Size()) > config.MaxUploadFileSize {
			c.log.Fatal("size reached",
				zap.String("size",
					fmt.Sprintf("%.1fMB",
						float64(fileInfo.Size())/config.MiB1)))
		}

		file, err := os.Open(request)
		if err != nil {
			c.log.Fatal("failed to open file", zap.String("file", request), zap.Error(err))
		}
		defer file.Close()

		//start uploading ...
		stream, err := c.service.UploadFile(c.ctx)
		if err != nil {
			c.log.Fatal("failed to create upload stream", zap.String("file", request), zap.Error(err))
		}

		req := &pb.UploadRequest{
			Data: &pb.UploadRequest_Info{
				Info: &pb.FileInfo{
					FileName: filepath.Base(request),
					Type:     filepath.Ext(request),
				},
			},
		}

		err = stream.Send(req)
		if err != nil {
			c.log.Fatal("cannot send file info to server", zap.Errors("errors", []error{err, stream.RecvMsg(nil)}))
		}

		//start progress bar
		progressBar := bar.NewOptions(int(fileInfo.Size()),
			bar.OptionSetWriter(ansi.NewAnsiStdout()),
			bar.OptionEnableColorCodes(true),
			bar.OptionShowBytes(true),
			bar.OptionSetWidth(15),
			bar.OptionSetDescription("[cyan]uploading[reset] "+fmt.Sprint(workerID)+" ..."),
			bar.OptionSetTheme(bar.Theme{
				Saucer:        "[green]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)

		//create a buffer of chunkSize to be streamed
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
				// bar.FinishPrint("failed to send chunk via stream file : " + request)
				c.log.Error("cannot send chunk to server", zap.Errors("errors", []error{err, stream.RecvMsg(nil)}))
				break
			}

			if err = progressBar.Add64(int64(n)); err != nil {
				c.log.Error("err occurred progressbar", zap.Error(err))
			}
		}

		res, err := stream.CloseAndRecv()
		if err != nil {
			c.log.Error("cannot receive response", zap.Error(err))

			progressBar.Finish()
			c.FailRequest <- request

			return
		}

		c.log.Info("writing for done", zap.String("file", request), zap.Int("workerID", workerID))

		c.DoneRequest <- request + " id:" + res.GetId() + " size:" + fmt.Sprintf("%.1fMB", float64(res.GetTotalSize())/config.MiB1)
		progressBar.Finish()
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
