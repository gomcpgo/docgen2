package handler

import (
	"context"
	"fmt"
	
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
)

// handleAddChapter adds a new chapter to a document
func (h *Handler) handleAddChapter(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	title, err := getString(args, "title", true)
	if err != nil {
		return nil, err
	}
	
	positionStr, _ := getString(args, "position", false)
	position := document.ParsePosition(positionStr)
	
	chapterID, err := h.storage.AddChapter(docID, title, position)
	if err != nil {
		return nil, fmt.Errorf("failed to add chapter: %w", err)
	}
	
	result := map[string]interface{}{
		"chapter_id": chapterID,
		"title":      title,
	}
	
	return jsonResponse(result)
}

// handleUpdateChapter updates a chapter's title
func (h *Handler) handleUpdateChapter(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// This would require implementation in storage layer
	return nil, fmt.Errorf("update_chapter not yet implemented")
}

// handleDeleteChapter deletes a chapter
func (h *Handler) handleDeleteChapter(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// This would require implementation in storage layer to:
	// 1. Remove chapter from manifest
	// 2. Delete chapter folder and all its contents
	return nil, fmt.Errorf("delete_chapter not yet implemented")
}

// handleMoveChapter moves a chapter to a new position
func (h *Handler) handleMoveChapter(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// This would require implementation to reorder chapters in manifest
	return nil, fmt.Errorf("move_chapter not yet implemented")
}

// handleExportDocument exports a document to various formats
func (h *Handler) handleExportDocument(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// This is a placeholder - export functionality will be implemented later
	// It will require:
	// 1. Consolidating all blocks into markdown
	// 2. Running Pandoc with appropriate options
	// 3. Returning the exported file path
	return nil, fmt.Errorf("export_document not yet implemented - requires Pandoc integration")
}