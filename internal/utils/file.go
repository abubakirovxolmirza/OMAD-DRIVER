package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AllowedImageExtensions defines allowed image file extensions
var AllowedImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif"}

// SaveUploadedFile saves an uploaded file to the specified directory
func SaveUploadedFile(file *multipart.FileHeader, uploadDir, subDir string) (string, error) {
	// Create upload directory if it doesn't exist
	fullDir := filepath.Join(uploadDir, subDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedExtension(ext) {
		return "", fmt.Errorf("file type not allowed: %s", ext)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(fullDir, filename)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return relative path
	return filepath.Join(subDir, filename), nil
}

// DeleteFile deletes a file from the upload directory
func DeleteFile(uploadDir, relativePath string) error {
	if relativePath == "" {
		return nil
	}

	fullPath := filepath.Join(uploadDir, relativePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// isAllowedExtension checks if the file extension is allowed
func isAllowedExtension(ext string) bool {
	for _, allowed := range AllowedImageExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}
