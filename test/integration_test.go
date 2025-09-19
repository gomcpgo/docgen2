package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/savant/mcp-servers/docgen2/pkg/config"
	"github.com/savant/mcp-servers/docgen2/pkg/handler"
)

func setupTestHandler(t *testing.T) (*handler.Handler, func()) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "docgen-integration-*")
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

func TestCreateDocumentFlow(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create a document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title":        "Test Document",
			"has_chapters": false,
			"author":       "Test Author",
		},
	}
	
	resp, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	
	if len(resp.Content) == 0 {
		t.Fatal("Expected response content")
	}
	
	// List documents
	listReq := &protocol.CallToolRequest{
		Name:      "list_documents",
		Arguments: map[string]interface{}{},
	}
	
	listResp, err := h.CallTool(ctx, listReq)
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}
	
	if len(listResp.Content) == 0 {
		t.Fatal("Expected list response")
	}
}

func TestAddBlocksFlow(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create document first
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Block Test",
		},
	}
	
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add heading
	headingReq := &protocol.CallToolRequest{
		Name: "add_heading",
		Arguments: map[string]interface{}{
			"document_id": "block-test",
			"level":       1,
			"text":        "Chapter One",
		},
	}
	
	_, err = h.CallTool(ctx, headingReq)
	if err != nil {
		t.Fatalf("Failed to add heading: %v", err)
	}
	
	// Add markdown
	markdownReq := &protocol.CallToolRequest{
		Name: "add_markdown",
		Arguments: map[string]interface{}{
			"document_id": "block-test",
			"content":     "This is some markdown content.",
		},
	}
	
	_, err = h.CallTool(ctx, markdownReq)
	if err != nil {
		t.Fatalf("Failed to add markdown: %v", err)
	}
	
	// Add table
	tableReq := &protocol.CallToolRequest{
		Name: "add_table",
		Arguments: map[string]interface{}{
			"document_id": "block-test",
			"headers":     []interface{}{"Name", "Age"},
			"rows": []interface{}{
				[]interface{}{"Alice", "30"},
				[]interface{}{"Bob", "25"},
			},
		},
	}
	
	_, err = h.CallTool(ctx, tableReq)
	if err != nil {
		t.Fatalf("Failed to add table: %v", err)
	}
	
	// Get overview
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "block-test",
		},
	}
	
	overviewResp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Failed to get overview: %v", err)
	}
	
	if len(overviewResp.Content) == 0 {
		t.Fatal("Expected overview content")
	}
}

func TestChapterFlow(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create chaptered document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title":        "Book Test",
			"has_chapters": true,
		},
	}
	
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add chapter
	chapterReq := &protocol.CallToolRequest{
		Name: "add_chapter",
		Arguments: map[string]interface{}{
			"document_id": "book-test",
			"title":       "Introduction",
		},
	}
	
	chapterResp, err := h.CallTool(ctx, chapterReq)
	if err != nil {
		t.Fatalf("Failed to add chapter: %v", err)
	}
	
	if len(chapterResp.Content) == 0 {
		t.Fatal("Expected chapter response")
	}
	
	// Add heading to chapter
	headingReq := &protocol.CallToolRequest{
		Name: "add_heading",
		Arguments: map[string]interface{}{
			"document_id": "book-test",
			"chapter_id":  "ch-001",
			"level":       2,
			"text":        "Welcome",
		},
	}
	
	_, err = h.CallTool(ctx, headingReq)
	if err != nil {
		t.Fatalf("Failed to add heading to chapter: %v", err)
	}
}

func TestSearchFlow(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create document with content
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Search Test",
		},
	}
	
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add content to search
	markdownReq := &protocol.CallToolRequest{
		Name: "add_markdown",
		Arguments: map[string]interface{}{
			"document_id": "search-test",
			"content":     "This document contains information about butterflies and their lifecycle.",
		},
	}
	
	_, err = h.CallTool(ctx, markdownReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add more content
	markdown2Req := &protocol.CallToolRequest{
		Name: "add_markdown",
		Arguments: map[string]interface{}{
			"document_id": "search-test",
			"content":     "Butterflies are beautiful insects with colorful wings.",
		},
	}
	
	_, err = h.CallTool(ctx, markdown2Req)
	if err != nil {
		t.Fatal(err)
	}
	
	// Search for content
	searchReq := &protocol.CallToolRequest{
		Name: "search_blocks",
		Arguments: map[string]interface{}{
			"document_id": "search-test",
			"query":       "butterflies",
		},
	}
	
	searchResp, err := h.CallTool(ctx, searchReq)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	
	if len(searchResp.Content) == 0 {
		t.Fatal("Expected search results")
	}
}

