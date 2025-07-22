package chunker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// SplitFile splits a file into chunks of specified size (in bytes).
func SplitFile(filePath string, chunkSize int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	totalSize := fileInfo.Size()
	totalParts := int(totalSize) / chunkSize
	if totalSize%int64(chunkSize) != 0 {
		totalParts++
	}

	outputDir := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)

	chunkPaths := []string{}

	buffer := make([]byte, chunkSize)
	for i := 0; i < totalParts; i++ {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}

		chunkFileName := fmt.Sprintf("%s.part_%d", fileName, i)
		chunkPath := filepath.Join(outputDir, chunkFileName)

		err = os.WriteFile(chunkPath, buffer[:n], os.ModePerm)
		if err != nil {
			return nil, err
		}

		chunkPaths = append(chunkPaths, chunkPath)
	}

	return chunkPaths, nil
}

// MergeChunks merges the list of chunks into a single output file.
func MergeChunks(chunkPaths []string, outputPath string) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for _, chunkPath := range chunkPaths {
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return err
		}

		_, err = outputFile.Write(chunkData)
		if err != nil {
			return err
		}
	}

	return nil
}
