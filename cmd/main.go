package main

import (
	"apex/internal"
	"apex/pkg/config"
	"fmt"
	"runtime"
	"sync"

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
	var n sync.WaitGroup

	go internal.RetriveFilesTOUpload([]string{"/mnt/e/image/GAmes"}, &n, fileNames)

	config.ZapLogger.Info("the files to upload")
	for fN := range fileNames {
		fmt.Println("recived all filenames in main: ", fN)
	}

	n.Wait()

	fmt.Println(runtime.NumGoroutine())
	panic("dump stack here:")
}