func TestMultipleBlocksFlow(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Multi Block Test",
		},
	}
	
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add multiple blocks
	multiReq := &protocol.CallToolRequest{
		Name: "add_multiple_blocks",
		Arguments: map[string]interface{}{
			"document_id": "multi-block-test",
			"blocks": []interface{}{
				map[string]interface{}{
					"type": "heading",
					"data": map[string]interface{}{
						"level": 1,
						"text":  "Title",
					},
				},
				map[string]interface{}{
					"type": "markdown",
					"data": map[string]interface{}{
						"content": "Some content here.",
					},
				},
				map[string]interface{}{
					"type": "heading",
					"data": map[string]interface{}{
						"level": 2,
						"text":  "Subtitle",
					},
				},
			},
		},
	}
	
	multiResp, err := h.CallTool(ctx, multiReq)
	if err != nil {
		t.Fatalf("Failed to add multiple blocks: %v", err)
	}
	
	if len(multiResp.Content) == 0 {
		t.Fatal("Expected response")
	}
}

func TestGetBlockFlow(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Get Block Test",
		},
	}
	
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add a heading
	headingReq := &protocol.CallToolRequest{
		Name: "add_heading",
		Arguments: map[string]interface{}{
			"document_id": "get-block-test",
			"level":       1,
			"text":        "Test Heading",
		},
	}
	
	_, err = h.CallTool(ctx, headingReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Get the block
	getReq := &protocol.CallToolRequest{
		Name: "get_block",
		Arguments: map[string]interface{}{
			"document_id": "get-block-test",
			"block_id":    "hd-001",
		},
	}
	
	getResp, err := h.CallTool(ctx, getReq)
	if err != nil {
		t.Fatalf("Failed to get block: %v", err)
	}
	
	if len(getResp.Content) == 0 {
		t.Fatal("Expected block content")
	}
}

func TestUpdateBlockFlow(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Update Block Test",
		},
	}
	
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add a heading
	addReq := &protocol.CallToolRequest{
		Name: "add_heading",
		Arguments: map[string]interface{}{
			"document_id": "update-block-test",
			"level":       1,
			"text":        "Original Heading",
		},
	}
	
	_, err = h.CallTool(ctx, addReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Update the heading
	updateReq := &protocol.CallToolRequest{
		Name: "update_block",
		Arguments: map[string]interface{}{
			"document_id": "update-block-test",
			"block_id":    "hd-001",
			"new_content": map[string]interface{}{
				"level": 2,
				"text":  "Updated Heading",
			},
		},
	}
	
	updateResp, err := h.CallTool(ctx, updateReq)
	if err != nil {
		t.Fatalf("Failed to update block: %v", err)
	}
	
	if len(updateResp.Content) == 0 {
		t.Fatal("Expected update response")
	}
	
	// Verify the update worked
	getReq := &protocol.CallToolRequest{
		Name: "get_block",
		Arguments: map[string]interface{}{
			"document_id": "update-block-test",
			"block_id":    "hd-001",
		},
	}
	
	getResp, err := h.CallTool(ctx, getReq)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(getResp.Content) == 0 {
		t.Fatal("Expected block content")
	}
	
	// The response should contain the updated content
	content := getResp.Content[0].Text
	if !contains(content, "Updated Heading") {
		t.Error("Block should contain updated text")
	}
	
	if !contains(content, "\"level\": 2") {
		t.Error("Block should have updated level")
	}
}

