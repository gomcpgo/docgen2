package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	
	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/config"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
	"gopkg.in/yaml.v3"
)

// Storage handles all file operations for documents
type Storage struct {
	config *config.Config
}

// NewStorage creates a new storage instance
func NewStorage(cfg *config.Config) *Storage {
	return &Storage{config: cfg}
}

// CreateDocument creates a new document
func (s *Storage) CreateDocument(title string, hasChapters bool, author string) (string, error) {
	// Generate document ID from title
	docID := s.generateDocumentID(title)
	
	// Create document folder
	docPath := s.config.GetDocumentFolder(docID)
	if err := os.MkdirAll(docPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create document folder: %w", err)
	}
	
	// Create assets folder
	assetsPath := filepath.Join(docPath, "assets")
	if err := os.MkdirAll(assetsPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create assets folder: %w", err)
	}
	
	// Create both blocks and chapters folders to support mixed documents
	blocksPath := filepath.Join(docPath, "blocks")
	if err := os.MkdirAll(blocksPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create blocks folder: %w", err)
	}
	
	chaptersPath := filepath.Join(docPath, "chapters")
	if err := os.MkdirAll(chaptersPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create chapters folder: %w", err)
	}
	
	// Create manifest
	doc := &document.Document{
		Title:       title,
		Author:      author,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		HasChapters: hasChapters,
		Blocks:      []blocks.BlockReference{},
		Chapters:    []document.ChapterReference{},
	}
	
	// Save manifest
	if err := s.SaveDocument(docID, doc); err != nil {
		// Clean up on error
		os.RemoveAll(docPath)
		return "", fmt.Errorf("failed to save manifest: %w", err)
	}
	
	return docID, nil
}

// generateDocumentID generates a unique document ID from the title
func (s *Storage) generateDocumentID(title string) string {
	// Clean title for filesystem
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)
	cleaned := reg.ReplaceAllString(strings.ToLower(title), "-")
	cleaned = strings.Trim(cleaned, "-")
	
	// Truncate if too long
	if len(cleaned) > 50 {
		cleaned = cleaned[:50]
	}
	
	// Check for uniqueness and add number if needed
	baseID := cleaned
	counter := 1
	for {
		docPath := s.config.GetDocumentFolder(cleaned)
		if _, err := os.Stat(docPath); os.IsNotExist(err) {
			break
		}
		cleaned = fmt.Sprintf("%s-%d", baseID, counter)
		counter++
	}
	
	return cleaned
}

// GetDocument loads a document manifest
func (s *Storage) GetDocument(docID string) (*document.Document, error) {
	manifestPath := filepath.Join(s.config.GetDocumentFolder(docID), "manifest.yaml")
	
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("document not found: %s", docID)
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	
	var doc document.Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	
	return &doc, nil
}

// SaveDocument saves a document manifest
func (s *Storage) SaveDocument(docID string, doc *document.Document) error {
	doc.UpdatedAt = time.Now()
	
	data, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}
	
	manifestPath := filepath.Join(s.config.GetDocumentFolder(docID), "manifest.yaml")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}
	
	return nil
}

// ListDocuments returns a list of all document IDs
func (s *Storage) ListDocuments() ([]string, error) {
	docsPath := s.config.GetDocumentsFolder()
	
	entries, err := os.ReadDir(docsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read documents folder: %w", err)
	}
	
	var docIDs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it's a valid document by looking for manifest.yaml
			manifestPath := filepath.Join(docsPath, entry.Name(), "manifest.yaml")
			if _, err := os.Stat(manifestPath); err == nil {
				docIDs = append(docIDs, entry.Name())
			}
		}
	}
	
	return docIDs, nil
}

// DeleteDocument deletes a document and all its contents
func (s *Storage) DeleteDocument(docID string) error {
	docPath := s.config.GetDocumentFolder(docID)
	
	// Check if document exists
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		return fmt.Errorf("document not found: %s", docID)
	}
	
	// Delete the entire document folder
	if err := os.RemoveAll(docPath); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	
	return nil
}