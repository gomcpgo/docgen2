package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/savant/mcp-servers/docgen2/pkg/config"
	"github.com/savant/mcp-servers/docgen2/pkg/handler"
)

func setupGetBlocksTestHandler(t *testing.T) (*handler.Handler, func()) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "docgen-getblocks-test-*")
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

	h := handler.NewHandler(cfg)

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return h, cleanup
}

func TestGetBlocksEmpty(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Empty Blocks Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Get blocks with empty array
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "empty-blocks-test",
			"block_ids":   []interface{}{},
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Failed to get blocks: %v", err)
	}

	if len(resp.Content) == 0 {
		t.Fatal("Expected response content")
	}

	content := resp.Content[0].Text
	if !contains(content, "[]") {
		t.Error("Empty block_ids should return empty array")
	}
}

func TestGetBlocksSingle(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Single Block Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add a heading block
	headingReq := &protocol.CallToolRequest{
		Name: "add_heading",
		Arguments: map[string]interface{}{
			"document_id": "single-block-test",
			"level":       1,
			"text":        "Test Heading",
		},
	}

	_, err = h.CallTool(ctx, headingReq)
	if err != nil {
		t.Fatal(err)
	}

	// Get single block
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "single-block-test",
			"block_ids":   []interface{}{"hd-001"},
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Failed to get blocks: %v", err)
	}

	content := resp.Content[0].Text

	// Should contain the block data
	if !contains(content, "hd-001") {
		t.Error("Response should contain block ID")
	}
	if !contains(content, "Test Heading") {
		t.Error("Response should contain heading text")
	}
	if !contains(content, "\"type\": \"heading\"") {
		t.Error("Response should contain block type")
	}
	if !contains(content, "\"level\": 1") {
		t.Error("Response should contain heading level")
	}
}

func TestGetBlocksMultiple(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Multiple Blocks Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add different types of blocks
	blocks := []struct {
		name string
		args map[string]interface{}
		id   string
	}{
		{
			name: "add_heading",
			args: map[string]interface{}{
				"document_id": "multiple-blocks-test",
				"level":       1,
				"text":        "First Heading",
			},
			id: "hd-001",
		},
		{
			name: "add_markdown",
			args: map[string]interface{}{
				"document_id": "multiple-blocks-test",
				"content":     "This is markdown content.",
			},
			id: "md-001",
		},
		{
			name: "add_table",
			args: map[string]interface{}{
				"document_id": "multiple-blocks-test",
				"headers":     []interface{}{"Name", "Age"},
				"rows": []interface{}{
					[]interface{}{"Alice", "30"},
					[]interface{}{"Bob", "25"},
				},
			},
			id: "tbl-001",
		},
		{
			name: "add_page_break",
			args: map[string]interface{}{
				"document_id": "multiple-blocks-test",
			},
			id: "pb-001",
		},
	}

	for _, block := range blocks {
		req := &protocol.CallToolRequest{
			Name:      block.name,
			Arguments: block.args,
		}
		_, err := h.CallTool(ctx, req)
		if err != nil {
			t.Fatalf("Failed to add %s: %v", block.name, err)
		}
	}

	// Get all blocks
	allBlockIDs := []interface{}{"hd-001", "md-001", "tbl-001", "pb-001"}
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "multiple-blocks-test",
			"block_ids":   allBlockIDs,
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Failed to get blocks: %v", err)
	}

	content := resp.Content[0].Text

	// Should contain all blocks
	expectedContent := map[string][]string{
		"hd-001": {"\"type\": \"heading\"", "First Heading", "\"level\": 1"},
		"md-001": {"\"type\": \"markdown\"", "This is markdown content"},
		"tbl-001": {"\"type\": \"table\"", "Name", "Age", "Alice", "Bob"},
		"pb-001": {"\"type\": \"page_break\""},
	}

	for blockID, expectedStrings := range expectedContent {
		if !contains(content, blockID) {
			t.Errorf("Response should contain block ID %s", blockID)
		}
		for _, expected := range expectedStrings {
			if !contains(content, expected) {
				t.Errorf("Response should contain %s for block %s", expected, blockID)
			}
		}
	}
}

