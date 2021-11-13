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
	Save(imageType string, imageData bytes.Buffer) (string, error)
}

// DiskFileStore stores image on disk, and its info on memory
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

// NewDiskFileStore returns a new DiskImageStore
func NewDiskFileStore(imageFolder string) *DiskFileStore {
	return &DiskFileStore{
		fileFolder: imageFolder,
		file:       make(map[string]*FileInfo),
	}
}

// Save adds a new file in store (inmmemory)
func (store *DiskFileStore) Save(
	imageType string,
	imageData bytes.Buffer,
) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id: %w", err)
	}

	imagePath := fmt.Sprintf("%s/%s%s", store.fileFolder, imageID, imageType)

	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}

	_, err = imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.file[imageID.String()] = &FileInfo{
		Type: imageType,
		Path: imagePath,
	}

	return imageID.String(), nil
}
