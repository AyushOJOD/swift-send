package uploader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AyushOJOD/file-sharing-backend/internal/models"
	"github.com/AyushOJOD/file-sharing-backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct {
	S3 *storage.S3Client
}

func NewUploadHandler(s3 *storage.S3Client) *UploadHandler {
	return &UploadHandler{
		S3: s3,
	}
}


func (h *UploadHandler) UploadChunk(c *gin.Context) {
	fileID := c.PostForm("fileID")
	chunkNoStr := c.PostForm("chunkNo")
	fileName := c.PostForm("fileName")

	if fileID == "" || chunkNoStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileID and chunkNo are required"})
		return
	}

	chunkNo, err := strconv.Atoi(chunkNoStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chunkNo must be an integer"})
		return
	}

	file, err := c.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chunk file is required"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer f.Close()

	fileBytes := make([]byte, file.Size)
	_, err = f.Read(fileBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	s3Key := fmt.Sprintf("files/%s/chunk_%d", fileID, chunkNo)
	err = h.S3.UploadFile(context.TODO(), s3Key, fileBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload to S3", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Chunk uploaded successfully",
		"fileID":   fileID,
		"chunkNo":  chunkNo,
		"s3_key":   s3Key,
		"fileName": filepath.Base(fileName),
	})
}


func (h *UploadHandler) StartUpload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	// Create a temp file in the system's temp directory with a random name
	out, err := os.CreateTemp("", "upload-*.tmp")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create temp file",
			"details": err.Error(),
		})
		return
	}
	defer out.Close()

	tmpFilePath := out.Name()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save temp file"})
		return
	}

	// Generate unique fileID
	fileID := uuid.New().String()

	// Start async upload in a goroutine
	go func() {
		err := h.uploadInChunks(fileID, tmpFilePath, fileHeader.Filename)
		if err != nil {
			log.Printf("Upload failed for fileID %s: %v", fileID, err)
		}
		_ = os.Remove(tmpFilePath)
	}()

	baseURL := os.Getenv("BASE_URL")
	downloadURL := fmt.Sprintf("%s/download/%s", baseURL, fileID)

	fmt.Println("downloadURL", downloadURL)

	c.JSON(http.StatusOK, gin.H{
		"fileID":      fileID,
		"downloadURL": downloadURL,
	})
}


func (h *UploadHandler) uploadInChunks(fileID, filePath, originalName string) error {
	const chunkSize = 5 * 1024 * 1024 // 5 MB

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, chunkSize)
	chunkNo := 0

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		chunkData := make([]byte, n)
		copy(chunkData, buffer[:n])

		s3Key := fmt.Sprintf("files/%s/chunk_%d", fileID, chunkNo)

		err = h.S3.UploadFile(context.TODO(), s3Key, chunkData)
		if err != nil {
			return err
		}

		log.Printf("Uploaded chunk %d for fileID %s", chunkNo, fileID)
		chunkNo++
	}

	manifest := models.FileMeta{
		FileID:      fileID,
		FileName:    originalName,
		TotalChunks: chunkNo,
	}
	manifestData, _ := json.Marshal(manifest)
	manifestKey := fmt.Sprintf("files/%s/manifest.json", fileID)
	return h.S3.UploadFile(context.TODO(), manifestKey, manifestData)
}
