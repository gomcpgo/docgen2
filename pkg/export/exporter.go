package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/savant/mcp-servers/docgen2/pkg/config"
	"github.com/savant/mcp-servers/docgen2/pkg/storage"
)

// Exporter handles document export operations
type Exporter struct {
	config          *config.Config
	storage         *storage.Storage
	markdownBuilder *MarkdownBuilder
	pandoc          *PandocWrapper
}

// NewExporter creates a new exporter
func NewExporter(cfg *config.Config, storage *storage.Storage) *Exporter {
	return &Exporter{
		config:          cfg,
		storage:         storage,
		markdownBuilder: NewMarkdownBuilder(storage),
		pandoc:          NewPandocWrapper(),
	}
}

// ExportDocument exports a document to the specified format
func (e *Exporter) ExportDocument(docID string, format string) (string, error) {
	// Validate format
	format = strings.ToLower(format)
	if format != "pdf" && format != "docx" && format != "html" {
		return "", fmt.Errorf("unsupported export format: %s (supported: pdf, docx, html)", format)
	}

	// Check if document exists
	doc, err := e.storage.GetDocument(docID)
	if err != nil {
		return "", fmt.Errorf("failed to get document: %w", err)
	}

	// Create exports directory
	exportsPath := filepath.Join(e.config.GetDocumentFolder(docID), "exports")
	if err := os.MkdirAll(exportsPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create exports directory: %w", err)
	}

	// Build markdown content
	markdownContent, err := e.markdownBuilder.BuildMarkdown(docID)
	if err != nil {
		return "", fmt.Errorf("failed to build markdown: %w", err)
	}

	// Generate output filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	sanitizedTitle := e.sanitizeFilename(doc.Title)
	outputFilename := fmt.Sprintf("%s-%s.%s", sanitizedTitle, timestamp, format)
	outputPath := filepath.Join(exportsPath, outputFilename)

	// Convert markdown to target format using Pandoc
	if err := e.pandoc.ConvertMarkdownToFormat(markdownContent, outputPath, format); err != nil {
		return "", fmt.Errorf("failed to convert document: %w", err)
	}

	// Return the path to the exported file
	return outputPath, nil
}

// ExportChapter exports a specific chapter to the specified format
func (e *Exporter) ExportChapter(docID, chapterID string, format string) (string, error) {
	// Validate format
	format = strings.ToLower(format)
	if format != "pdf" && format != "docx" && format != "html" {
		return "", fmt.Errorf("unsupported export format: %s (supported: pdf, docx, html)", format)
	}

	// Check if document and chapter exist
	doc, err := e.storage.GetDocument(docID)
	if err != nil {
		return "", fmt.Errorf("failed to get document: %w", err)
	}

	if !doc.HasChapters {
		return "", fmt.Errorf("document does not have chapters")
	}

	chapter, err := e.storage.GetChapter(docID, chapterID)
	if err != nil {
		return "", fmt.Errorf("failed to get chapter: %w", err)
	}

	// Create exports directory
	exportsPath := filepath.Join(e.config.GetDocumentFolder(docID), "exports")
	if err := os.MkdirAll(exportsPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create exports directory: %w", err)
	}

	// Build markdown content for chapter
	markdownContent, err := e.markdownBuilder.BuildChapterMarkdown(docID, chapterID)
	if err != nil {
		return "", fmt.Errorf("failed to build chapter markdown: %w", err)
	}

	// Generate output filename
	timestamp := time.Now().Format("20060102-150405")
	sanitizedTitle := e.sanitizeFilename(chapter.Title)
	outputFilename := fmt.Sprintf("%s-%s-%s.%s", sanitizedTitle, chapterID, timestamp, format)
	outputPath := filepath.Join(exportsPath, outputFilename)

	// Convert markdown to target format
	if err := e.pandoc.ConvertMarkdownToFormat(markdownContent, outputPath, format); err != nil {
		return "", fmt.Errorf("failed to convert chapter: %w", err)
	}

	return outputPath, nil
}

// GetSupportedFormats returns the list of supported export formats
func (e *Exporter) GetSupportedFormats() []string {
	return []string{"pdf", "docx", "html"}
}

// CheckDependencies checks if all required dependencies are installed
func (e *Exporter) CheckDependencies() error {
	return e.pandoc.CheckPandocInstalled()
}

// sanitizeFilename removes characters that might cause issues in filenames
func (e *Exporter) sanitizeFilename(name string) string {
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")
	
	// Remove any characters that aren't alphanumeric, underscore, or hyphen
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
		   (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}
	
	sanitized := result.String()
	
	// Truncate if too long
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}
	
	// Default name if empty
	if sanitized == "" {
		sanitized = "document"
	}
	
	return sanitized
}