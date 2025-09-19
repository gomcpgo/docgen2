package handler

import (
	"context"
	"fmt"
	
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
)

// handleAddHeading adds a heading block
func (h *Handler) handleAddHeading(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	chapterID, _ := getString(args, "chapter_id", false)
	
	level, err := getInt(args, "level", 1)
	if err != nil {
		return nil, err
	}
	
	if level < 1 || level > 6 {
		return nil, fmt.Errorf("heading level must be between 1 and 6")
	}
	
	text, err := getString(args, "text", true)
	if err != nil {
		return nil, err
	}
	
	positionStr, _ := getString(args, "position", false)
	position := document.ParsePosition(positionStr)
	
	heading := &blocks.HeadingBlock{
		Level: level,
		Text:  text,
	}
	
	if err := h.storage.AddBlock(docID, chapterID, heading, position); err != nil {
		return nil, fmt.Errorf("failed to add heading: %w", err)
	}
	
	return successResponse(fmt.Sprintf("Added heading '%s' to document", text)), nil
}

// handleAddMarkdown adds a markdown block
func (h *Handler) handleAddMarkdown(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	chapterID, _ := getString(args, "chapter_id", false)
	
	content, err := getString(args, "content", true)
	if err != nil {
		return nil, err
	}
	
	positionStr, _ := getString(args, "position", false)
	position := document.ParsePosition(positionStr)
	
	markdown := &blocks.MarkdownBlock{
		Content: content,
	}
	
	if err := h.storage.AddBlock(docID, chapterID, markdown, position); err != nil {
		return nil, fmt.Errorf("failed to add markdown: %w", err)
	}
	
	return successResponse("Added markdown block to document"), nil
}

// handleAddImage adds an image block
func (h *Handler) handleAddImage(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	chapterID, _ := getString(args, "chapter_id", false)
	
	imagePath, err := getString(args, "image_path", true)
	if err != nil {
		return nil, err
	}
	
	caption, _ := getString(args, "caption", false)
	altText, _ := getString(args, "alt_text", false)
	
	positionStr, _ := getString(args, "position", false)
	position := document.ParsePosition(positionStr)
	
	// Copy image to assets folder
	assetPath, err := h.storage.CopyImageToAssets(docID, imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy image: %w", err)
	}
	
	image := &blocks.ImageBlock{
		Path:    assetPath,
		Caption: caption,
		AltText: altText,
	}
	
	if err := h.storage.AddBlock(docID, chapterID, image, position); err != nil {
		return nil, fmt.Errorf("failed to add image: %w", err)
	}
	
	return successResponse("Added image block to document"), nil
}

// handleAddTable adds a table block
func (h *Handler) handleAddTable(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	chapterID, _ := getString(args, "chapter_id", false)
	
	headers, err := getStringArray(args, "headers", true)
	if err != nil {
		return nil, err
	}
	
	rows, err := getStringArray2D(args, "rows", true)
	if err != nil {
		return nil, err
	}
	
	positionStr, _ := getString(args, "position", false)
	position := document.ParsePosition(positionStr)
	
	table := &blocks.TableBlock{
		Headers: headers,
		Rows:    rows,
	}
	
	if err := h.storage.AddBlock(docID, chapterID, table, position); err != nil {
		return nil, fmt.Errorf("failed to add table: %w", err)
	}
	
	return successResponse(fmt.Sprintf("Added table with %d columns and %d rows", len(headers), len(rows))), nil
}

// handleAddPageBreak adds a page break block
func (h *Handler) handleAddPageBreak(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	chapterID, _ := getString(args, "chapter_id", false)
	
	positionStr, _ := getString(args, "position", false)
	position := document.ParsePosition(positionStr)
	
	pageBreak := &blocks.PageBreakBlock{}
	
	if err := h.storage.AddBlock(docID, chapterID, pageBreak, position); err != nil {
		return nil, fmt.Errorf("failed to add page break: %w", err)
	}
	
	return successResponse("Added page break to document"), nil
}

