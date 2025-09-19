package storage

import (
	"fmt"
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

func TestUpdateBlock(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document
	docID, err := storage.CreateDocument("Update Test", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add a heading block
	heading := &blocks.HeadingBlock{
		Level: 1,
		Text:  "Original Title",
	}
	
	err = storage.AddBlock(docID, "", heading, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	// Get the document to find the block ID
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	blockID := doc.Blocks[0].ID
	
	// Update the block
	updatedHeading := &blocks.HeadingBlock{
		Level: 2,
		Text:  "Updated Title",
	}
	
	err = storage.UpdateBlock(docID, blockID, updatedHeading)
	if err != nil {
		t.Fatalf("Failed to update block: %v", err)
	}
	
	// Load and verify the updated block
	blockRef := doc.Blocks[0]
	block, err := storage.LoadBlock(docID, blockRef)
	if err != nil {
		t.Fatal(err)
	}
	
	h, ok := block.(*blocks.HeadingBlock)
	if !ok {
		t.Fatal("Expected HeadingBlock")
	}
	
	if h.Text != "Updated Title" {
		t.Errorf("Expected 'Updated Title', got '%s'", h.Text)
	}
	
	if h.Level != 2 {
		t.Errorf("Expected level 2, got %d", h.Level)
	}
}

func TestUpdateBlockWrongType(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document with heading
	docID, err := storage.CreateDocument("Type Test", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	heading := &blocks.HeadingBlock{
		Level: 1,
		Text:  "Title",
	}
	
	err = storage.AddBlock(docID, "", heading, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	blockID := doc.Blocks[0].ID
	
	// Try to update with different type
	markdown := &blocks.MarkdownBlock{
		Content: "Some content",
	}
	
	err = storage.UpdateBlock(docID, blockID, markdown)
	if err == nil {
		t.Error("Expected error when changing block type")
	}
}

func TestDeleteBlock(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document with multiple blocks
	docID, err := storage.CreateDocument("Delete Test", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add three blocks
	for i := 1; i <= 3; i++ {
		heading := &blocks.HeadingBlock{
			Level: 1,
			Text:  fmt.Sprintf("Heading %d", i),
		}
		err = storage.AddBlock(docID, "", heading, document.Position{Type: document.PositionEnd})
		if err != nil {
			t.Fatal(err)
		}
	}
	
	// Get document to check blocks
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(doc.Blocks))
	}
	
	// Delete the middle block
	middleBlockID := doc.Blocks[1].ID
	
	err = storage.DeleteBlock(docID, middleBlockID)
	if err != nil {
		t.Fatalf("Failed to delete block: %v", err)
	}
	
	// Verify block is gone
	doc, err = storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != 2 {
		t.Errorf("Expected 2 blocks after deletion, got %d", len(doc.Blocks))
	}
	
	// Verify the remaining blocks are correct
	firstBlock, err := storage.LoadBlock(docID, doc.Blocks[0])
	if err != nil {
		t.Fatal(err)
	}
	
	h1, ok := firstBlock.(*blocks.HeadingBlock)
	if !ok || h1.Text != "Heading 1" {
		t.Error("First block should be 'Heading 1'")
	}
	
	lastBlock, err := storage.LoadBlock(docID, doc.Blocks[1])
	if err != nil {
		t.Fatal(err)
	}
	
	h3, ok := lastBlock.(*blocks.HeadingBlock)
	if !ok || h3.Text != "Heading 3" {
		t.Error("Last block should be 'Heading 3'")
	}
}

func TestMoveBlock(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document with three blocks
	docID, err := storage.CreateDocument("Move Test", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add three blocks: A, B, C
	for _, letter := range []string{"A", "B", "C"} {
		heading := &blocks.HeadingBlock{
			Level: 1,
			Text:  "Block " + letter,
		}
		err = storage.AddBlock(docID, "", heading, document.Position{Type: document.PositionEnd})
		if err != nil {
			t.Fatal(err)
		}
	}
	
	// Get initial order
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	// Should be: A, B, C
	blockA_ID := doc.Blocks[0].ID
	blockB_ID := doc.Blocks[1].ID
	blockC_ID := doc.Blocks[2].ID
	
	// Move B to start (should be: B, A, C)
	err = storage.MoveBlock(docID, blockB_ID, document.Position{Type: document.PositionStart})
	if err != nil {
		t.Fatalf("Failed to move block: %v", err)
	}
	
	// Verify new order
	doc, err = storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	expectedOrder := []string{blockB_ID, blockA_ID, blockC_ID}
	for i, expectedID := range expectedOrder {
		if doc.Blocks[i].ID != expectedID {
			t.Errorf("Position %d: expected %s, got %s", i, expectedID, doc.Blocks[i].ID)
		}
	}
	
	// Move C after A (should be: B, A, C)
	err = storage.MoveBlock(docID, blockC_ID, document.Position{
		Type:    document.PositionAfter,
		BlockID: blockA_ID,
	})
	if err != nil {
		t.Fatalf("Failed to move block: %v", err)
	}
	
	// Verify final order is still B, A, C (C was already after A)
	doc, err = storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	finalOrder := []string{blockB_ID, blockA_ID, blockC_ID}
	for i, expectedID := range finalOrder {
		if doc.Blocks[i].ID != expectedID {
			t.Errorf("Final position %d: expected %s, got %s", i, expectedID, doc.Blocks[i].ID)
		}
	}
}

func TestMoveBlockInChapter(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create chaptered document
	docID, err := storage.CreateDocument("Chapter Move Test", true, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add chapter
	chapterID, err := storage.AddChapter(docID, "Test Chapter", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	// Add blocks to chapter
	for _, letter := range []string{"X", "Y"} {
		heading := &blocks.HeadingBlock{
			Level: 1,
			Text:  "Block " + letter,
		}
		err = storage.AddBlock(docID, chapterID, heading, document.Position{Type: document.PositionEnd})
		if err != nil {
			t.Fatal(err)
		}
	}
	
	// Get chapter blocks
	chapter, err := storage.GetChapter(docID, chapterID)
	if err != nil {
		t.Fatal(err)
	}
	
	blockX_ID := chapter.Blocks[0].ID
	blockY_ID := chapter.Blocks[1].ID
	
	// Move Y to start
	err = storage.MoveBlock(docID, blockY_ID, document.Position{Type: document.PositionStart})
	if err != nil {
		t.Fatalf("Failed to move block in chapter: %v", err)
	}
	
	// Verify order changed
	chapter, err = storage.GetChapter(docID, chapterID)
	if err != nil {
		t.Fatal(err)
	}
	
	if chapter.Blocks[0].ID != blockY_ID {
		t.Error("Block Y should be first")
	}
	
	if chapter.Blocks[1].ID != blockX_ID {
		t.Error("Block X should be second")
	}
}

func TestFindBlockLocation(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create flat document
	docID, err := storage.CreateDocument("Location Test", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add block
	heading := &blocks.HeadingBlock{
		Level: 1,
		Text:  "Test Block",
	}
	
	err = storage.AddBlock(docID, "", heading, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	// Get block ID
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	blockID := doc.Blocks[0].ID
	
	// Find location
	chapterID, index, err := storage.FindBlockLocation(docID, blockID)
	if err != nil {
		t.Fatalf("Failed to find block location: %v", err)
	}
	
	if chapterID != "" {
		t.Error("Expected empty chapter ID for flat document")
	}
	
	if index != 0 {
		t.Errorf("Expected index 0, got %d", index)
	}
}