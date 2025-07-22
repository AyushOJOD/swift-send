package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/AyushOJOD/file-sharing-backend/internal/models"
	"github.com/AyushOJOD/file-sharing-backend/internal/storage"
	"github.com/gin-gonic/gin"
)

type DownloadHandler struct {
	S3 *storage.S3Client
}

func NewDownloadHandler(s3 *storage.S3Client) *DownloadHandler {
	return &DownloadHandler{
		S3: s3,
	}
}

func (h *DownloadHandler) DownloadFile(c *gin.Context) {
	fileID := c.Param("fileID")
	if fileID == "" {
		c.JSON(400, gin.H{"error": "fileID is required"})
		return
	}

	ctx := context.TODO()

	// Fetch manifest
	manifestKey := fmt.Sprintf("files/%s/manifest.json", fileID)
	manifestBytes, err := h.S3.DownloadFile(ctx, manifestKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch manifest", "details": err.Error()})
		return
	}

	var manifest models.FileMeta
	err = json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to parse manifest", "details": err.Error()})
		return
	}

	expectedChunks := manifest.TotalChunks
	tmpDir := fmt.Sprintf("/tmp/%s", fileID)
	os.MkdirAll(tmpDir, 0755)

	log.Printf("Starting download of %d chunks for fileID %s", expectedChunks, fileID)

	var wg sync.WaitGroup
	var mu sync.Mutex
	downloaded := make(map[int]bool)

	for {
		allDone := false

		for i := 0; i < expectedChunks; i++ {
			mu.Lock()
			if downloaded[i] {
				mu.Unlock()
				continue
			}
			mu.Unlock()

			wg.Add(1)
			go func(chunkNo int) {
				defer wg.Done()
				chunkKey := fmt.Sprintf("files/%s/chunk_%d", fileID, chunkNo)

				for {
					chunkData, err := h.S3.DownloadFile(ctx, chunkKey)
					if err != nil {
						time.Sleep(500 * time.Millisecond) // retry
						continue
					}

					// Save locally
					err = os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("chunk_%d", chunkNo)), chunkData, 0644)
					if err != nil {
						log.Printf("Failed to save chunk %d: %v", chunkNo, err)
						return
					}

					log.Printf("Downloaded chunk %d", chunkNo)

					mu.Lock()
					downloaded[chunkNo] = true
					mu.Unlock()
					break
				}
			}(i)
		}

		wg.Wait()

		mu.Lock()
		if len(downloaded) == expectedChunks {
			allDone = true
		}
		mu.Unlock()

		if allDone {
			break
		}

		time.Sleep(200 * time.Millisecond)
	}

	// Merge chunks
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", manifest.FileName))
	c.Header("Content-Type", "application/octet-stream")

	for i := 0; i < expectedChunks; i++ {
		chunkPath := filepath.Join(tmpDir, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			c.Status(500)
			c.Writer.WriteString(fmt.Sprintf("Failed to open chunk %d: %v", i, err))
			return
		}

		_, err = io.Copy(c.Writer, chunkFile)
		chunkFile.Close()
		if err != nil {
			c.Status(500)
			c.Writer.WriteString(fmt.Sprintf("Failed to write chunk %d: %v", i, err))
			return
		}
		c.Writer.Flush()
	}

	log.Printf("Finished delivering fileID %s", fileID)

	// Clean up tmp dir
	os.RemoveAll(tmpDir)
}
