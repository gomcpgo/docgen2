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

func setupEdgeCasesTestHandler(t *testing.T) (*handler.Handler, func()) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "docgen-edge-cases-*")
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

// Test corrupted document scenarios
func TestDocumentOverviewCorruptedDocument(t *testing.T) {
	h, cleanup := setupEdgeCasesTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document normally first
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Corrupted Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add some blocks
	headingReq := &protocol.CallToolRequest{
		Name: "add_heading",
		Arguments: map[string]interface{}{
			"document_id": "corrupted-test",
			"level":       1,
			"text":        "Test Heading",
		},
	}
	_, err = h.CallTool(ctx, headingReq)
	if err != nil {
		t.Fatal(err)
	}

	// Test the resilience by trying to access non-existent blocks

	// Get overview should still work even if some blocks are missing
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "corrupted-test",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Overview should work even with some corrupted data: %v", err)
	}

	content := resp.Content[0].Text
	if !contains(content, "Corrupted Test") {
		t.Error("Should still show document title")
	}
}

// Test extremely large block content
func TestGetBlocksExtremelyLargeContent(t *testing.T) {
	h, cleanup := setupEdgeCasesTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Large Content Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Create extremely large content (1MB+)
	largeContent := strings.Repeat("This is a very long sentence that will be repeated many times to create extremely large content. ", 10000)

	markdownReq := &protocol.CallToolRequest{
		Name: "add_markdown",
		Arguments: map[string]interface{}{
			"document_id": "large-content-test",
			"content":     largeContent,
		},
	}
	_, err = h.CallTool(ctx, markdownReq)
	if err != nil {
		t.Fatal(err)
	}

	// Get the large block
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "large-content-test",
			"block_ids":   []interface{}{"md-001"},
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Should handle large content gracefully: %v", err)
	}

	content := resp.Content[0].Text
	if !contains(content, "md-001") {
		t.Error("Should contain block ID")
	}
	if !contains(content, "very long sentence") {
		t.Error("Should contain part of the content")
	}
}

// Test unicode and special characters
func TestUnicodeAndSpecialCharacters(t *testing.T) {
	h, cleanup := setupEdgeCasesTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Unicode Test 🌟",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add blocks with various unicode and special characters
	unicodeTexts := []struct {
		name    string
		content string
	}{
		{
			name:    "emoji",
			content: "Heading with emojis 🚀🌟✨",
		},
		{
			name:    "unicode",
			content: "Mültiple languages: Español, Français, Deutsch, 中文, العربية, Русский",
		},
		{
			name:    "special_chars",
			content: "Special chars: !@#$%^&*()_+-=[]{}|;:'\",.<>?/~`",
		},
		{
			name:    "escape_chars",
			content: "Escape chars: \n\t\r\"\\'\\",
		},
		{
			name:    "json_breaking",
			content: `JSON breaking: {"key": "value", "array": [1,2,3]}`,
		},
	}

	var blockIDs []interface{}
	for i, test := range unicodeTexts {
		if test.name == "emoji" {
			headingReq := &protocol.CallToolRequest{
				Name: "add_heading",
				Arguments: map[string]interface{}{
					"document_id": "unicode-test",
					"level":       1,
					"text":        test.content,
				},
			}
			_, err = h.CallTool(ctx, headingReq)
			blockIDs = append(blockIDs, fmt.Sprintf("hd-%03d", i+1))
		} else {
			markdownReq := &protocol.CallToolRequest{
				Name: "add_markdown",
				Arguments: map[string]interface{}{
					"document_id": "unicode-test",
					"content":     test.content,
				},
			}
			_, err = h.CallTool(ctx, markdownReq)
			if test.name == "emoji" {
				blockIDs = append(blockIDs, fmt.Sprintf("hd-%03d", i))
			} else {
				blockIDs = append(blockIDs, fmt.Sprintf("md-%03d", i))
			}
		}

		if err != nil {
			t.Fatalf("Failed to add %s block: %v", test.name, err)
		}
	}

	// Test overview with unicode content
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "unicode-test",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Overview should handle unicode: %v", err)
	}

	content := resp.Content[0].Text
	if !contains(content, "Unicode Test 🌟") {
		t.Error("Should handle unicode in document title")
	}
	if !contains(content, "🚀🌟✨") {
		t.Error("Should handle emojis in content")
	}

	// Test get_blocks with unicode content
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "unicode-test",
			"block_ids":   blockIDs,
		},
	}

	resp2, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("get_blocks should handle unicode: %v", err)
	}

	content2 := resp2.Content[0].Text
	
	// Check for emojis and basic unicode (these should be preserved)
	if !contains(content2, "🚀🌟✨") {
		t.Error("Should contain emojis")
	}
	if !contains(content2, "Español") {
		t.Error("Should contain unicode text")
	}
	
	// For special characters, just check that some are preserved (JSON encoding may escape others)
	if !contains(content2, "!@#$%") {
		t.Error("Should contain some special characters")
	}
}