// handleAddMultipleBlocks adds multiple blocks at once
func (h *Handler) handleAddMultipleBlocks(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	chapterID, _ := getString(args, "chapter_id", false)
	
	blocksData, ok := args["blocks"]
	if !ok {
		return nil, fmt.Errorf("blocks parameter is required")
	}
	
	blocksArray, ok := blocksData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("blocks must be an array")
	}
	
	positionStr, _ := getString(args, "position", false)
	position := document.ParsePosition(positionStr)
	
	addedCount := 0
	var lastBlockID string
	
	for _, blockData := range blocksArray {
		blockMap, ok := blockData.(map[string]interface{})
		if !ok {
			continue
		}
		
		blockType, _ := getString(blockMap, "type", true)
		data, ok := blockMap["data"].(map[string]interface{})
		if !ok {
			continue
		}
		
		var block blocks.Block
		
		switch blockType {
		case "heading":
			level, _ := getInt(data, "level", 1)
			text, _ := getString(data, "text", true)
			block = &blocks.HeadingBlock{
				Level: level,
				Text:  text,
			}
		
		case "markdown":
			content, _ := getString(data, "content", true)
			block = &blocks.MarkdownBlock{
				Content: content,
			}
		
		case "image":
			imagePath, _ := getString(data, "image_path", true)
			caption, _ := getString(data, "caption", false)
			altText, _ := getString(data, "alt_text", false)
			
			// Copy image to assets
			assetPath, err := h.storage.CopyImageToAssets(docID, imagePath)
			if err != nil {
				continue
			}
			
			block = &blocks.ImageBlock{
				Path:    assetPath,
				Caption: caption,
				AltText: altText,
			}
		
		case "table":
			headers, _ := getStringArray(data, "headers", true)
			rows, _ := getStringArray2D(data, "rows", true)
			block = &blocks.TableBlock{
				Headers: headers,
				Rows:    rows,
			}
		
		case "page_break":
			block = &blocks.PageBreakBlock{}
		
		default:
			continue
		}
		
		// For subsequent blocks, position them after the last added block
		if addedCount > 0 && lastBlockID != "" {
			position = document.Position{
				Type:    document.PositionAfter,
				BlockID: lastBlockID,
			}
		}
		
		if err := h.storage.AddBlock(docID, chapterID, block, position); err == nil {
			addedCount++
			// Get the ID of the block we just added (would need to modify AddBlock to return it)
			// For now, we'll continue with the original position
		}
	}
	
	return successResponse(fmt.Sprintf("Added %d blocks to document", addedCount)), nil
}

// handleUpdateBlock updates an existing block
func (h *Handler) handleUpdateBlock(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// This would require more complex implementation to:
	// 1. Load the existing block
	// 2. Update its content
	// 3. Save it back
	// For now, returning not implemented
	return nil, fmt.Errorf("update_block not yet implemented")
}

// handleDeleteBlock deletes a block
func (h *Handler) handleDeleteBlock(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// This would require implementation in storage layer to:
	// 1. Find the block in document/chapter
	// 2. Remove it from the manifest
	// 3. Delete the block file
	return nil, fmt.Errorf("delete_block not yet implemented")
}

// handleMoveBlock moves a block to a new position
func (h *Handler) handleMoveBlock(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// This would require implementation to:
	// 1. Find the block in document/chapter
	// 2. Remove it from current position
	// 3. Insert it at new position
	return nil, fmt.Errorf("move_block not yet implemented")
}

// handleGetBlock gets a specific block's content
func (h *Handler) handleGetBlock(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := getString(args, "document_id", true)
	if err != nil {
		return nil, err
	}
	
	blockID, err := getString(args, "block_id", true)
	if err != nil {
		return nil, err
	}
	
	// Find the block in the document
	doc, err := h.storage.GetDocument(docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	
	var blockRef *blocks.BlockReference
	var found bool
	
	if doc.HasChapters {
		// Search in chapters
		for _, chapterRef := range doc.Chapters {
			chapter, err := h.storage.GetChapter(docID, chapterRef.ID)
			if err != nil {
				continue
			}
			
			for _, ref := range chapter.Blocks {
				if ref.ID == blockID {
					blockRef = &ref
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	} else {
		// Search in flat document
		for _, ref := range doc.Blocks {
			if ref.ID == blockID {
				blockRef = &ref
				found = true
				break
			}
		}
	}
	
	if !found {
		return nil, fmt.Errorf("block not found: %s", blockID)
	}
	
	// Load the block
	block, err := h.storage.LoadBlock(docID, *blockRef)
	if err != nil {
		return nil, fmt.Errorf("failed to load block: %w", err)
	}
	
	// Convert block to response format
	var result map[string]interface{}
	
	switch b := block.(type) {
	case *blocks.HeadingBlock:
		result = map[string]interface{}{
			"id":    b.ID,
			"type":  "heading",
			"level": b.Level,
			"text":  b.Text,
		}
	case *blocks.MarkdownBlock:
		result = map[string]interface{}{
			"id":      b.ID,
			"type":    "markdown",
			"content": b.Content,
		}
	case *blocks.ImageBlock:
		result = map[string]interface{}{
			"id":       b.ID,
			"type":     "image",
			"path":     b.Path,
			"caption":  b.Caption,
			"alt_text": b.AltText,
		}
	case *blocks.TableBlock:
		result = map[string]interface{}{
			"id":      b.ID,
			"type":    "table",
			"headers": b.Headers,
			"rows":    b.Rows,
		}
	case *blocks.PageBreakBlock:
		result = map[string]interface{}{
			"id":   b.ID,
			"type": "page_break",
		}
	}
	
	return jsonResponse(result)
}