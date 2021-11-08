package internal

// This variant traverses all directories in parallel.
// It uses a concurrency-limiting counting semaphore
// to avoid opening too many files at once.

import (
	"apex/pkg/config"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
)

type fileInfo struct {
	Filename string
	Size     int64
}

func Exec(roots []string, fileNames chan<- string) {
	if len(roots) == 0 {
		roots = []string{"."}
	}

	// Traverse each root of the file tree in parallel.
	discoveredFiles := make(chan fileInfo)
	// // Files to upload in parallel.
	// fileNames := make(chan string)

	var n sync.WaitGroup
	for _, root := range roots {
		n.Add(1)
		go traverseDir(root, &n, discoveredFiles)
	}

	go func() {
		n.Wait()
		close(discoveredFiles)
	}()

	var nbytes int64
	for fInfo := range discoveredFiles {
		fmt.Println("finfo reading from discoveredFiles: ", fInfo.Filename)
		nbytes += fInfo.Size
		fmt.Println("before if")
		// validation check for Max Upload Size.
		if float64(nbytes) >= config.MaxUploadFileSize {
			config.ZapLogger.Warn("Max upload size limit reached ...")
			close(fileNames)

			break
		}

		fmt.Println("after If")
		fileNames <- fInfo.Filename
		fmt.Println("next iter of for")
	}
}

// walkDir recursively walks the file tree rooted at dir
// and sends the size of each found file on fileSizes.
func traverseDir(dir string, n *sync.WaitGroup, discoveredFiles chan<- fileInfo) {
	defer n.Done()
	for _, entry := range dirEnts(dir) {
		if entry.IsDir() {
			n.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go traverseDir(subdir, n, discoveredFiles)
		} else {
			fmt.Println("adding entry from traverseDir", entry.Name())

			discoveredFiles <- fileInfo{
				dir + "/" + entry.Name(),
				entry.Size(),
			}
		}
	}
}

// sema is a counting semaphore for limiting concurrency in dirents.
var sema = make(chan struct{}, 20)

// dirents returns the entries of directory dir.
func dirEnts(dir string) []os.FileInfo {
	sema <- struct{}{}        // acquire token
	defer func() { <-sema }() // release token

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		config.ZapLogger.Error("failed to read dir", zap.Error(err))
		return nil
	}

	return entries
}
