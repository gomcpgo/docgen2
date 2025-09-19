package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the configuration for the DocGen2 server
type Config struct {
	RootFolder string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	
	// Get root folder from environment or use default
	cfg.RootFolder = os.Getenv("DOCGEN_ROOT")
	if cfg.RootFolder == "" {
		// Default to current directory/docgen_data
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		cfg.RootFolder = filepath.Join(cwd, "docgen_data")
	}
	
	// Ensure root folder is absolute
	absPath, err := filepath.Abs(cfg.RootFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	cfg.RootFolder = absPath
	
	// Create root folder if it doesn't exist
	if err := os.MkdirAll(cfg.RootFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root folder: %w", err)
	}
	
	// Create documents subfolder
	docsFolder := filepath.Join(cfg.RootFolder, "documents")
	if err := os.MkdirAll(docsFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create documents folder: %w", err)
	}
	
	return cfg, nil
}

// GetDocumentsFolder returns the path to the documents folder
func (c *Config) GetDocumentsFolder() string {
	return filepath.Join(c.RootFolder, "documents")
}

// GetDocumentFolder returns the path to a specific document folder
func (c *Config) GetDocumentFolder(docID string) string {
	return filepath.Join(c.GetDocumentsFolder(), docID)
}