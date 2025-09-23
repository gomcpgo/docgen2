package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestUpdateChapter(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create a chaptered document
	docID, err := storage.CreateDocument("Chapter Test Doc", true, "Test Author")
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	
	// Add a chapter
	chapterID, err := storage.AddChapter(docID, "Original Title", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add chapter: %v", err)
	}
	
	// Update the chapter title
	newTitle := "Updated Chapter Title"
	err = storage.UpdateChapter(docID, chapterID, newTitle)
	if err != nil {
		t.Fatalf("Failed to update chapter: %v", err)
	}
	
	// Verify the chapter was updated in the manifest
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	
	found := false
	for _, chapterRef := range doc.Chapters {
		if chapterRef.ID == chapterID && chapterRef.Title == newTitle {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Chapter title was not updated in document manifest")
	}
	
	// Verify the chapter file was updated
	chapter, err := storage.GetChapter(docID, chapterID)
	if err != nil {
		t.Fatalf("Failed to get chapter: %v", err)
	}
	
	if chapter.Title != newTitle {
		t.Errorf("Expected chapter title %s, got %s", newTitle, chapter.Title)
	}
}

func TestUpdateChapterNonChapteredDoc(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create a flat document
	docID, err := storage.CreateDocument("Flat Doc", false, "Test Author")
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	
	// Try to update a chapter in a non-chaptered document
	err = storage.UpdateChapter(docID, "ch-001", "New Title")
	if err == nil {
		t.Error("Expected error when updating chapter in non-chaptered document")
	}
}

func TestDeleteChapter(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create a chaptered document
	docID, err := storage.CreateDocument("Chapter Test Doc", true, "Test Author")
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	
	// Add two chapters
	chapter1ID, err := storage.AddChapter(docID, "Chapter 1", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add chapter 1: %v", err)
	}
	
	chapter2ID, err := storage.AddChapter(docID, "Chapter 2", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add chapter 2: %v", err)
	}
	
	// Add a block to the first chapter
	headingBlock := &blocks.HeadingBlock{
		Level: 1,
		Text:  "Test Heading in Chapter 1",
	}
	
	err = storage.AddBlock(docID, chapter1ID, headingBlock, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add block to chapter: %v", err)
	}
	
	// Delete the first chapter
	err = storage.DeleteChapter(docID, chapter1ID)
	if err != nil {
		t.Fatalf("Failed to delete chapter: %v", err)
	}
	
	// Verify the chapter was removed from the manifest
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	
	if len(doc.Chapters) != 1 {
		t.Errorf("Expected 1 chapter after deletion, got %d", len(doc.Chapters))
	}
	
	if doc.Chapters[0].ID != chapter2ID {
		t.Errorf("Expected remaining chapter to be %s, got %s", chapter2ID, doc.Chapters[0].ID)
	}
	
	// Verify the chapter file is gone (should return error)
	_, err = storage.GetChapter(docID, chapter1ID)
	if err == nil {
		t.Error("Expected error when getting deleted chapter")
	}
}

func TestMoveChapter(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create a chaptered document
	docID, err := storage.CreateDocument("Chapter Test Doc", true, "Test Author")
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	
	// Add three chapters
	chapter1ID, err := storage.AddChapter(docID, "Chapter 1", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add chapter 1: %v", err)
	}
	
	chapter2ID, err := storage.AddChapter(docID, "Chapter 2", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add chapter 2: %v", err)
	}
	
	chapter3ID, err := storage.AddChapter(docID, "Chapter 3", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add chapter 3: %v", err)
	}
	
	// Move chapter 3 to the start
	err = storage.MoveChapter(docID, chapter3ID, document.Position{Type: document.PositionStart})
	if err != nil {
		t.Fatalf("Failed to move chapter: %v", err)
	}
	
	// Verify the new order
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	
	expectedOrder := []string{chapter3ID, chapter1ID, chapter2ID}
	for i, expected := range expectedOrder {
		if doc.Chapters[i].ID != expected {
			t.Errorf("Expected chapter at position %d to be %s, got %s", i, expected, doc.Chapters[i].ID)
		}
	}
	
	// Move chapter 1 after chapter 3
	err = storage.MoveChapter(docID, chapter1ID, document.Position{
		Type:    document.PositionAfter,
		BlockID: chapter3ID,
	})
	if err != nil {
		t.Fatalf("Failed to move chapter after another: %v", err)
	}
	
	// Verify the new order
	doc, err = storage.GetDocument(docID)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	
	expectedOrder = []string{chapter3ID, chapter1ID, chapter2ID}
	for i, expected := range expectedOrder {
		if doc.Chapters[i].ID != expected {
			t.Errorf("Expected chapter at position %d to be %s, got %s", i, expected, doc.Chapters[i].ID)
		}
	}
}

func TestAddImageBlock(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document
	docID, err := storage.CreateDocument("Image Test", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a test image file
	testImagePath := filepath.Join(os.TempDir(), "test-image.png")
	testImageContent := []byte("fake png content")
	err = os.WriteFile(testImagePath, testImageContent, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testImagePath)
	
	// Copy image to assets and get relative path
	relativePath, err := storage.CopyImageToAssets(docID, testImagePath)
	if err != nil {
		t.Fatalf("Failed to copy image to assets: %v", err)
	}
	
	// Verify relative path format
	if !strings.HasPrefix(relativePath, "assets/") {
		t.Errorf("Expected relative path to start with 'assets/', got '%s'", relativePath)
	}
	
	// Create image block with relative path
	image := &blocks.ImageBlock{
		Path:    relativePath,
		Caption: "Test image caption",
		AltText: "Test alt text",
	}
	
	err = storage.AddBlock(docID, "", image, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add image block: %v", err)
	}
	
	// Verify block was added
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(doc.Blocks))
	}
	
	if doc.Blocks[0].Type != blocks.TypeImage {
		t.Errorf("Expected image block, got %s", doc.Blocks[0].Type)
	}
	
	// Load and verify block content
	block, err := storage.LoadBlock(docID, doc.Blocks[0])
	if err != nil {
		t.Fatal(err)
	}
	
	img, ok := block.(*blocks.ImageBlock)
	if !ok {
		t.Fatal("Expected ImageBlock type")
	}
	
	// Verify path is stored as relative
	if img.Path != relativePath {
		t.Errorf("Expected path '%s', got '%s'", relativePath, img.Path)
	}
	
	if img.Caption != "Test image caption" {
		t.Errorf("Caption mismatch")
	}
	
	if img.AltText != "Test alt text" {
		t.Errorf("Alt text mismatch")
	}
	
	// Verify the actual image file was copied
	fullImagePath := filepath.Join(storage.config.GetDocumentFolder(docID), relativePath)
	if _, err := os.Stat(fullImagePath); os.IsNotExist(err) {
		t.Error("Image file was not copied to assets folder")
	}
}

