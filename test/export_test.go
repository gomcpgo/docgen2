package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/config"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
	"github.com/savant/mcp-servers/docgen2/pkg/export"
	"github.com/savant/mcp-servers/docgen2/pkg/storage"
)

func setupTestDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "docgen-export-*")
	if err != nil {
		t.Fatal(err)
	}
	
	// Create documents folder
	docsPath := filepath.Join(tempDir, "documents")
	if err := os.MkdirAll(docsPath, 0755); err != nil {
		t.Fatal(err)
	}
	
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	
	return tempDir, cleanup
}

func TestExportDocumentToHTML(t *testing.T) {
	// Setup
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	cfg := &config.Config{RootFolder: tmpDir}
	stor := storage.NewStorage(cfg)
	exp := export.NewExporter(cfg, stor)

	// Check if Pandoc is available (skip test if not)
	if err := exp.CheckDependencies(); err != nil {
		t.Skip("Pandoc not installed, skipping export test")
	}

	// Create a document with content
	docID, err := stor.CreateDocument("Test Export Document", false, "Test Author")
	if err != nil {
		t.Fatal(err)
	}

	// Add some blocks
	heading := &blocks.HeadingBlock{
		BaseBlock: blocks.BaseBlock{ID: ""},
		Level:     1,
		Text:      "Main Heading",
	}
	err = stor.AddBlock(docID, "", heading, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}

	markdown := &blocks.MarkdownBlock{
		BaseBlock: blocks.BaseBlock{ID: ""},
		Content:   "This is a paragraph with **bold** and *italic* text.",
	}
	err = stor.AddBlock(docID, "", markdown, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}

	table := &blocks.TableBlock{
		BaseBlock: blocks.BaseBlock{ID: ""},
		Headers:   []string{"Name", "Age", "City"},
		Rows: [][]string{
			{"Alice", "30", "New York"},
			{"Bob", "25", "London"},
		},
	}
	err = stor.AddBlock(docID, "", table, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}

	// Export to HTML
	outputPath, err := exp.ExportDocument(docID, "html")
	if err != nil {
		t.Fatalf("Failed to export document: %v", err)
	}

	// Check that file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Exported file does not exist: %s", outputPath)
	}

	// Read and check content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	htmlContent := string(content)

	// Check that expected content is present
	if !strings.Contains(htmlContent, "Main Heading") {
		t.Error("Exported HTML does not contain expected heading")
	}

	if !strings.Contains(htmlContent, "bold") || !strings.Contains(htmlContent, "italic") {
		t.Error("Exported HTML does not contain expected formatted text")
	}

	// HTML table should have table tags
	if !strings.Contains(htmlContent, "<table") {
		t.Error("Exported HTML does not contain table")
	}

	// Clean up
	os.Remove(outputPath)
}

func TestExportChapteredDocument(t *testing.T) {
	// Setup
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	cfg := &config.Config{RootFolder: tmpDir}
	stor := storage.NewStorage(cfg)
	exp := export.NewExporter(cfg, stor)

	// Check if Pandoc is available
	if err := exp.CheckDependencies(); err != nil {
		t.Skip("Pandoc not installed, skipping export test")
	}

	// Create a chaptered document
	docID, err := stor.CreateDocument("Chaptered Document", true, "Test Author")
	if err != nil {
		t.Fatal(err)
	}

	// Add chapters with content
	chapter1ID, err := stor.AddChapter(docID, "Chapter One", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}

	heading1 := &blocks.HeadingBlock{
		BaseBlock: blocks.BaseBlock{ID: ""},
		Level:     2,
		Text:      "Section 1.1",
	}
	err = stor.AddBlock(docID, chapter1ID, heading1, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}

	content1 := &blocks.MarkdownBlock{
		BaseBlock: blocks.BaseBlock{ID: ""},
		Content:   "Content for chapter one.",
	}
	err = stor.AddBlock(docID, chapter1ID, content1, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}

	chapter2ID, err := stor.AddChapter(docID, "Chapter Two", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}

	content2 := &blocks.MarkdownBlock{
		BaseBlock: blocks.BaseBlock{ID: ""},
		Content:   "Content for chapter two.",
	}
	err = stor.AddBlock(docID, chapter2ID, content2, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}

	// Export full document to HTML
	outputPath, err := exp.ExportDocument(docID, "html")
	if err != nil {
		t.Fatalf("Failed to export chaptered document: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Exported file does not exist: %s", outputPath)
	}

	// Read content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	htmlContent := string(content)

	// Check chapters are present
	if !strings.Contains(htmlContent, "Chapter One") {
		t.Error("Chapter One not found in exported document")
	}

	if !strings.Contains(htmlContent, "Chapter Two") {
		t.Error("Chapter Two not found in exported document")
	}

	if !strings.Contains(htmlContent, "Section 1.1") {
		t.Error("Section heading not found in exported document")
	}

	// Export single chapter
	chapterPath, err := exp.ExportChapter(docID, chapter1ID, "html")
	if err != nil {
		t.Fatalf("Failed to export chapter: %v", err)
	}

	// Check chapter export exists
	if _, err := os.Stat(chapterPath); os.IsNotExist(err) {
		t.Fatalf("Exported chapter file does not exist: %s", chapterPath)
	}

	// Read chapter content
	chapterContent, err := os.ReadFile(chapterPath)
	if err != nil {
		t.Fatal(err)
	}

	chapterHTML := string(chapterContent)

	// Check chapter content
	if !strings.Contains(chapterHTML, "Chapter One") {
		t.Error("Chapter title not in chapter export")
	}

	if strings.Contains(chapterHTML, "Chapter Two") {
		t.Error("Chapter Two should not be in Chapter One export")
	}

	// Clean up
	os.Remove(outputPath)
	os.Remove(chapterPath)
}

