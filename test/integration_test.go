package test

import (
	"context"
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