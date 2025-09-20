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

func setupOverviewTestHandler(t *testing.T) (*handler.Handler, func()) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "docgen-overview-test-*")
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

func TestDocumentOverviewEmpty(t *testing.T) {
	h, cleanup := setupOverviewTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create empty document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title":  "Empty Document",
			"author": "Test Author",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Get overview
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "empty-document",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Failed to get overview: %v", err)
	}

	if len(resp.Content) == 0 {
		t.Fatal("Expected response content")
	}

	content := resp.Content[0].Text
	if !contains(content, "Empty Document") {
		t.Error("Overview should contain document title")
	}
	if !contains(content, "Test Author") {
		t.Error("Overview should contain author")
	}
	if !contains(content, "\"has_chapters\": false") {
		t.Error("Overview should indicate flat document")
	}
}

func TestDocumentOverviewSmallDocument(t *testing.T) {
	h, cleanup := setupOverviewTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Small Document",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add a few blocks
	blocks := []map[string]interface{}{
		{
			"type": "add_heading",
			"args": map[string]interface{}{
				"document_id": "small-document",
				"level":       1,
				"text":        "Introduction",
			},
		},
		{
			"type": "add_markdown",
			"args": map[string]interface{}{
				"document_id": "small-document",
				"content":     "This is a short introduction to the document.",
			},
		},
		{
			"type": "add_table",
			"args": map[string]interface{}{
				"document_id": "small-document",
				"headers":     []interface{}{"Name", "Value"},
				"rows": []interface{}{
					[]interface{}{"Item 1", "10"},
					[]interface{}{"Item 2", "20"},
				},
			},
		},
	}

	for _, block := range blocks {
		req := &protocol.CallToolRequest{
			Name:      block["type"].(string),
			Arguments: block["args"].(map[string]interface{}),
		}
		_, err := h.CallTool(ctx, req)
		if err != nil {
			t.Fatalf("Failed to add block: %v", err)
		}
	}

	// Get overview
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "small-document",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Failed to get overview: %v", err)
	}

	content := resp.Content[0].Text

	// Should contain all block previews
	if !contains(content, "H1: Introduction") {
		t.Error("Overview should contain heading preview")
	}
	if !contains(content, "This is a short introduction") {
		t.Error("Overview should contain markdown preview")
	}
	if !contains(content, "Table: 2 columns, 2 rows") {
		t.Error("Overview should contain table preview")
	}

	// Should show block IDs
	if !contains(content, "hd-001") {
		t.Error("Overview should contain heading block ID")
	}
	if !contains(content, "md-001") {
		t.Error("Overview should contain markdown block ID")
	}
	if !contains(content, "tbl-001") {
		t.Error("Overview should contain table block ID")
	}
}

func TestDocumentOverviewLargeDocument(t *testing.T) {
	h, cleanup := setupOverviewTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Large Document",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add many blocks (50 blocks to test large document handling)
	for i := 1; i <= 50; i++ {
		// Add heading
		headingReq := &protocol.CallToolRequest{
			Name: "add_heading",
			Arguments: map[string]interface{}{
				"document_id": "large-document",
				"level":       i%6 + 1, // Level 1-6
				"text":        fmt.Sprintf("Section %d", i),
			},
		}
		_, err := h.CallTool(ctx, headingReq)
		if err != nil {
			t.Fatalf("Failed to add heading %d: %v", i, err)
		}

		// Add markdown with long content
		longContent := strings.Repeat(fmt.Sprintf("This is paragraph %d with lots of content. ", i), 10)
		markdownReq := &protocol.CallToolRequest{
			Name: "add_markdown",
			Arguments: map[string]interface{}{
				"document_id": "large-document",
				"content":     longContent,
			},
		}
		_, err = h.CallTool(ctx, markdownReq)
		if err != nil {
			t.Fatalf("Failed to add markdown %d: %v", i, err)
		}
	}

	// Get overview
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "large-document",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Failed to get overview: %v", err)
	}

	content := resp.Content[0].Text

	// Should contain first and last blocks
	// For i=1: level = 1%6 + 1 = 2, so H2: Section 1
	// For i=50: level = 50%6 + 1 = 3, so H3: Section 50
	if !contains(content, "H2: Section 1") {
		t.Error("Overview should contain first heading")
	}
	if !contains(content, "H3: Section 50") {
		t.Error("Overview should contain last heading")
	}

	// Should show truncated previews (not full content)
	if !contains(content, "...") {
		t.Error("Long content should be truncated in overview")
	}

	// Should contain all block IDs
	if !contains(content, "hd-001") {
		t.Error("Overview should contain first heading ID")
	}
	if !contains(content, "hd-050") {
		t.Error("Overview should contain last heading ID")
	}
	if !contains(content, "md-050") {
		t.Error("Overview should contain last markdown ID")
	}

	// Overview should still be reasonably sized (not gigantic)
	if len(content) > 50000 {
		t.Errorf("Overview seems too large (%d chars), should be truncated", len(content))
	}
}

