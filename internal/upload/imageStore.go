package upload

import (
	"fmt"
	"os"
	"sync"
)

// DiskFileStore stores file on disk, and its info on memory
type DiskFileStore struct {
	mutex      sync.RWMutex
	fileFolder string
	file       map[string]*FileInfo
}

// FileInfo contains information of the uploaded file
type FileInfo struct {
	Type string
	Path string
}

// NewDiskFileStore returns a new DiskFileStore
func NewDiskFileStore(fileFolder string) Repository {
	return &DiskFileStore{
		fileFolder: fileFolder,
		file:       make(map[string]*FileInfo),
	}
}

// Save adds a new file in store (inmmemory)
func (store *DiskFileStore) InsertFile(f File) error {
	filePath := fmt.Sprintf("%s/%s%s", store.fileFolder, f.ID, f.FileType)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}

	_, err = f.FIle_DataBuffer.WriteTo(file)
	if err != nil {
		return fmt.Errorf("cannot write : %w", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.file[f.ID] = &FileInfo{
		Type: f.FileType,
		Path: filePath,
	}

	return err
}
