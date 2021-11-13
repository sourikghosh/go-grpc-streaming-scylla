package upload

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

// FileStore is an interface to store upload files
type FileStore interface {
	// Save saves a new file to the store
	Save(fileType string, fileData bytes.Buffer) (string, error)
}

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
func NewDiskFileStore(fileFolder string) *DiskFileStore {
	return &DiskFileStore{
		fileFolder: fileFolder,
		file:       make(map[string]*FileInfo),
	}
}

// Save adds a new file in store (inmmemory)
func (store *DiskFileStore) Save(
	fileType string,
	fileData bytes.Buffer,
) (string, error) {
	fileID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate file id: %w", err)
	}

	filePath := fmt.Sprintf("%s/%s%s", store.fileFolder, fileID, fileType)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot create file: %w", err)
	}

	_, err = fileData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write : %w", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.file[fileID.String()] = &FileInfo{
		Type: fileType,
		Path: filePath,
	}

	return fileID.String(), nil
}