func TestDocumentOverviewChapteredDocument(t *testing.T) {
	h, cleanup := setupOverviewTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create chaptered document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title":        "Chaptered Book",
			"has_chapters": true,
			"author":       "Book Author",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add chapters
	chapters := []string{"Introduction", "Main Content", "Conclusion"}
	for i, chapterTitle := range chapters {
		// Add chapter
		chapterReq := &protocol.CallToolRequest{
			Name: "add_chapter",
			Arguments: map[string]interface{}{
				"document_id": "chaptered-book",
				"title":       chapterTitle,
			},
		}
		_, err := h.CallTool(ctx, chapterReq)
		if err != nil {
			t.Fatalf("Failed to add chapter: %v", err)
		}

		chapterID := fmt.Sprintf("ch-%03d", i+1)

		// Add blocks to each chapter
		headingReq := &protocol.CallToolRequest{
			Name: "add_heading",
			Arguments: map[string]interface{}{
				"document_id": "chaptered-book",
				"chapter_id":  chapterID,
				"level":       1,
				"text":        chapterTitle,
			},
		}
		_, err = h.CallTool(ctx, headingReq)
		if err != nil {
			t.Fatalf("Failed to add heading to chapter: %v", err)
		}

		markdownReq := &protocol.CallToolRequest{
			Name: "add_markdown",
			Arguments: map[string]interface{}{
				"document_id": "chaptered-book",
				"chapter_id":  chapterID,
				"content":     fmt.Sprintf("Content for %s chapter.", chapterTitle),
			},
		}
		_, err = h.CallTool(ctx, markdownReq)
		if err != nil {
			t.Fatalf("Failed to add markdown to chapter: %v", err)
		}
	}

	// Get overview
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "chaptered-book",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Failed to get overview: %v", err)
	}

	content := resp.Content[0].Text

	// Should show document info
	if !contains(content, "Chaptered Book") {
		t.Error("Overview should contain document title")
	}
	if !contains(content, "\"has_chapters\": true") {
		t.Error("Overview should indicate chaptered document")
	}

	// Should show all chapters
	for i, chapterTitle := range chapters {
		chapterID := fmt.Sprintf("ch-%03d", i+1)
		if !contains(content, chapterID) {
			t.Errorf("Overview should contain chapter ID %s", chapterID)
		}
		if !contains(content, chapterTitle) {
			t.Errorf("Overview should contain chapter title %s", chapterTitle)
		}
	}

	// Should show blocks within chapters
	if !contains(content, "H1: Introduction") {
		t.Error("Overview should contain heading in Introduction chapter")
	}
	if !contains(content, "Content for Introduction") {
		t.Error("Overview should contain markdown in Introduction chapter")
	}
}

func TestDocumentOverviewLargeChapteredDocument(t *testing.T) {
	h, cleanup := setupOverviewTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create chaptered document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title":        "Large Book",
			"has_chapters": true,
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add many chapters with multiple blocks each
	for chapterNum := 1; chapterNum <= 20; chapterNum++ {
		// Add chapter
		chapterReq := &protocol.CallToolRequest{
			Name: "add_chapter",
			Arguments: map[string]interface{}{
				"document_id": "large-book",
				"title":       fmt.Sprintf("Chapter %d", chapterNum),
			},
		}
		_, err := h.CallTool(ctx, chapterReq)
		if err != nil {
			t.Fatalf("Failed to add chapter %d: %v", chapterNum, err)
		}

		chapterID := fmt.Sprintf("ch-%03d", chapterNum)

		// Add multiple blocks to each chapter
		for blockNum := 1; blockNum <= 5; blockNum++ {
			headingReq := &protocol.CallToolRequest{
				Name: "add_heading",
				Arguments: map[string]interface{}{
					"document_id": "large-book",
					"chapter_id":  chapterID,
					"level":       blockNum%3 + 1,
					"text":        fmt.Sprintf("Section %d.%d", chapterNum, blockNum),
				},
			}
			_, err := h.CallTool(ctx, headingReq)
			if err != nil {
				t.Fatalf("Failed to add heading: %v", err)
			}

			longContent := strings.Repeat(fmt.Sprintf("Content for chapter %d section %d. ", chapterNum, blockNum), 20)
			markdownReq := &protocol.CallToolRequest{
				Name: "add_markdown",
				Arguments: map[string]interface{}{
					"document_id": "large-book",
					"chapter_id":  chapterID,
					"content":     longContent,
				},
			}
			_, err = h.CallTool(ctx, markdownReq)
			if err != nil {
				t.Fatalf("Failed to add markdown: %v", err)
			}
		}
	}

	// Get overview
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "large-book",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Failed to get overview: %v", err)
	}

	content := resp.Content[0].Text

	// Should contain all chapters
	if !contains(content, "ch-001") {
		t.Error("Overview should contain first chapter")
	}
	if !contains(content, "ch-020") {
		t.Error("Overview should contain last chapter")
	}

	// Should contain block previews for all chapters
	// For blockNum=1: level = 1%3 + 1 = 2, so H2: Section 1.1
	// For blockNum=1: level = 1%3 + 1 = 2, so H2: Section 20.1
	if !contains(content, "H2: Section 1.1") {
		t.Error("Overview should contain first heading")
	}
	if !contains(content, "H2: Section 20.1") {
		t.Error("Overview should contain last chapter's heading")
	}

	// Content should be truncated appropriately
	if !contains(content, "...") {
		t.Error("Long content should be truncated")
	}

	// Overview should be manageable size even for large documents
	if len(content) > 100000 {
		t.Errorf("Overview too large (%d chars) for large document", len(content))
	}
}

