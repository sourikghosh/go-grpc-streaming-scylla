package internal

// // This variant traverses all directories in parallel.
// // It uses a concurrency-limiting counting semaphore
// // to avoid opening too many files at once.

// import (
// 	"apex/pkg/config"
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"
// 	"sync"

// 	"go.uber.org/zap"
// )

// type fileInfo struct {
// 	Filename string
// 	Size     int64
// }

// func RetriveFilesTOUpload(roots []string, n *sync.WaitGroup, fileNames chan<- string) {
// 	limitReached := make(chan struct{}, 1)

// 	if len(roots) == 0 {
// 		roots = []string{"."}
// 	}

// 	// Traverse each root of the file tree in parallel.
// 	discoveredFiles := make(chan fileInfo)

// 	for _, root := range roots {
// 		n.Add(1)
// 		go traverseDir(root, n, discoveredFiles, limitReached)
// 	}

// 	go func() {
// 		n.Wait()
// 		close(discoveredFiles)
// 		close(fileNames)
// 	}()

// 	var nbytes int64
// 	for fInfo := range discoveredFiles {
// 		nbytes += fInfo.Size

// 		// validation check for Max Upload Size.
// 		checkVal := float64(nbytes)
// 		fmt.Printf("size reached %.1fmb\n", checkVal/1e6)

// 		if checkVal >= config.MaxUploadFileSize {
// 			config.ZapLogger.Warn("Max upload size limit reached ...")

// 			close(fileNames)

// 			limitReached <- struct{}{}
// 			<-discoveredFiles

// 			return
// 		}

// 		fileNames <- fInfo.Filename
// 	}
// }

// // walkDir recursively walks the file tree rooted at dir
// // and sends the size of each found file on fileSizes.
// func traverseDir(dir string, n *sync.WaitGroup, discoveredFiles chan<- fileInfo, quit <-chan struct{}) {
// 	defer func() {
// 		n.Done()
// 	}()

// 	for _, entry := range dirEnts(dir) {
// 		select {
// 		case <-quit:
// 			close(discoveredFiles)
// 			return

// 		default:
// 			subdir := filepath.Join(dir, entry.Name())
// 			if entry.IsDir() {
// 				n.Add(1)
// 				go traverseDir(subdir, n, discoveredFiles, quit)
// 			} else {
// 				discoveredFiles <- fileInfo{
// 					subdir,
// 					entry.Size(),
// 				}
// 			}
// 		}
// 	}
// }

// // sema is a counting semaphore for limiting concurrency in dirents.
// var sema = make(chan struct{}, 20)

// // dirents returns the entries of directory dir.
// func dirEnts(dir string) []os.FileInfo {
// 	sema <- struct{}{}        // acquire token
// 	defer func() { <-sema }() // release token

// 	entries, err := ioutil.ReadDir(dir)
// 	if err != nil {
// 		config.ZapLogger.Error("failed to read", zap.String("dir", dir), zap.Error(err))
// 		return nil
// 	}

// 	return entries
// }
