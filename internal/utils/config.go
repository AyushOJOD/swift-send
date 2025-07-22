package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AwsBucketName string
	AwsRegion     string
	AwsAccessKey  string
	AwsSecretKey  string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	return Config{
		AwsBucketName: getEnv("AWS_BUCKET_NAME", ""),
		AwsRegion:     getEnv("AWS_REGION", "ap-south-1"),
		AwsAccessKey:  getEnv("AWS_ACCESS_KEY_ID", ""),
		AwsSecretKey:  getEnv("AWS_SECRET_ACCESS_KEY", ""),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
