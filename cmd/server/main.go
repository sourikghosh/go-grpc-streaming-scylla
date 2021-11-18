package main

import (
	"apex/internal/pb"
	"apex/internal/upload"
	"apex/pkg/scylla"
	"flag"
	"fmt"
	"net"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	repoImpelementaion := flag.String("repo", "inmem", "choose the repo implementaion it can be either inmem or scylla")
	port := flag.Int("port", 1500, "the server port")
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("✋🏾 logger init failed %v", err.Error())
	}

	defer logger.Sync()

	var repo upload.Repository

	if *repoImpelementaion == "scylla" {
		cluster := scylla.CreateCluster(gocql.Quorum, "upload", "localhost:1801", "localhost:1802", "localhost:1803")
		session, err := gocql.NewSession(*cluster)
		if err != nil {
			logger.Fatal("unable to connect to scylla", zap.Error(err))
		}

		logger.Info("successfully connected to scylla cluster")
		// Initialize scylla implementation of Repository
		repo = upload.NewScyllaRepository(logger, session)
	}

	if repo == nil {
		// initialize inmem implementation of Repository
		repo = upload.NewDiskFileStore("files")
		logger.Info("successfully connected to Disk FileStore")
	}

	uploadSrv := upload.NewServer(repo, logger)
	grpcSrv := grpc.NewServer()

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("cannot start server ", zap.Error(err))
	}

	logger.Info("gRPC server binding", zap.String("protocol", "tcp"), zap.String("addr", address))

	pb.RegisterUploadServiceServer(grpcSrv, uploadSrv)
	err = grpcSrv.Serve(listener)
	if err != nil {
		logger.Error("cannot start gRPC server", zap.Error(err))
	}
}
