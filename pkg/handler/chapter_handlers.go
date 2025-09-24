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
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	chapterID, err := getString(args, "chapter_id", true)
	if err != nil {
		return nil, err
	}
	
	newTitle, err := getString(args, "new_title", true)
	if err != nil {
		return nil, err
	}
	
	if err := h.storage.UpdateChapter(docID, chapterID, newTitle); err != nil {
		return nil, fmt.Errorf("failed to update chapter: %w", err)
	}
	
	result := map[string]interface{}{
		"chapter_id": chapterID,
		"new_title":  newTitle,
		"message":    fmt.Sprintf("Updated chapter '%s'", newTitle),
	}
	
	return jsonResponse(result)
}

// handleDeleteChapter deletes a chapter
func (h *Handler) handleDeleteChapter(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	chapterID, err := getString(args, "chapter_id", true)
	if err != nil {
		return nil, err
	}
	
	if err := h.storage.DeleteChapter(docID, chapterID); err != nil {
		return nil, fmt.Errorf("failed to delete chapter: %w", err)
	}
	
	result := map[string]interface{}{
		"chapter_id": chapterID,
		"message":    fmt.Sprintf("Deleted chapter %s", chapterID),
	}
	
	return jsonResponse(result)
}

// handleMoveChapter moves a chapter to a new position
func (h *Handler) handleMoveChapter(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	chapterID, err := getString(args, "chapter_id", true)
	if err != nil {
		return nil, err
	}
	
	positionStr, err := getString(args, "new_position", true)
	if err != nil {
		return nil, err
	}
	
	position := document.ParsePosition(positionStr)
	
	if err := h.storage.MoveChapter(docID, chapterID, position); err != nil {
		return nil, fmt.Errorf("failed to move chapter: %w", err)
	}
	
	result := map[string]interface{}{
		"chapter_id":   chapterID,
		"new_position": positionStr,
		"message":      fmt.Sprintf("Moved chapter %s to position %s", chapterID, positionStr),
	}
	
	return jsonResponse(result)
}

// handleExportDocument exports a document to various formats
func (h *Handler) handleExportDocument(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	format, err := getString(args, "format", true)
	if err != nil {
		return nil, err
	}
	
	// Export the document
	outputPath, err := h.exporter.ExportDocument(docID, format)
	if err != nil {
		return nil, fmt.Errorf("failed to export document: %w", err)
	}
	
	result := map[string]interface{}{
		"document_id": docID,
		"format":      format,
		"output_path": outputPath,
		"message":     fmt.Sprintf("Document exported successfully to %s", outputPath),
	}
	
	return jsonResponse(result)
}