func TestDocumentOverviewWithAllBlockTypes(t *testing.T) {
	h, cleanup := setupOverviewTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "All Block Types Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add all types of blocks
	blocks := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "add_heading",
			args: map[string]interface{}{
				"document_id": "all-block-types-test",
				"level":       2,
				"text":        "Test Heading Level 2",
			},
		},
		{
			name: "add_markdown",
			args: map[string]interface{}{
				"document_id": "all-block-types-test",
				"content":     "This is **markdown** content with *formatting* and `code`.",
			},
		},
		{
			name: "add_table",
			args: map[string]interface{}{
				"document_id": "all-block-types-test",
				"headers":     []interface{}{"Column A", "Column B", "Column C"},
				"rows": []interface{}{
					[]interface{}{"Row 1A", "Row 1B", "Row 1C"},
					[]interface{}{"Row 2A", "Row 2B", "Row 2C"},
					[]interface{}{"Row 3A", "Row 3B", "Row 3C"},
				},
			},
		},
		{
			name: "add_page_break",
			args: map[string]interface{}{
				"document_id": "all-block-types-test",
			},
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

	// Get overview
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "all-block-types-test",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Failed to get overview: %v", err)
	}

	content := resp.Content[0].Text

	// Check that all block types appear with correct previews
	if !contains(content, "H2: Test Heading Level 2") {
		t.Error("Overview should contain heading preview with level")
	}
	if !contains(content, "This is **markdown** content") {
		t.Error("Overview should contain markdown preview")
	}
	if !contains(content, "Table: 3 columns, 3 rows") {
		t.Error("Overview should contain table preview with dimensions")
	}
	if !contains(content, "Page Break") {
		t.Error("Overview should contain page break preview")
	}

	// Check block IDs
	expectedIDs := []string{"hd-001", "md-001", "tbl-001", "pb-001"}
	for _, id := range expectedIDs {
		if !contains(content, id) {
			t.Errorf("Overview should contain block ID %s", id)
		}
	}
}

func TestDocumentOverviewNonexistentDocument(t *testing.T) {
	h, cleanup := setupOverviewTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Try to get overview of non-existent document
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "nonexistent-document",
		},
	}

	_, err := h.CallTool(ctx, overviewReq)
	if err == nil {
		t.Error("Expected error for non-existent document")
	}
}

func TestDocumentOverviewTruncation(t *testing.T) {
	h, cleanup := setupOverviewTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title": "Truncation Test",
		},
	}

	_, err := h.CallTool(ctx, createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add heading with very long text
	longHeadingText := strings.Repeat("Very Long Heading Text ", 20)
	headingReq := &protocol.CallToolRequest{
		Name: "add_heading",
		Arguments: map[string]interface{}{
			"document_id": "truncation-test",
			"level":       1,
			"text":        longHeadingText,
		},
	}
	_, err = h.CallTool(ctx, headingReq)
	if err != nil {
		t.Fatal(err)
	}

	// Add markdown with very long content
	longMarkdownContent := strings.Repeat("This is a very long paragraph with lots and lots of text that should be truncated in the overview. ", 50)
	markdownReq := &protocol.CallToolRequest{
		Name: "add_markdown",
		Arguments: map[string]interface{}{
			"document_id": "truncation-test",
			"content":     longMarkdownContent,
		},
	}
	_, err = h.CallTool(ctx, markdownReq)
	if err != nil {
		t.Fatal(err)
	}

	// Get overview
	overviewReq := &protocol.CallToolRequest{
		Name: "get_document_overview",
		Arguments: map[string]interface{}{
			"document_id": "truncation-test",
		},
	}

	resp, err := h.CallTool(ctx, overviewReq)
	if err != nil {
		t.Fatalf("Failed to get overview: %v", err)
	}

	content := resp.Content[0].Text

	// Should contain truncation indicators
	if !contains(content, "...") {
		t.Error("Long content should be truncated with ellipsis")
	}

	// The preview should be much shorter than the original content
	if contains(content, longHeadingText) {
		t.Error("Full long heading should not appear in preview")
	}
	if contains(content, longMarkdownContent) {
		t.Error("Full long markdown should not appear in preview")
	}

	// But should still contain the beginning of the content
	if !contains(content, "Very Long Heading Text") {
		t.Error("Preview should contain beginning of heading text")
	}
	if !contains(content, "This is a very long paragraph") {
		t.Error("Preview should contain beginning of markdown text")
	}
}