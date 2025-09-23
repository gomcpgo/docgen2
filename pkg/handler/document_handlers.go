package handler

import (
	"context"
	"fmt"
	"strings"
	
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
)

// handleCreateDocument creates a new document
func (h *Handler) handleCreateDocument(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	title, err := getString(args, "title", true)
	if err != nil {
		return nil, err
	}
	
	hasChapters := getBool(args, "has_chapters", false)
	author, _ := getString(args, "author", false)
	
	docID, err := h.storage.CreateDocument(title, hasChapters, author)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}
	
	result := map[string]interface{}{
		"document_id":  docID,
		"title":        title,
		"has_chapters": hasChapters,
		"author":       author,
	}
	
	return jsonResponse(result)
}

// handleListDocuments lists all available documents
func (h *Handler) handleListDocuments(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docIDs, err := h.storage.ListDocuments()
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	
	// Get details for each document
	var documents []map[string]interface{}
	for _, docID := range docIDs {
		doc, err := h.storage.GetDocument(docID)
		if err != nil {
			continue // Skip documents that can't be loaded
		}
		
		documents = append(documents, map[string]interface{}{
			"id":          docID,
			"title":       doc.Title,
			"author":      doc.Author,
			"has_chapters": doc.HasChapters,
			"created_at":  doc.CreatedAt,
			"updated_at":  doc.UpdatedAt,
		})
	}
	
	return jsonResponse(documents)
}

// handleGetDocumentOverview returns a hierarchical overview of a document
func (h *Handler) handleGetDocumentOverview(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	doc, err := h.storage.GetDocument(docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	
	overview := document.DocumentOverview{
		ID:          docID,
		Title:       doc.Title,
		Author:      doc.Author,
		HasChapters: doc.HasChapters,
	}
	
	if doc.HasChapters && len(doc.Chapters) > 0 {
		// Build chapter overview
		for _, chapterRef := range doc.Chapters {
			chapter, err := h.storage.GetChapter(docID, chapterRef.ID)
			if err != nil {
				continue
			}
			
			chapterOverview := document.ChapterOverview{
				ID:     chapter.ID,
				Title:  chapter.Title,
				Blocks: h.buildBlockOverviews(docID, chapter.Blocks),
			}
			overview.Chapters = append(overview.Chapters, chapterOverview)
		}
	}
	
	// Always include flat blocks if they exist, even in chaptered documents
	// This handles edge cases where has_chapters is true but no actual chapters exist
	if len(doc.Blocks) > 0 {
		overview.Blocks = h.buildBlockOverviews(docID, doc.Blocks)
		
		// Log a warning if we have both chapters and flat blocks, or has_chapters is true with no chapters
		if doc.HasChapters && len(doc.Chapters) == 0 {
			// Document is marked as chaptered but has no chapters - this is a data inconsistency
			// We'll still return the flat blocks so the document is usable
			fmt.Printf("[WARNING] Document %s has has_chapters=true but no chapters, using flat blocks\n", docID)
		} else if len(overview.Chapters) > 0 {
			// Document has both chapters and flat blocks - unusual but allowed
			fmt.Printf("[WARNING] Document %s has both chapters and flat blocks\n", docID)
		}
	}
	
	return jsonResponse(overview)
}

// buildBlockOverviews creates overview for blocks
func (h *Handler) buildBlockOverviews(docID string, blockRefs []blocks.BlockReference) []document.BlockOverview {
	var overviews []document.BlockOverview
	
	for _, ref := range blockRefs {
		block, err := h.storage.LoadBlock(docID, ref)
		if err != nil {
			continue
		}
		
		preview := h.getBlockPreview(block)
		overviews = append(overviews, document.BlockOverview{
			ID:      ref.ID,
			Type:    string(ref.Type),
			Preview: preview,
		})
	}
	
	return overviews
}

// getBlockPreview generates a preview string for a block
func (h *Handler) getBlockPreview(block blocks.Block) string {
	switch b := block.(type) {
	case *blocks.HeadingBlock:
		return fmt.Sprintf("H%d: %s", b.Level, truncateString(b.Text, 100))
	case *blocks.MarkdownBlock:
		return truncateString(b.Content, 100)
	case *blocks.ImageBlock:
		preview := "Image"
		if b.Caption != "" {
			preview += ": " + truncateString(b.Caption, 80)
		}
		return preview
	case *blocks.TableBlock:
		return fmt.Sprintf("Table: %d columns, %d rows", len(b.Headers), len(b.Rows))
	case *blocks.PageBreakBlock:
		return "Page Break"
	default:
		return "Unknown block type"
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// handleDeleteDocument deletes a document
func (h *Handler) handleDeleteDocument(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	if err := h.storage.DeleteDocument(docID); err != nil {
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}
	
	return successResponse(fmt.Sprintf("Document '%s' deleted successfully", docID)), nil
}

// handleSearchBlocks searches for blocks containing specific text
func (h *Handler) handleSearchBlocks(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	query, err := getString(args, "query", true)
	if err != nil {
		return nil, err
	}
	
	chapterID, _ := getString(args, "chapter_id", false)
	
	// Use searcher to find results
	results, err := h.searcher.SearchDocument(docID, query, chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to search document: %w", err)
	}
	
	return jsonResponse(results)
}

