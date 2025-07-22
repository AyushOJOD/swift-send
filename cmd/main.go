package main

import (
	"log"
	"time"

	"github.com/AyushOJOD/file-sharing-backend/internal/routes"
	"github.com/AyushOJOD/file-sharing-backend/internal/storage"
	"github.com/AyushOJOD/file-sharing-backend/internal/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	config := utils.LoadConfig()

	// Initialize S3
	s3Client := storage.NewS3Client(config.AwsBucketName)

	// Start Gin server
	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, 
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))



	routes.SetupRoutes(r, s3Client)
	
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
