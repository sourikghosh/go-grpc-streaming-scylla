package main

import (
	"apex/internal"
	"apex/pkg/config"
	"fmt"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("✋🏾 logger init failed %v", err.Error())
	}

	defer logger.Sync()

	config.ZapLogger = logger
	fileNames := make(chan string)

	go internal.Exec([]string{"/mnt/e/image/GAmes"}, fileNames)

	config.ZapLogger.Info("the files to upload")
	for fN := range fileNames {
		fmt.Println("recived all filenames in main: ", fN)
	}
}
