package storage

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/config"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
)

func setupTestStorage(t *testing.T) (*Storage, func()) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "docgen-test-*")
	if err != nil {
		t.Fatal(err)
	}
	
	cfg := &config.Config{
		RootFolder: tempDir,
	}
	
	// Create documents folder
	docsPath := filepath.Join(tempDir, "documents")
	if err := os.MkdirAll(docsPath, 0755); err != nil {
		t.Fatal(err)
	}
	
	storage := NewStorage(cfg)
	
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	
	return storage, cleanup
}

func TestCreateDocument(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Test creating a flat document
	docID, err := storage.CreateDocument("Test Document", false, "Test Author")
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	
	if docID != "test-document" {
		t.Errorf("Expected docID 'test-document', got '%s'", docID)
	}
	
	// Verify document was created
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	
	if doc.Title != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%s'", doc.Title)
	}
	
	if doc.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", doc.Author)
	}
	
	if doc.HasChapters {
		t.Error("Expected flat document, got chaptered")
	}
}

func TestCreateChapteredDocument(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Test creating a chaptered document
	docID, err := storage.CreateDocument("Book Title", true, "Book Author")
	if err != nil {
		t.Fatalf("Failed to create chaptered document: %v", err)
	}
	
	// Verify document was created
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	
	if !doc.HasChapters {
		t.Error("Expected chaptered document, got flat")
	}
}

func TestAddHeadingBlock(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document
	docID, err := storage.CreateDocument("Test Doc", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add heading block
	heading := &blocks.HeadingBlock{
		Level: 1,
		Text:  "Introduction",
	}
	
	err = storage.AddBlock(docID, "", heading, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add heading block: %v", err)
	}
	
	// Verify block was added
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(doc.Blocks))
	}
	
	if doc.Blocks[0].Type != blocks.TypeHeading {
		t.Errorf("Expected heading block, got %s", doc.Blocks[0].Type)
	}
	
	// Load and verify block content
	block, err := storage.LoadBlock(docID, doc.Blocks[0])
	if err != nil {
		t.Fatal(err)
	}
	
	h, ok := block.(*blocks.HeadingBlock)
	if !ok {
		t.Fatal("Expected HeadingBlock type")
	}
	
	if h.Text != "Introduction" {
		t.Errorf("Expected text 'Introduction', got '%s'", h.Text)
	}
}

func TestAddMarkdownBlock(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document
	docID, err := storage.CreateDocument("Test Doc", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add markdown block
	markdown := &blocks.MarkdownBlock{
		Content: "This is some **markdown** content with *emphasis*.",
	}
	
	err = storage.AddBlock(docID, "", markdown, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add markdown block: %v", err)
	}
	
	// Verify block was added
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(doc.Blocks))
	}
	
	// Load and verify block content
	block, err := storage.LoadBlock(docID, doc.Blocks[0])
	if err != nil {
		t.Fatal(err)
	}
	
	md, ok := block.(*blocks.MarkdownBlock)
	if !ok {
		t.Fatal("Expected MarkdownBlock type")
	}
	
	if md.Content != markdown.Content {
		t.Errorf("Content mismatch")
	}
}

func TestAddTableBlock(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document
	docID, err := storage.CreateDocument("Test Doc", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add table block
	table := &blocks.TableBlock{
		Headers: []string{"Name", "Age", "City"},
		Rows: [][]string{
			{"Alice", "30", "New York"},
			{"Bob", "25", "San Francisco"},
		},
	}
	
	err = storage.AddBlock(docID, "", table, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add table block: %v", err)
	}
	
	// Load and verify block content
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	block, err := storage.LoadBlock(docID, doc.Blocks[0])
	if err != nil {
		t.Fatal(err)
	}
	
	tbl, ok := block.(*blocks.TableBlock)
	if !ok {
		t.Fatal("Expected TableBlock type")
	}
	
	if len(tbl.Headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(tbl.Headers))
	}
	
	if len(tbl.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(tbl.Rows))
	}
}

func TestAddChapter(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create chaptered document
	docID, err := storage.CreateDocument("Book", true, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add chapter
	chapterID, err := storage.AddChapter(docID, "Chapter One", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add chapter: %v", err)
	}
	
	if chapterID == "" {
		t.Error("Expected chapter ID, got empty string")
	}
	
	// Verify chapter was added
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Chapters) != 1 {
		t.Errorf("Expected 1 chapter, got %d", len(doc.Chapters))
	}
	
	if doc.Chapters[0].Title != "Chapter One" {
		t.Errorf("Expected title 'Chapter One', got '%s'", doc.Chapters[0].Title)
	}
	
	// Get chapter
	chapter, err := storage.GetChapter(docID, chapterID)
	if err != nil {
		t.Fatalf("Failed to get chapter: %v", err)
	}
	
	if chapter.Title != "Chapter One" {
		t.Errorf("Expected chapter title 'Chapter One', got '%s'", chapter.Title)
	}
}

func TestBlockPositioning(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document
	docID, err := storage.CreateDocument("Test Doc", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add first block
	block1 := &blocks.HeadingBlock{Level: 1, Text: "First"}
	err = storage.AddBlock(docID, "", block1, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	// Add block at start
	block2 := &blocks.HeadingBlock{Level: 1, Text: "Start"}
	err = storage.AddBlock(docID, "", block2, document.Position{Type: document.PositionStart})
	if err != nil {
		t.Fatal(err)
	}
	
	// Get document to check order
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(doc.Blocks))
	}
	
	// Load first block (should be "Start")
	firstBlock, err := storage.LoadBlock(docID, doc.Blocks[0])
	if err != nil {
		t.Fatal(err)
	}
	
	h1, ok := firstBlock.(*blocks.HeadingBlock)
	if !ok {
		t.Fatal("Expected HeadingBlock")
	}
	
	if h1.Text != "Start" {
		t.Errorf("Expected first block to be 'Start', got '%s'", h1.Text)
	}
	
	// Add block after first block
	block3 := &blocks.HeadingBlock{Level: 1, Text: "Middle"}
	err = storage.AddBlock(docID, "", block3, document.Position{
		Type:    document.PositionAfter,
		BlockID: doc.Blocks[0].ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify order
	doc, err = storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(doc.Blocks))
	}
	
	// Load middle block
	middleBlock, err := storage.LoadBlock(docID, doc.Blocks[1])
	if err != nil {
		t.Fatal(err)
	}
	
	h2, ok := middleBlock.(*blocks.HeadingBlock)
	if !ok {
		t.Fatal("Expected HeadingBlock")
	}
	
	if h2.Text != "Middle" {
		t.Errorf("Expected middle block to be 'Middle', got '%s'", h2.Text)
	}
}

func TestListDocuments(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create multiple documents
	_, err := storage.CreateDocument("Doc One", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	_, err = storage.CreateDocument("Doc Two", true, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// List documents
	docIDs, err := storage.ListDocuments()
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}
	
	if len(docIDs) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(docIDs))
	}
}

func TestDeleteDocument(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document
	docID, err := storage.CreateDocument("To Delete", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify it exists
	_, err = storage.GetDocument(docID)
	if err != nil {
		t.Fatal("Document should exist")
	}
	
	// Delete document
	err = storage.DeleteDocument(docID)
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}
	
	// Verify it's gone
	_, err = storage.GetDocument(docID)
	if err == nil {
		t.Error("Document should not exist after deletion")
	}
}