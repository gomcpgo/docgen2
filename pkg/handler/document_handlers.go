package handler

import (
	"context"
	"fmt"
	"strings"
	
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
	"github.com/savant/mcp-servers/docgen2/pkg/style"
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
		Blocks:      []document.BlockOverview{},    // Always initialize as empty array
		Chapters:    []document.ChapterOverview{},  // Always initialize as empty array
	}
	
	// Always process document-level blocks first (if they exist)
	if len(doc.Blocks) > 0 {
		overview.Blocks = h.buildBlockOverviews(docID, doc.Blocks)
	}
	
	// Then process chapters (if they exist)
	if len(doc.Chapters) > 0 {
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

// handleGetDocumentStyle returns the style configuration for a document
func (h *Handler) handleGetDocumentStyle(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	doc, err := h.storage.GetDocument(docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	
	// Return the style config if it exists, otherwise return default
	if doc.Style != nil {
		return jsonResponse(doc.Style)
	}
	
	// Return default style
	defaultStyle := h.storage.GetDefaultStyle()
	return jsonResponse(defaultStyle)
}

// handleUpdateDocumentStyle updates the style configuration for a document
func (h *Handler) handleUpdateDocumentStyle(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	// Get the document
	doc, err := h.storage.GetDocument(docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	
	// Parse the style configuration from args
	styleData, ok := args["style"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("style configuration is required")
	}
	
	// Update the document's style
	newStyle, err := h.parseStyleConfig(styleData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse style configuration: %w", err)
	}
	
	doc.Style = newStyle
	
	// Save the updated document
	if err := h.storage.SaveDocument(docID, doc); err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}
	
	return successResponse("Document style updated successfully"), nil
}

// parseStyleConfig parses style configuration from map to StyleConfig struct
func (h *Handler) parseStyleConfig(data map[string]interface{}) (*style.StyleConfig, error) {
	config := &style.StyleConfig{}
	
	// Parse fonts
	if fonts, ok := data["fonts"].(map[string]interface{}); ok {
		config.Fonts.BodyFamily = getStringFromMap(fonts, "body_family", "Times New Roman")
		config.Fonts.HeadingFamily = getStringFromMap(fonts, "heading_family", "Arial")
		config.Fonts.MonospaceFamily = getStringFromMap(fonts, "monospace_family", "Courier New")
		config.Fonts.BodySize = getIntFromMap(fonts, "body_size", 11)
		
		// Parse heading sizes
		if headingSizes, ok := fonts["heading_sizes"].(map[string]interface{}); ok {
			config.Fonts.HeadingSizes = make(map[string]int)
			for k, v := range headingSizes {
				if size, ok := v.(float64); ok {
					config.Fonts.HeadingSizes[k] = int(size)
				}
			}
		} else {
			// Use defaults
			config.Fonts.HeadingSizes = map[string]int{
				"h1": 20, "h2": 16, "h3": 14, "h4": 12, "h5": 11, "h6": 10,
			}
		}
	}
	
	// Parse colors
	if colors, ok := data["colors"].(map[string]interface{}); ok {
		config.Colors.BodyText = getStringFromMap(colors, "body_text", "0,0,0")
		config.Colors.HeadingText = getStringFromMap(colors, "heading_text", "0,0,0")
	}
	
	// Parse page config
	if page, ok := data["page"].(map[string]interface{}); ok {
		config.Page.Size = getStringFromMap(page, "size", "a4")
		config.Page.Orientation = getStringFromMap(page, "orientation", "portrait")
		
		// Parse margins
		if margins, ok := page["margins"].(map[string]interface{}); ok {
			config.Page.Margins.Top = getIntFromMap(margins, "top", 72)
			config.Page.Margins.Bottom = getIntFromMap(margins, "bottom", 72)
			config.Page.Margins.Left = getIntFromMap(margins, "left", 72)
			config.Page.Margins.Right = getIntFromMap(margins, "right", 72)
		}
	}
	
	// Parse spacing
	if spacing, ok := data["spacing"].(map[string]interface{}); ok {
		config.Spacing.LineSpacing = getFloatFromMap(spacing, "line_spacing", 1.2)
		config.Spacing.ParagraphSpacing = getIntFromMap(spacing, "paragraph_spacing", 6)
	}
	
	// Parse header
	if header, ok := data["header"].(map[string]interface{}); ok {
		config.Header.Enabled = getBoolFromMap(header, "enabled", true)
		config.Header.Content = getStringFromMap(header, "content", "{title}")
		config.Header.Align = getStringFromMap(header, "align", "center")
		config.Header.FontSize = getIntFromMap(header, "font_size", 10)
	}
	
	// Parse footer
	if footer, ok := data["footer"].(map[string]interface{}); ok {
		config.Footer.Enabled = getBoolFromMap(footer, "enabled", true)
		config.Footer.Content = getStringFromMap(footer, "content", "Page {page}")
		config.Footer.Align = getStringFromMap(footer, "align", "center")
		config.Footer.FontSize = getIntFromMap(footer, "font_size", 10)
	}
	
	return config, nil
}

// Helper functions for parsing map values
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultValue
}

func getIntFromMap(m map[string]interface{}, key string, defaultValue int) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return defaultValue
}

func getFloatFromMap(m map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return defaultValue
}

func getBoolFromMap(m map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultValue
}

