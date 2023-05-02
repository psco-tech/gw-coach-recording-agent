package uploader

import (
	"log"
	"os"
	"path/filepath"
)

func GetUploadsDirectory() (string, error) {
	// Set outputDir to the directory where the application is running
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
		return "", err
	}
	baseDir := filepath.Dir(executablePath)
	outputDir := filepath.Join(baseDir, "uploads")
	log.Printf("Output directory: %s", outputDir)
	// Create the 'uploads' directory if it doesn't exist
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("Error creating 'uploads' directory: %v", err)
		return "", err
	}
	return outputDir, nil
}
