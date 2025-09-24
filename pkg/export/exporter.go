package export

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
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

	// Copy images to export directory and get updated markdown with relative paths
	markdownContent, err := e.prepareMarkdownWithImages(docID, exportsPath)
	if err != nil {
		return "", fmt.Errorf("failed to prepare markdown with images: %w", err)
	}

	// Generate output filename (without timestamp - will overwrite existing)
	sanitizedTitle := e.sanitizeFilename(doc.Title)
	outputFilename := fmt.Sprintf("%s.%s", sanitizedTitle, format)
	outputPath := filepath.Join(exportsPath, outputFilename)

	// Convert markdown to target format using Pandoc from temp directory
	tempDir := "/tmp/docgen2-images"
	if err := e.pandoc.ConvertMarkdownToFormatInDir(markdownContent, outputPath, format, tempDir); err != nil {
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

	// Generate output filename (without timestamp - will overwrite existing)
	sanitizedTitle := e.sanitizeFilename(chapter.Title)
	outputFilename := fmt.Sprintf("%s-%s.%s", sanitizedTitle, chapterID, format)
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
	// Replace spaces with hyphens (slugify)
	name = strings.ReplaceAll(name, " ", "-")
	
	// Convert to lowercase for consistency
	name = strings.ToLower(name)
	
	// Remove any characters that aren't alphanumeric, underscore, or hyphen
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || 
		   (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}
	
	sanitized := result.String()
	
	// Truncate to 40 characters max
	if len(sanitized) > 40 {
		sanitized = sanitized[:40]
	}
	
	// Trim any trailing hyphens or underscores
	sanitized = strings.TrimRight(sanitized, "-_")
	
	// Default name if empty
	if sanitized == "" {
		sanitized = "document"
	}
	
	return sanitized
}

// prepareMarkdownWithImages copies images to export directory and returns markdown with relative paths
func (e *Exporter) prepareMarkdownWithImages(docID, exportsPath string) (string, error) {
	// Get the original markdown
	markdownContent, err := e.markdownBuilder.BuildMarkdown(docID)
	if err != nil {
		return "", fmt.Errorf("failed to build markdown: %w", err)
	}

	// Create temporary images directory with no spaces in path
	tempImagesDir := "/tmp/docgen2-images"
	if err := os.RemoveAll(tempImagesDir); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to clean temp images directory: %w", err)
	}
	if err := os.MkdirAll(tempImagesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp images directory: %w", err)
	}

	// Get document to access its blocks and chapters
	doc, err := e.storage.GetDocument(docID)
	if err != nil {
		return "", fmt.Errorf("failed to get document: %w", err)
	}

	// Collect all image paths from the document
	imagePaths := make(map[string]string) // original path -> relative path
	
	// Process document-level blocks
	if err := e.collectImagePaths(docID, doc.Blocks, imagePaths, tempImagesDir); err != nil {
		return "", fmt.Errorf("failed to collect document images: %w", err)
	}

	// Process chapter blocks
	for _, chapterRef := range doc.Chapters {
		chapter, err := e.storage.GetChapter(docID, chapterRef.ID)
		if err != nil {
			continue // Skip if chapter can't be loaded
		}
		if err := e.collectImagePaths(docID, chapter.Blocks, imagePaths, tempImagesDir); err != nil {
			return "", fmt.Errorf("failed to collect chapter images: %w", err)
		}
	}

	// Replace absolute paths with relative paths in markdown
	for originalPath, relativePath := range imagePaths {
		// Replace both angle bracket and normal formats
		markdownContent = strings.ReplaceAll(markdownContent, fmt.Sprintf("](<%s>)", originalPath), fmt.Sprintf("](%s)", relativePath))
		markdownContent = strings.ReplaceAll(markdownContent, fmt.Sprintf("](%s)", originalPath), fmt.Sprintf("](%s)", relativePath))
	}

	// Image path replacement completed successfully

	return markdownContent, nil
}

// collectImagePaths processes blocks and copies images, building the path mapping
func (e *Exporter) collectImagePaths(docID string, blockRefs []blocks.BlockReference, imagePaths map[string]string, imagesDir string) error {
	imageCounter := 1
	
	for _, blockRef := range blockRefs {
		block, err := e.storage.LoadBlock(docID, blockRef)
		if err != nil {
			continue // Skip if block can't be loaded
		}

		if imgBlock, ok := block.(*blocks.ImageBlock); ok {
			// Build the absolute source path
			config := e.storage.GetConfig()
			docFolder := config.GetDocumentFolder(docID)
			sourcePath := filepath.Join(docFolder, imgBlock.Path)

			// Create simple destination filename without spaces
			// Extract file extension
			originalName := filepath.Base(imgBlock.Path)
			ext := filepath.Ext(originalName)
			simpleName := fmt.Sprintf("img_%03d%s", imageCounter, ext)
			destPath := filepath.Join(imagesDir, simpleName)
			imageCounter++

			// Copy the image file
			if err := e.copyFile(sourcePath, destPath); err != nil {
				// If copy fails, skip this image but don't fail the whole export
				continue
			}

			// Store the path mapping with just the filename
			imagePaths[sourcePath] = simpleName
		}
	}
	return nil
}

// copyFile copies a file from source to destination, converting WebP to PNG if needed
func (e *Exporter) copyFile(src, dst string) error {
	// Check if the source file is actually a WebP file
	if e.isWebPFile(src) {
		// Convert WebP to PNG using ImageMagick
		return e.convertWebPToPNG(src, dst)
	}
	
	// Regular file copy for other formats
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// isWebPFile checks if a file is actually a WebP image
func (e *Exporter) isWebPFile(filePath string) bool {
	cmd := exec.Command("file", filePath)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	outputStr := string(output)
	return strings.Contains(outputStr, "Web/P image") || strings.Contains(outputStr, "WebP")
}

// convertWebPToPNG converts a WebP image to PNG using ImageMagick
func (e *Exporter) convertWebPToPNG(src, dst string) error {
	cmd := exec.Command("convert", src, dst)
	
	// Capture stderr for error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to convert WebP to PNG: %s: %w", string(output), err)
	}
	
	return nil
}