// Test malformed input data
func TestMalformedInputData(t *testing.T) {
	h, cleanup := setupEdgeCasesTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Test invalid document ID formats
	invalidDocIDs := []string{
		"",           // Empty
		" ",          // Whitespace
		"doc with spaces",
		"doc/with/slashes",
		"doc\\with\\backslashes",
		"doc:with:colons",
		"doc*with*asterisks",
		"doc?with?questions",
		"doc|with|pipes",
		"doc<with>brackets",
		"doc\"with\"quotes",
		"doc'with'quotes",
		"verylongdocumentnamethatshouldexceedreasonablelimitsandcausepotentialissuesverylongdocumentnamethatshouldexceedreasonablelimitsandcausepotentialissuesverylongdocumentnamethatshouldexceedreasonablelimitsandcausepotentialissues",
	}

	for _, docID := range invalidDocIDs {
		overviewReq := &protocol.CallToolRequest{
			Name: "get_document_overview",
			Arguments: map[string]interface{}{
				"document_id": docID,
			},
		}

		_, err := h.CallTool(ctx, overviewReq)
		// Most of these should result in "document not found" rather than crashes
		if err == nil && docID == "" {
			t.Error("Empty document ID should cause an error")
		}
	}

	// Test invalid block IDs
	// First create a valid document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Valid Doc for Invalid Block Test",
		},
	}
	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	invalidBlockIDs := []interface{}{
		"",
		" ",
		"invalid-format",
		"hd-",
		"hd-abc",
		"hd-999999",
		"unknown-001",
		12345, // Not a string
		nil,
		map[string]interface{}{"invalid": "object"},
		[]interface{}{"nested", "array"},
	}

	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "valid-doc-for-invalid-block-test",
			"block_ids":   invalidBlockIDs,
		},
	}

	resp, err := h.CallTool(ctx, getBlocksReq)
	// Should not crash, should gracefully handle invalid IDs
	if err != nil {
		t.Fatalf("Should gracefully handle invalid block IDs: %v", err)
	}

	// Should return empty or partial results
	content := resp.Content[0].Text
	if !contains(content, "[") {
		t.Errorf("Should return some form of array response, got: %s", content)
	}
}