func TestMarkdownBuilder(t *testing.T) {
	// Setup
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	cfg := &config.Config{RootFolder: tmpDir}
	stor := storage.NewStorage(cfg)
	mb := export.NewMarkdownBuilder(stor)

	// Create document
	docID, err := stor.CreateDocument("Markdown Test", false, "Test Author")
	if err != nil {
		t.Fatal(err)
	}

	// Add various block types
	heading := &blocks.HeadingBlock{
		BaseBlock: blocks.BaseBlock{ID: ""},
		Level:     2,
		Text:      "Test Heading",
	}
	stor.AddBlock(docID, "", heading, document.Position{Type: document.PositionEnd})

	pageBreak := &blocks.PageBreakBlock{
		BaseBlock: blocks.BaseBlock{ID: ""},
	}
	stor.AddBlock(docID, "", pageBreak, document.Position{Type: document.PositionEnd})

	table := &blocks.TableBlock{
		BaseBlock: blocks.BaseBlock{ID: ""},
		Headers:   []string{"Col1", "Col2"},
		Rows: [][]string{
			{"A", "B"},
			{"C", "D"},
		},
	}
	stor.AddBlock(docID, "", table, document.Position{Type: document.PositionEnd})

	// Build markdown
	markdown, err := mb.BuildMarkdown(docID)
	if err != nil {
		t.Fatalf("Failed to build markdown: %v", err)
	}

	// Check markdown content
	if !strings.Contains(markdown, "## Test Heading") {
		t.Error("Heading not properly converted to markdown")
	}

	if !strings.Contains(markdown, "\\newpage") {
		t.Error("Page break not in markdown")
	}

	if !strings.Contains(markdown, "| Col1 | Col2 |") {
		t.Error("Table headers not in markdown")
	}

	if !strings.Contains(markdown, "| --- | --- |") {
		t.Error("Table separator not in markdown")
	}

	if !strings.Contains(markdown, "| A | B |") {
		t.Error("Table row not in markdown")
	}
}

func TestExportInvalidFormat(t *testing.T) {
	// Setup
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	cfg := &config.Config{RootFolder: tmpDir}
	stor := storage.NewStorage(cfg)
	exp := export.NewExporter(cfg, stor)

	// Create document
	docID, err := stor.CreateDocument("Test Doc", false, "Author")
	if err != nil {
		t.Fatal(err)
	}

	// Try to export with invalid format
	_, err = exp.ExportDocument(docID, "invalid")
	if err == nil {
		t.Error("Expected error for invalid format")
	}

	if !strings.Contains(err.Error(), "unsupported export format") {
		t.Errorf("Expected 'unsupported export format' error, got: %v", err)
	}
}

func TestExportNonExistentDocument(t *testing.T) {
	// Setup
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	cfg := &config.Config{RootFolder: tmpDir}
	stor := storage.NewStorage(cfg)
	exp := export.NewExporter(cfg, stor)

	// Try to export non-existent document
	_, err := exp.ExportDocument("non-existent", "html")
	if err == nil {
		t.Error("Expected error for non-existent document")
	}

	if !strings.Contains(err.Error(), "document not found") {
		t.Errorf("Expected 'document not found' error, got: %v", err)
	}
}

func TestGetSupportedFormats(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	cfg := &config.Config{RootFolder: tmpDir}
	stor := storage.NewStorage(cfg)
	exp := export.NewExporter(cfg, stor)

	formats := exp.GetSupportedFormats()

	// Check expected formats
	expectedFormats := []string{"pdf", "docx", "html"}
	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, got %d", len(expectedFormats), len(formats))
	}

	for _, expected := range expectedFormats {
		found := false
		for _, format := range formats {
			if format == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format '%s' not found", expected)
		}
	}
}