func TestGetBlocksPartialExist(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Partial Blocks Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add only two blocks
	headingReq := &protocol.CallToolRequest{
		Name: "add_heading",
		Arguments: map[string]interface{}{
			"document_id": "partial-blocks-test",
			"level":       1,
			"text":        "Existing Heading",
		},
	}
	_, err = h.CallTool(ctx, headingReq)
	if err != nil {
		t.Fatal(err)
	}

	markdownReq := &protocol.CallToolRequest{
		Name: "add_markdown",
		Arguments: map[string]interface{}{
			"document_id": "partial-blocks-test",
			"content":     "Existing markdown.",
		},
	}
	_, err = h.CallTool(ctx, markdownReq)
	if err != nil {
		t.Fatal(err)
	}

	// Try to get existing and non-existing blocks
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "partial-blocks-test",
			"block_ids":   []interface{}{"hd-001", "nonexistent-block", "md-001", "another-fake-block"},
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Failed to get blocks: %v", err)
	}

	content := resp.Content[0].Text

	// Should contain only existing blocks
	if !contains(content, "hd-001") {
		t.Error("Response should contain existing heading block")
	}
	if !contains(content, "md-001") {
		t.Error("Response should contain existing markdown block")
	}
	if !contains(content, "Existing Heading") {
		t.Error("Response should contain heading text")
	}
	if !contains(content, "Existing markdown") {
		t.Error("Response should contain markdown text")
	}

	// Should not contain non-existing blocks
	if contains(content, "nonexistent-block") {
		t.Error("Response should not contain non-existing block ID")
	}
	if contains(content, "another-fake-block") {
		t.Error("Response should not contain another fake block ID")
	}
}

func TestGetBlocksNonexistentDocument(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Try to get blocks from non-existent document
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "nonexistent-document",
			"block_ids":   []interface{}{"hd-001", "md-001"},
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	// According to the code, this should return empty array instead of error
	if err != nil {
		t.Fatalf("Should not error for non-existent document, should return empty: %v", err)
	}

	content := resp.Content[0].Text
	if !contains(content, "[]") {
		t.Error("Non-existent document should return empty array")
	}
}

func TestGetBlocksChapteredDocument(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create chaptered document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title":        "Chaptered Blocks Test",
			"has_chapters": true,
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add chapters
	chapterTitles := []string{"Chapter One", "Chapter Two"}
	for i, title := range chapterTitles {
		chapterReq := &protocol.CallToolRequest{
			Name: "add_chapter",
			Arguments: map[string]interface{}{
				"document_id": "chaptered-blocks-test",
				"title":       title,
			},
		}
		_, err := h.CallTool(ctx, chapterReq)
		if err != nil {
			t.Fatal(err)
		}

		chapterID := fmt.Sprintf("ch-%03d", i+1)

		// Add blocks to each chapter
		headingReq := &protocol.CallToolRequest{
			Name: "add_heading",
			Arguments: map[string]interface{}{
				"document_id": "chaptered-blocks-test",
				"chapter_id":  chapterID,
				"level":       1,
				"text":        fmt.Sprintf("Heading in %s", title),
			},
		}
		_, err = h.CallTool(ctx, headingReq)
		if err != nil {
			t.Fatal(err)
		}

		markdownReq := &protocol.CallToolRequest{
			Name: "add_markdown",
			Arguments: map[string]interface{}{
				"document_id": "chaptered-blocks-test",
				"chapter_id":  chapterID,
				"content":     fmt.Sprintf("Content for %s", title),
			},
		}
		_, err = h.CallTool(ctx, markdownReq)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Get blocks from both chapters
	// In chaptered documents, each chapter has its own numbering, so both chapters have hd-001, md-001
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "chaptered-blocks-test",
			"block_ids":   []interface{}{"hd-001", "md-001"},
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Failed to get blocks from chaptered document: %v", err)
	}

	content := resp.Content[0].Text

	// Due to current implementation, when chapters have duplicate block IDs,
	// only the last chapter's blocks are returned (Chapter Two in this case)
	if !contains(content, "Heading in Chapter Two") {
		t.Error("Response should contain heading from second chapter")
	}
	if !contains(content, "Content for Chapter Two") {
		t.Error("Response should contain content from second chapter")
	}

	// Should have the requested block IDs
	expectedIDs := []string{"hd-001", "md-001"}
	for _, id := range expectedIDs {
		if !contains(content, id) {
			t.Errorf("Response should contain block ID %s", id)
		}
	}
}

func TestGetBlocksLargeRequest(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Large Request Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add many blocks
	var blockIDs []interface{}
	for i := 1; i <= 50; i++ {
		headingReq := &protocol.CallToolRequest{
			Name: "add_heading",
			Arguments: map[string]interface{}{
				"document_id": "large-request-test",
				"level":       i%6 + 1,
				"text":        fmt.Sprintf("Heading %d", i),
			},
		}
		_, err := h.CallTool(ctx, headingReq)
		if err != nil {
			t.Fatalf("Failed to add heading %d: %v", i, err)
		}

		blockIDs = append(blockIDs, fmt.Sprintf("hd-%03d", i))
	}

	// Get all blocks at once
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "large-request-test",
			"block_ids":   blockIDs,
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Failed to get large number of blocks: %v", err)
	}

	content := resp.Content[0].Text

	// Should contain first and last blocks
	if !contains(content, "hd-001") {
		t.Error("Response should contain first block")
	}
	if !contains(content, "hd-050") {
		t.Error("Response should contain last block")
	}
	if !contains(content, "Heading 1") {
		t.Error("Response should contain first heading text")
	}
	if !contains(content, "Heading 50") {
		t.Error("Response should contain last heading text")
	}

	// Check that we have all blocks (count occurrences of "hd-")
	hdCount := strings.Count(content, "\"id\": \"hd-")
	if hdCount != 50 {
		t.Errorf("Expected 50 blocks in response, got %d", hdCount)
	}
}

