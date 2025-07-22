package routes

import (
	"github.com/AyushOJOD/file-sharing-backend/internal/downloader"
	"github.com/AyushOJOD/file-sharing-backend/internal/storage"
	"github.com/AyushOJOD/file-sharing-backend/internal/uploader"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, s3Client *storage.S3Client) {
	uploadHandler := uploader.NewUploadHandler(s3Client)
	downloadHandler := downloader.NewDownloadHandler(s3Client)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.POST("/upload/chunk", uploadHandler.UploadChunk)

	r.POST("/upload", uploadHandler.StartUpload)

	r.GET("download/:fileID", downloadHandler.DownloadFile)
}