func TestDeleteBlockFlow(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Delete Block Test",
		},
	}
	
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add multiple blocks
	for i := 1; i <= 3; i++ {
		addReq := &protocol.CallToolRequest{
			Name: "add_heading",
			Arguments: map[string]interface{}{
				"document_id": "delete-block-test",
				"level":       1,
				"text":        fmt.Sprintf("Heading %d", i),
			},
		}
		
		_, err = h.CallTool(ctx, addReq)
		if err != nil {
			t.Fatal(err)
		}
	}
	
	// Get overview to verify all blocks are there
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "delete-block-test",
		},
	}
	
	overviewResp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatal(err)
	}
	
	content := overviewResp.Content[0].Text
	if !contains(content, "hd-001") || !contains(content, "hd-002") || !contains(content, "hd-003") {
		t.Fatal("All three blocks should exist")
	}
	
	// Delete the middle block
	deleteReq := &protocol.CallToolRequest{
		Name: "delete_block",
		Arguments: map[string]interface{}{
			"document_id": "delete-block-test",
			"block_id":    "hd-002",
		},
	}
	
	deleteResp, err := h.CallTool(ctx, deleteReq)
	if err != nil {
		t.Fatalf("Failed to delete block: %v", err)
	}
	
	if len(deleteResp.Content) == 0 {
		t.Fatal("Expected delete response")
	}
	
	// Verify the block is gone
	overviewResp2, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatal(err)
	}
	
	content2 := overviewResp2.Content[0].Text
	if contains(content2, "hd-002") {
		t.Error("hd-002 should be deleted")
	}
	
	if !contains(content2, "hd-001") || !contains(content2, "hd-003") {
		t.Error("Other blocks should still exist")
	}
}

func TestMoveBlockFlow(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Move Block Test",
		},
	}
	
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add three blocks: A, B, C
	for _, letter := range []string{"A", "B", "C"} {
		addReq := &protocol.CallToolRequest{
			Name: "add_heading",
			Arguments: map[string]interface{}{
				"document_id": "move-block-test",
				"level":       1,
				"text":        "Block " + letter,
			},
		}
		
		_, err = h.CallTool(ctx, addReq)
		if err != nil {
			t.Fatal(err)
		}
	}
	
	// Get initial overview
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "move-block-test",
		},
	}
	
	overviewResp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatal(err)
	}
	
	initialContent := overviewResp.Content[0].Text
	
	// Move block B to start
	moveReq := &protocol.CallToolRequest{
		Name: "move_block",
		Arguments: map[string]interface{}{
			"document_id":  "move-block-test",
			"block_id":     "hd-002", // Block B
			"new_position": "start",
		},
	}
	
	moveResp, err := h.CallTool(ctx, moveReq)
	if err != nil {
		t.Fatalf("Failed to move block: %v", err)
	}
	
	if len(moveResp.Content) == 0 {
		t.Fatal("Expected move response")
	}
	
	// Verify the order changed
	overviewResp2, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatal(err)
	}
	
	finalContent := overviewResp2.Content[0].Text
	
	// The content should be different after the move
	if initialContent == finalContent {
		t.Error("Block order should have changed")
	}
	
	// Should still contain all blocks
	if !contains(finalContent, "Block A") || !contains(finalContent, "Block B") || !contains(finalContent, "Block C") {
		t.Error("All blocks should still exist after move")
	}
}

func TestUpdateBlockWrongType(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create document with heading
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Type Error Test",
		},
	}
	
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add heading
	addReq := &protocol.CallToolRequest{
		Name: "add_heading",
		Arguments: map[string]interface{}{
			"document_id": "type-error-test",
			"level":       1,
			"text":        "Test Heading",
		},
	}
	
	_, err = h.CallTool(ctx, addReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Try to update with markdown content (wrong type)
	updateReq := &protocol.CallToolRequest{
		Name: "update_block",
		Arguments: map[string]interface{}{
			"document_id": "type-error-test",
			"block_id":    "hd-001",
			"new_content": map[string]interface{}{
				"content": "This is markdown content",
			},
		},
	}
	
	_, err = h.CallTool(ctx, updateReq)
	if err == nil {
		t.Error("Expected error when updating heading with markdown content")
	}
}

func TestDeleteNonexistentBlock(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Delete Error Test",
		},
	}
	
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Try to delete non-existent block
	deleteReq := &protocol.CallToolRequest{
		Name: "delete_block",
		Arguments: map[string]interface{}{
			"document_id": "delete-error-test",
			"block_id":    "nonexistent-block",
		},
	}
	
	_, err = h.CallTool(ctx, deleteReq)
	if err == nil {
		t.Error("Expected error when deleting non-existent block")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || len(s) > len(substr) && 
		   (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		   findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}