func TestImageBlockLegacyAbsolutePath(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document
	docID, err := storage.CreateDocument("Legacy Image Test", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	// Create image block with absolute path (legacy format)
	absolutePath := "/some/absolute/path/to/image.png"
	image := &blocks.ImageBlock{
		Path:    absolutePath,
		Caption: "Legacy image",
		AltText: "Legacy alt text",
	}
	
	// Manually add the block to simulate legacy data
	err = storage.AddBlock(docID, "", image, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add legacy image block: %v", err)
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
	
	img, ok := block.(*blocks.ImageBlock)
	if !ok {
		t.Fatal("Expected ImageBlock type")
	}
	
	// Verify absolute path is preserved for legacy compatibility
	if img.Path != absolutePath {
		t.Errorf("Expected legacy absolute path '%s', got '%s'", absolutePath, img.Path)
	}
}

func TestMixedDocumentBlocksAndChapters(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create a chaptered document that will have both document blocks and chapters
	docID, err := storage.CreateDocument("Mixed Document Test", true, "Test Author")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add document-level blocks first
	docHeading := &blocks.HeadingBlock{
		Level: 1,
		Text:  "Document Introduction",
	}
	err = storage.AddBlock(docID, "", docHeading, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add document-level heading: %v", err)
	}
	
	docMarkdown := &blocks.MarkdownBlock{
		Content: "This is document-level content that appears before chapters.",
	}
	err = storage.AddBlock(docID, "", docMarkdown, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add document-level markdown: %v", err)
	}
	
	// Add a chapter
	chapterID, err := storage.AddChapter(docID, "Chapter One", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	// Add blocks to the chapter
	chapterHeading := &blocks.HeadingBlock{
		Level: 2,
		Text:  "Chapter One Content",
	}
	err = storage.AddBlock(docID, chapterID, chapterHeading, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add chapter heading: %v", err)
	}
	
	chapterMarkdown := &blocks.MarkdownBlock{
		Content: "This is chapter-specific content.",
	}
	err = storage.AddBlock(docID, chapterID, chapterMarkdown, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatalf("Failed to add chapter markdown: %v", err)
	}
	
	// Verify the document structure
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	// Should have both document-level blocks and chapters
	if len(doc.Blocks) != 2 {
		t.Errorf("Expected 2 document-level blocks, got %d", len(doc.Blocks))
	}
	
	if len(doc.Chapters) != 1 {
		t.Errorf("Expected 1 chapter, got %d", len(doc.Chapters))
	}
	
	if !doc.HasChapters {
		t.Error("Document should be marked as having chapters")
	}
	
	// Verify chapter has blocks
	chapter, err := storage.GetChapter(docID, chapterID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(chapter.Blocks) != 2 {
		t.Errorf("Expected 2 chapter blocks, got %d", len(chapter.Blocks))
	}
	
	// Verify both blocks/ and chapters/ folders exist
	docPath := storage.config.GetDocumentFolder(docID)
	blocksPath := filepath.Join(docPath, "blocks")
	chaptersPath := filepath.Join(docPath, "chapters")
	
	if _, err := os.Stat(blocksPath); os.IsNotExist(err) {
		t.Error("blocks/ folder should exist for mixed documents")
	}
	
	if _, err := os.Stat(chaptersPath); os.IsNotExist(err) {
		t.Error("chapters/ folder should exist for mixed documents")
	}
}

func TestDocumentWithEmptyChapters(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create a chaptered document
	docID, err := storage.CreateDocument("Empty Chapters Test", true, "Test Author")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add document-level blocks
	heading := &blocks.HeadingBlock{
		Level: 1,
		Text:  "Document with Empty Chapters",
	}
	err = storage.AddBlock(docID, "", heading, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	// Add a chapter but don't add any blocks to it
	chapterID, err := storage.AddChapter(docID, "Empty Chapter", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify document structure
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != 1 {
		t.Errorf("Expected 1 document-level block, got %d", len(doc.Blocks))
	}
	
	if len(doc.Chapters) != 1 {
		t.Errorf("Expected 1 chapter, got %d", len(doc.Chapters))
	}
	
	// Verify chapter exists but is empty
	chapter, err := storage.GetChapter(docID, chapterID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(chapter.Blocks) != 0 {
		t.Errorf("Expected empty chapter, got %d blocks", len(chapter.Blocks))
	}
}

func TestPureBlocksOnlyDocument(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create a flat document (no chapters)
	docID, err := storage.CreateDocument("Blocks Only Test", false, "Test Author")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add document-level blocks
	heading := &blocks.HeadingBlock{
		Level: 1,
		Text:  "Blocks Only Document",
	}
	err = storage.AddBlock(docID, "", heading, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	markdown := &blocks.MarkdownBlock{
		Content: "This document only has blocks, no chapters.",
	}
	err = storage.AddBlock(docID, "", markdown, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify document structure
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != 2 {
		t.Errorf("Expected 2 document-level blocks, got %d", len(doc.Blocks))
	}
	
	if len(doc.Chapters) != 0 {
		t.Errorf("Expected 0 chapters, got %d", len(doc.Chapters))
	}
	
	if doc.HasChapters {
		t.Error("Document should not be marked as having chapters")
	}
}

func TestPureChaptersOnlyDocument(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create a chaptered document
	docID, err := storage.CreateDocument("Chapters Only Test", true, "Test Author")
	if err != nil {
		t.Fatal(err)
	}
	
	// Add chapters but no document-level blocks
	chapter1ID, err := storage.AddChapter(docID, "Chapter One", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	chapter2ID, err := storage.AddChapter(docID, "Chapter Two", document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	// Add blocks to chapters
	heading1 := &blocks.HeadingBlock{
		Level: 1,
		Text:  "Chapter One Content",
	}
	err = storage.AddBlock(docID, chapter1ID, heading1, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	heading2 := &blocks.HeadingBlock{
		Level: 1,
		Text:  "Chapter Two Content",
	}
	err = storage.AddBlock(docID, chapter2ID, heading2, document.Position{Type: document.PositionEnd})
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify document structure
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != 0 {
		t.Errorf("Expected 0 document-level blocks, got %d", len(doc.Blocks))
	}
	
	if len(doc.Chapters) != 2 {
		t.Errorf("Expected 2 chapters, got %d", len(doc.Chapters))
	}
	
	if !doc.HasChapters {
		t.Error("Document should be marked as having chapters")
	}
	
	// Verify chapter blocks
	chapter1, err := storage.GetChapter(docID, chapter1ID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(chapter1.Blocks) != 1 {
		t.Errorf("Expected 1 block in chapter 1, got %d", len(chapter1.Blocks))
	}
	
	chapter2, err := storage.GetChapter(docID, chapter2ID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(chapter2.Blocks) != 1 {
		t.Errorf("Expected 1 block in chapter 2, got %d", len(chapter2.Blocks))
	}
}

func TestImagePathConversion(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()
	
	// Create document
	docID, err := storage.CreateDocument("Path Conversion Test", false, "")
	if err != nil {
		t.Fatal(err)
	}
	
	testCases := []struct {
		name        string
		storedPath  string
		description string
	}{
		{
			name:        "relative_path",
			storedPath:  "assets/test-image-001.png",
			description: "Relative path should be stored as-is",
		},
		{
			name:        "absolute_path_legacy",
			storedPath:  "/Users/test/absolute/path/image.png",
			description: "Absolute path should be preserved for legacy compatibility",
		},
	}
	
	var blockIDs []string
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create image block with the test path
			image := &blocks.ImageBlock{
				Path:    tc.storedPath,
				Caption: fmt.Sprintf("Test image for %s", tc.name),
				AltText: fmt.Sprintf("Alt text for %s", tc.name),
			}
			
			err := storage.AddBlock(docID, "", image, document.Position{Type: document.PositionEnd})
			if err != nil {
				t.Fatalf("Failed to add image block: %v", err)
			}
			
			// Get the document to find the block ID
			doc, err := storage.GetDocument(docID)
			if err != nil {
				t.Fatal(err)
			}
			
			// Find the latest block (the one we just added)
			latestBlockRef := doc.Blocks[len(doc.Blocks)-1]
			blockIDs = append(blockIDs, latestBlockRef.ID)
			
			// Load and verify the block
			block, err := storage.LoadBlock(docID, latestBlockRef)
			if err != nil {
				t.Fatal(err)
			}
			
			img, ok := block.(*blocks.ImageBlock)
			if !ok {
				t.Fatal("Expected ImageBlock type")
			}
			
			// Verify the path is stored exactly as provided
			if img.Path != tc.storedPath {
				t.Errorf("Expected stored path '%s', got '%s'", tc.storedPath, img.Path)
			}
			
			// Test that relative paths start with assets/ (indicating they're relative)
			if strings.HasPrefix(tc.storedPath, "assets/") {
				if !strings.HasPrefix(img.Path, "assets/") {
					t.Errorf("Relative path should start with 'assets/', got '%s'", img.Path)
				}
			}
			
			// Test that absolute paths start with / (indicating they're absolute)
			if strings.HasPrefix(tc.storedPath, "/") {
				if !strings.HasPrefix(img.Path, "/") {
					t.Errorf("Absolute path should start with '/', got '%s'", img.Path)
				}
			}
		})
	}
	
	// Verify we can load all blocks by ID
	if len(blockIDs) != len(testCases) {
		t.Errorf("Expected %d block IDs, got %d", len(testCases), len(blockIDs))
	}
	
	// Test that we can retrieve all blocks and their paths are preserved
	doc, err := storage.GetDocument(docID)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(doc.Blocks) != len(testCases) {
		t.Errorf("Expected %d blocks in document, got %d", len(testCases), len(doc.Blocks))
	}
}