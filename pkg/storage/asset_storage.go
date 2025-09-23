package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyImageToAssets copies an image to the document's assets folder
func (s *Storage) CopyImageToAssets(docID, sourcePath string) (string, error) {
	// Open source file
	source, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source image: %w", err)
	}
	defer source.Close()
	
	// Generate asset filename
	ext := filepath.Ext(sourcePath)
	baseName := strings.TrimSuffix(filepath.Base(sourcePath), ext)
	
	// Create a unique filename in assets
	assetNum := 1
	var assetName string
	var assetPath string
	
	for {
		assetName = fmt.Sprintf("%s-%03d%s", baseName, assetNum, ext)
		assetPath = filepath.Join(s.config.GetDocumentFolder(docID), "assets", assetName)
		
		if _, err := os.Stat(assetPath); os.IsNotExist(err) {
			break
		}
		assetNum++
	}
	
	// Create destination file
	dest, err := os.Create(assetPath)
	if err != nil {
		return "", fmt.Errorf("failed to create asset file: %w", err)
	}
	defer dest.Close()
	
	// Copy file
	if _, err := io.Copy(dest, source); err != nil {
		return "", fmt.Errorf("failed to copy image: %w", err)
	}
	
	// Return absolute path for use with artifact server
	return assetPath, nil
}