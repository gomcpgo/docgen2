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
	
	if doc.HasChapters {
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
	} else {
		// Build flat document overview
		overview.Blocks = h.buildBlockOverviews(docID, doc.Blocks)
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
	
	doc, err := h.storage.GetDocument(docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	
	var results []document.SearchResult
	queryLower := strings.ToLower(query)
	
	if doc.HasChapters {
		// Search in chapters
		for _, chapterRef := range doc.Chapters {
			// Skip if specific chapter requested and this isn't it
			if chapterID != "" && chapterRef.ID != chapterID {
				continue
			}
			
			chapter, err := h.storage.GetChapter(docID, chapterRef.ID)
			if err != nil {
				continue
			}
			
			results = append(results, h.searchInBlocks(docID, chapter.Blocks, queryLower, chapterRef.ID)...)
		}
	} else {
		// Search in flat document
		results = append(results, h.searchInBlocks(docID, doc.Blocks, queryLower, "")...)
	}
	
	return jsonResponse(results)
}

// searchInBlocks searches for query in blocks
func (h *Handler) searchInBlocks(docID string, blockRefs []blocks.BlockReference, query string, chapterID string) []document.SearchResult {
	var results []document.SearchResult
	
	for i, ref := range blockRefs {
		block, err := h.storage.LoadBlock(docID, ref)
		if err != nil {
			continue
		}
		
		content := h.getBlockContent(block)
		if strings.Contains(strings.ToLower(content), query) {
			// Find snippet around the match
			snippet := h.extractSnippet(content, query)
			
			result := document.SearchResult{
				BlockID:   ref.ID,
				BlockType: string(ref.Type),
				ChapterID: chapterID,
				Snippet:   snippet,
				Position:  i,
			}
			results = append(results, result)
		}
	}
	
	return results
}

// getBlockContent extracts searchable text from a block
func (h *Handler) getBlockContent(block blocks.Block) string {
	switch b := block.(type) {
	case *blocks.HeadingBlock:
		return b.Text
	case *blocks.MarkdownBlock:
		return b.Content
	case *blocks.ImageBlock:
		return b.Caption + " " + b.AltText
	case *blocks.TableBlock:
		// Combine headers and all cells
		content := strings.Join(b.Headers, " ")
		for _, row := range b.Rows {
			content += " " + strings.Join(row, " ")
		}
		return content
	default:
		return ""
	}
}

// extractSnippet extracts a snippet around the query match
func (h *Handler) extractSnippet(content, query string) string {
	contentLower := strings.ToLower(content)
	index := strings.Index(contentLower, strings.ToLower(query))
	if index < 0 {
		return truncateString(content, 150)
	}
	
	// Get 50 chars before and after the match
	start := index - 50
	if start < 0 {
		start = 0
	}
	
	end := index + len(query) + 50
	if end > len(content) {
		end = len(content)
	}
	
	snippet := content[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}
	
	return snippet
}