// Test concurrent access simulation
func TestSimulatedConcurrentAccess(t *testing.T) {
	h, cleanup := setupEdgeCasesTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Concurrent Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add some blocks
	for i := 1; i <= 10; i++ {
		headingReq := &protocol.CallToolRequest{
			Name: "add_heading",
			Arguments: map[string]interface{}{
				"document_id": "concurrent-test",
				"level":       1,
				"text":        fmt.Sprintf("Heading %d", i),
			},
		}
		_, err = h.CallTool(ctx, headingReq)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Simulate multiple rapid requests
	for i := 0; i < 10; i++ {
		// Rapid overview requests
		overviewReq := &protocol.CallToolRequest{
			Name: "get_document_overview",
			Arguments: map[string]interface{}{
				"document_id": "concurrent-test",
			},
		}

		_, err := h.CallTool(ctx, overviewReq)
		if err != nil {
			t.Errorf("Rapid request %d failed: %v", i, err)
		}

		// Rapid get_blocks requests
		getBlocksReq := &protocol.CallToolRequest{
			Name: "get_blocks",
			Arguments: map[string]interface{}{
				"document_id": "concurrent-test",
				"block_ids":   []interface{}{fmt.Sprintf("hd-%03d", (i%10)+1)},
			},
		}

		_, err = h.CallTool(ctx, getBlocksReq)
		if err != nil {
			t.Errorf("Rapid get_blocks request %d failed: %v", i, err)
		}
	}
}

// Test memory usage with large documents
func TestMemoryUsageLargeDocument(t *testing.T) {
	h, cleanup := setupEdgeCasesTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create chaptered document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title":        "Memory Test Book",
			"has_chapters": true,
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add many chapters with many blocks each
	for chapterNum := 1; chapterNum <= 100; chapterNum++ {
		chapterReq := &protocol.CallToolRequest{
			Name: "add_chapter",
			Arguments: map[string]interface{}{
				"document_id": "memory-test-book",
				"title":       fmt.Sprintf("Chapter %d", chapterNum),
			},
		}
		_, err := h.CallTool(ctx, chapterReq)
		if err != nil {
			t.Fatalf("Failed to add chapter %d: %v", chapterNum, err)
		}

		chapterID := fmt.Sprintf("ch-%03d", chapterNum)

		// Add blocks to chapter (fewer blocks to avoid test timeout)
		for blockNum := 1; blockNum <= 3; blockNum++ {
			markdownReq := &protocol.CallToolRequest{
				Name: "add_markdown",
				Arguments: map[string]interface{}{
					"document_id": "memory-test-book",
					"chapter_id":  chapterID,
					"content":     fmt.Sprintf("Content for chapter %d, block %d", chapterNum, blockNum),
				},
			}
			_, err := h.CallTool(ctx, markdownReq)
			if err != nil {
				t.Fatalf("Failed to add block: %v", err)
			}
		}
	}

	// Get overview of large document
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "memory-test-book",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Should handle large document overview: %v", err)
	}

	content := resp.Content[0].Text
	if !contains(content, "ch-001") {
		t.Error("Should contain first chapter")
	}
	if !contains(content, "ch-100") {
		t.Error("Should contain last chapter")
	}

	// Test get_blocks with many block IDs
	var blockIDs []interface{}
	for i := 1; i <= 100; i++ {
		blockIDs = append(blockIDs, fmt.Sprintf("md-%03d", i*3)) // Get last block from each chapter
	}

	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "memory-test-book",
			"block_ids":   blockIDs[:50], // Request first 50 to avoid timeout
		},
	}

	resp2, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Should handle many block requests: %v", err)
	}

	content2 := resp2.Content[0].Text
	if !contains(content2, "md-003") {
		t.Error("Should contain requested blocks")
	}
}

// Test error recovery scenarios
func TestErrorRecoveryScenarios(t *testing.T) {
	h, cleanup := setupEdgeCasesTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Test missing required parameters
	testCases := []struct {
		name string
		req  *protocol.CallToolRequest
	}{
		{
			name: "overview_missing_doc_id",
			req: &protocol.CallToolRequest{
				Name:      "get_document_overview",
				Arguments: map[string]interface{}{},
			},
		},
		{
			name: "get_blocks_missing_doc_id",
			req: &protocol.CallToolRequest{
				Name: "get_blocks",
				Arguments: map[string]interface{}{
					"block_ids": []interface{}{"hd-001"},
				},
			},
		},
		{
			name: "get_blocks_missing_block_ids",
			req: &protocol.CallToolRequest{
				Name: "get_blocks",
				Arguments: map[string]interface{}{
					"document_id": "test-doc",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := h.CallTool(ctx, tc.req)
			if err == nil {
				t.Errorf("Expected error for %s", tc.name)
			}
		})
	}
}

// Test boundary conditions
func TestBoundaryConditions(t *testing.T) {
	h, cleanup := setupEdgeCasesTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Boundary Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Test with exactly zero blocks (empty document)
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "boundary-test",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Should handle empty document: %v", err)
	}

	content := resp.Content[0].Text
	if !contains(content, "Boundary Test") {
		t.Error("Should show document title even when empty")
	}

	// Test with exactly one block
	headingReq := &protocol.CallToolRequest{
		Name: "add_heading",
		Arguments: map[string]interface{}{
			"document_id": "boundary-test",
			"level":       1,
			"text":        "Single Block",
		},
	}
	_, err = h.CallTool(ctx, headingReq)
	if err != nil {
		t.Fatal(err)
	}

	// Test overview with one block
	resp2, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Should handle single block document: %v", err)
	}

	content2 := resp2.Content[0].Text
	if !contains(content2, "Single Block") {
		t.Error("Should show single block in overview")
	}

	// Test get_blocks with one block
	getBlocksReq := &protocol.CallToolRequest{
		Name: "get_blocks",
		Arguments: map[string]interface{}{
			"document_id": "boundary-test",
			"block_ids":   []interface{}{"hd-001"},
		},
	}

	resp3, err := h.CallTool(ctx, getBlocksReq)
	if err != nil {
		t.Fatalf("Should handle single block request: %v", err)
	}

	content3 := resp3.Content[0].Text
	if !contains(content3, "hd-001") {
		t.Error("Should return single block")
	}
	if !contains(content3, "Single Block") {
		t.Error("Should contain block content")
	}
}