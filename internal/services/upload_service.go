package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/AyushOJOD/file-sharing-backend/internal/models"
	"github.com/AyushOJOD/file-sharing-backend/internal/storage"
)

type UploadService struct {
	S3 *storage.S3Client
}

// UploadChunks uploads all chunks concurrently to S3.
func (u *UploadService) UploadChunks(fileID string, fileName string, chunkPaths []string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(chunkPaths))

	for i, chunkPath := range chunkPaths {
		wg.Add(1)

		go func(chunkNo int, path string) {
			defer wg.Done()

			data, err := os.ReadFile(path)
			if err != nil {
				errChan <- fmt.Errorf("failed to read chunk %d: %v", chunkNo, err)
				return
			}

			key := fmt.Sprintf("files/%s/chunk_%d", fileID, chunkNo)
			err = u.S3.UploadFile(context.TODO(), key, data)
			if err != nil {
				errChan <- fmt.Errorf("failed to upload chunk %d: %v", chunkNo, err)
				return
			}

			log.Printf("Uploaded chunk %d", chunkNo)
		}(i, chunkPath)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Upload manifest
	manifest := models.FileMeta{
		FileID:      fileID,
		FileName:    fileName,
		TotalChunks: len(chunkPaths),
	}

	manifestData, _ := json.Marshal(manifest)
	manifestKey := fmt.Sprintf("files/%s/manifest.json", fileID)

	return u.S3.UploadFile(context.TODO(), manifestKey, manifestData)
}