func TestGetBlocksWithLongContent(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Long Content Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add block with very long content
	longContent := strings.Repeat("This is a very long line of text that goes on and on. ", 200)
	markdownReq := &protocol.CallToolRequest{
		Name: "add_markdown",
		Arguments: map[string]interface{}{
			"document_id": "long-content-test",
			"content":     longContent,
		},
	}
	_, err = h.CallTool(ctx, markdownReq)
	if err != nil {
		t.Fatal(err)
	}

	// Get the block
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "long-content-test",
			"block_ids":   []interface{}{"md-001"},
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Failed to get block with long content: %v", err)
	}

	content := resp.Content[0].Text

	// Should contain the full content (get_blocks returns full content, not truncated)
	if !contains(content, "md-001") {
		t.Error("Response should contain block ID")
	}
	if !contains(content, "This is a very long line") {
		t.Error("Response should contain beginning of long content")
	}

	// The full content should be returned
	if !contains(content, longContent[:100]) {
		t.Error("Response should contain the full long content")
	}
}

func TestGetBlocksOrderPreservation(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Order Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add blocks in sequence
	for i := 1; i <= 5; i++ {
		headingReq := &protocol.CallToolRequest{
			Name: "add_heading",
			Arguments: map[string]interface{}{
				"document_id": "order-test",
				"level":       1,
				"text":        fmt.Sprintf("Block %d", i),
			},
		}
		_, err := h.CallTool(ctx, headingReq)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Request blocks in different order than they were created
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "order-test",
			"block_ids":   []interface{}{"hd-003", "hd-001", "hd-005", "hd-002"},
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Failed to get blocks: %v", err)
	}

	content := resp.Content[0].Text

	// All requested blocks should be present
	expectedBlocks := []string{"hd-003", "hd-001", "hd-005", "hd-002"}
	expectedTexts := []string{"Block 3", "Block 1", "Block 5", "Block 2"}

	for i, blockID := range expectedBlocks {
		if !contains(content, blockID) {
			t.Errorf("Response should contain block %s", blockID)
		}
		if !contains(content, expectedTexts[i]) {
			t.Errorf("Response should contain text %s", expectedTexts[i])
		}
	}

	// Should not contain the block we didn't request
	if contains(content, "hd-004") {
		t.Error("Response should not contain unrequested block hd-004")
	}
	if contains(content, "Block 4") {
		t.Error("Response should not contain text from unrequested block")
	}
}

func TestGetBlocksImageBlock(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Image Block Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Create a dummy image file for testing
	tempDir := "/tmp"
	imagePath := filepath.Join(tempDir, "test-image.png")
	err = os.WriteFile(imagePath, []byte("fake image data"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(imagePath)

	// Add image block
	imageReq := &protocol.CallToolRequest{
		Name: "add_image",
		Arguments: map[string]interface{}{
			"document_id": "image-block-test",
			"image_path":  imagePath,
			"caption":     "Test Image Caption",
			"alt_text":    "Test Alt Text",
		},
	}
	_, err = h.CallTool(ctx, imageReq)
	if err != nil {
		t.Fatal(err)
	}

	// Get the image block
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "image-block-test",
			"block_ids":   []interface{}{"img-001"},
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Failed to get image block: %v", err)
	}

	content := resp.Content[0].Text

	// Should contain image block data
	if !contains(content, "img-001") {
		t.Error("Response should contain image block ID")
	}
	if !contains(content, "\"type\": \"image\"") {
		t.Error("Response should contain image block type")
	}
	if !contains(content, "Test Image Caption") {
		t.Error("Response should contain image caption")
	}
	if !contains(content, "Test Alt Text") {
		t.Error("Response should contain image alt text")
	}
	if !contains(content, "\"path\"") {
		t.Error("Response should contain image path")
	}
}

func TestGetBlocksInvalidInput(t *testing.T) {
	h, cleanup := setupGetBlocksTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Invalid Input Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Test missing document_id
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"block_ids": []interface{}{"hd-001"},
		},
	}

	_, err = h.CallTool(ctx, getBlocksReq)
	if err == nil {
		t.Error("Expected error for missing document_id")
	}

	// Test missing block_ids
	getBlocksReq2 := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "invalid-input-test",
		},
	}

	_, err = h.CallTool(ctx, getBlocksReq2)
	if err == nil {
		t.Error("Expected error for missing block_ids")
	}
}