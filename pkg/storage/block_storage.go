package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	
	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
	"gopkg.in/yaml.v3"
)

// AddBlock adds a new block to a document or chapter
func (s *Storage) AddBlock(docID string, chapterID string, block blocks.Block, position document.Position) error {
	doc, err := s.GetDocument(docID)
	if err != nil {
		return err
	}
	
	// Generate block ID
	blockID := s.generateBlockID(docID, chapterID, block.GetType())
	
	// Set the block ID
	switch b := block.(type) {
	case *blocks.HeadingBlock:
		b.ID = blockID
	case *blocks.MarkdownBlock:
		b.ID = blockID
	case *blocks.ImageBlock:
		b.ID = blockID
	case *blocks.TableBlock:
		b.ID = blockID
	case *blocks.PageBreakBlock:
		b.ID = blockID
	}
	
	// Save block file
	blockFile, err := s.saveBlockFile(docID, chapterID, block)
	if err != nil {
		return err
	}
	
	// Create block reference
	blockRef := blocks.BlockReference{
		ID:   blockID,
		Type: block.GetType(),
		File: blockFile,
	}
	
	// Add to document manifest
	if doc.HasChapters && chapterID != "" {
		// Add to chapter
		chapter, err := s.GetChapter(docID, chapterID)
		if err != nil {
			return err
		}
		
		chapter.Blocks = s.insertBlockAtPosition(chapter.Blocks, blockRef, position)
		
		if err := s.SaveChapter(docID, chapterID, chapter); err != nil {
			return err
		}
	} else {
		// Add to flat document
		doc.Blocks = s.insertBlockAtPosition(doc.Blocks, blockRef, position)
	}
	
	return s.SaveDocument(docID, doc)
}

// generateBlockID generates a unique block ID
func (s *Storage) generateBlockID(docID, chapterID string, blockType blocks.BlockType) string {
	// Determine prefix based on type
	var prefix string
	switch blockType {
	case blocks.TypeHeading:
		prefix = "hd"
	case blocks.TypeMarkdown:
		prefix = "md"
	case blocks.TypeImage:
		prefix = "img"
	case blocks.TypeTable:
		prefix = "tbl"
	case blocks.TypePageBreak:
		prefix = "pb"
	default:
		prefix = "blk"
	}
	
	// Count existing blocks of this type
	doc, err := s.GetDocument(docID)
	if err != nil {
		return fmt.Sprintf("%s-001", prefix)
	}
	
	maxNum := 0
	re := regexp.MustCompile(fmt.Sprintf(`^%s-(\d+)$`, prefix))
	
	if doc.HasChapters && chapterID != "" {
		chapter, err := s.GetChapter(docID, chapterID)
		if err == nil {
			for _, ref := range chapter.Blocks {
				if matches := re.FindStringSubmatch(ref.ID); matches != nil {
					var num int
					fmt.Sscanf(matches[1], "%d", &num)
					if num > maxNum {
						maxNum = num
					}
				}
			}
		}
	} else {
		for _, ref := range doc.Blocks {
			if matches := re.FindStringSubmatch(ref.ID); matches != nil {
				var num int
				fmt.Sscanf(matches[1], "%d", &num)
				if num > maxNum {
					maxNum = num
				}
			}
		}
	}
	
	return fmt.Sprintf("%s-%03d", prefix, maxNum+1)
}

// insertBlockAtPosition inserts a block at the specified position
func (s *Storage) insertBlockAtPosition(blockList []blocks.BlockReference, newBlock blocks.BlockReference, position document.Position) []blocks.BlockReference {
	switch position.Type {
	case document.PositionStart:
		return append([]blocks.BlockReference{newBlock}, blockList...)
	
	case document.PositionAfter:
		for i, block := range blockList {
			if block.ID == position.BlockID {
				// Insert after this block
				result := make([]blocks.BlockReference, 0, len(blockList)+1)
				result = append(result, blockList[:i+1]...)
				result = append(result, newBlock)
				result = append(result, blockList[i+1:]...)
				return result
			}
		}
		// If block not found, append to end
		return append(blockList, newBlock)
	
	case document.PositionEnd:
		fallthrough
	default:
		return append(blockList, newBlock)
	}
}

// saveBlockFile saves a block to a file and returns the relative path
func (s *Storage) saveBlockFile(docID, chapterID string, block blocks.Block) (string, error) {
	var basePath string
	if chapterID != "" {
		basePath = filepath.Join(s.config.GetDocumentFolder(docID), "chapters", chapterID, "blocks")
	} else {
		basePath = filepath.Join(s.config.GetDocumentFolder(docID), "blocks")
	}
	
	// Ensure blocks directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", err
	}
	
	var relativePath string
	var fullPath string
	
	switch b := block.(type) {
	case *blocks.HeadingBlock:
		filename := fmt.Sprintf("%s-heading.yaml", b.ID)
		fullPath = filepath.Join(basePath, filename)
		
		data, err := yaml.Marshal(b)
		if err != nil {
			return "", err
		}
		
		if err := os.WriteFile(fullPath, data, 0644); err != nil {
			return "", err
		}
		
		if chapterID != "" {
			relativePath = filepath.Join("chapters", chapterID, "blocks", filename)
		} else {
			relativePath = filepath.Join("blocks", filename)
		}
	
	case *blocks.MarkdownBlock:
		filename := fmt.Sprintf("%s.md", b.ID)
		fullPath = filepath.Join(basePath, filename)
		
		if err := os.WriteFile(fullPath, []byte(b.Content), 0644); err != nil {
			return "", err
		}
		
		if chapterID != "" {
			relativePath = filepath.Join("chapters", chapterID, "blocks", filename)
		} else {
			relativePath = filepath.Join("blocks", filename)
		}
	
	case *blocks.ImageBlock:
		filename := fmt.Sprintf("%s-image.yaml", b.ID)
		fullPath = filepath.Join(basePath, filename)
		
		data, err := yaml.Marshal(b)
		if err != nil {
			return "", err
		}
		
		if err := os.WriteFile(fullPath, data, 0644); err != nil {
			return "", err
		}
		
		if chapterID != "" {
			relativePath = filepath.Join("chapters", chapterID, "blocks", filename)
		} else {
			relativePath = filepath.Join("blocks", filename)
		}
	
	case *blocks.TableBlock:
		filename := fmt.Sprintf("%s-table.yaml", b.ID)
		fullPath = filepath.Join(basePath, filename)
		
		data, err := yaml.Marshal(b)
		if err != nil {
			return "", err
		}
		
		if err := os.WriteFile(fullPath, data, 0644); err != nil {
			return "", err
		}
		
		if chapterID != "" {
			relativePath = filepath.Join("chapters", chapterID, "blocks", filename)
		} else {
			relativePath = filepath.Join("blocks", filename)
		}
		
	case *blocks.PageBreakBlock:
		filename := fmt.Sprintf("%s-pagebreak.yaml", b.ID)
		fullPath = filepath.Join(basePath, filename)
		
		data, err := yaml.Marshal(b)
		if err != nil {
			return "", err
		}
		
		if err := os.WriteFile(fullPath, data, 0644); err != nil {
			return "", err
		}
		
		if chapterID != "" {
			relativePath = filepath.Join("chapters", chapterID, "blocks", filename)
		} else {
			relativePath = filepath.Join("blocks", filename)
		}
		
	default:
		return "", fmt.Errorf("unknown block type: %T", block)
	}
	
	return relativePath, nil
}

// LoadBlock loads a block from storage
func (s *Storage) LoadBlock(docID string, blockRef blocks.BlockReference) (blocks.Block, error) {
	filePath := filepath.Join(s.config.GetDocumentFolder(docID), blockRef.File)
	
	switch blockRef.Type {
	case blocks.TypeHeading:
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		
		var heading blocks.HeadingBlock
		if err := yaml.Unmarshal(data, &heading); err != nil {
			return nil, err
		}
		return &heading, nil
		
	case blocks.TypeMarkdown:
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		
		return &blocks.MarkdownBlock{
			BaseBlock: blocks.BaseBlock{ID: blockRef.ID},
			Content:   string(content),
		}, nil
		
	case blocks.TypeImage:
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		
		var image blocks.ImageBlock
		if err := yaml.Unmarshal(data, &image); err != nil {
			return nil, err
		}
		return &image, nil
		
	case blocks.TypeTable:
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		
		var table blocks.TableBlock
		if err := yaml.Unmarshal(data, &table); err != nil {
			return nil, err
		}
		return &table, nil
		
	case blocks.TypePageBreak:
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		
		var pageBreak blocks.PageBreakBlock
		if err := yaml.Unmarshal(data, &pageBreak); err != nil {
			return nil, err
		}
		return &pageBreak, nil
		
	default:
		return nil, fmt.Errorf("unknown block type: %s", blockRef.Type)
	}
}

// UpdateBlock updates an existing block
func (s *Storage) UpdateBlock(docID, blockID string, newBlock blocks.Block) error {
	// Find the block's location
	chapterID, blockIndex, err := s.FindBlockLocation(docID, blockID)
	if err != nil {
		return err
	}
	
	// Get the document
	doc, err := s.GetDocument(docID)
	if err != nil {
		return err
	}
	
	// Get the block reference
	var blockRef *blocks.BlockReference
	if chapterID != "" {
		chapter, err := s.GetChapter(docID, chapterID)
		if err != nil {
			return err
		}
		if blockIndex < 0 || blockIndex >= len(chapter.Blocks) {
			return fmt.Errorf("block index out of range")
		}
		blockRef = &chapter.Blocks[blockIndex]
	} else {
		if blockIndex < 0 || blockIndex >= len(doc.Blocks) {
			return fmt.Errorf("block index out of range")
		}
		blockRef = &doc.Blocks[blockIndex]
	}
	
	// Verify the block type matches
	if blockRef.Type != newBlock.GetType() {
		return fmt.Errorf("cannot change block type from %s to %s", blockRef.Type, newBlock.GetType())
	}
	
	// Set the block ID to maintain consistency
	switch b := newBlock.(type) {
	case *blocks.HeadingBlock:
		b.ID = blockID
	case *blocks.MarkdownBlock:
		b.ID = blockID
	case *blocks.ImageBlock:
		b.ID = blockID
	case *blocks.TableBlock:
		b.ID = blockID
	case *blocks.PageBreakBlock:
		b.ID = blockID
	}
	
	// Save the updated block
	_, err = s.saveBlockFile(docID, chapterID, newBlock)
	if err != nil {
		return err
	}
	
	// Update document timestamp
	return s.SaveDocument(docID, doc)
}

// DeleteBlock deletes a block from the document
func (s *Storage) DeleteBlock(docID, blockID string) error {
	// Find the block's location
	chapterID, blockIndex, err := s.FindBlockLocation(docID, blockID)
	if err != nil {
		return err
	}
	
	// Get the document
	doc, err := s.GetDocument(docID)
	if err != nil {
		return err
	}
	
	// Remove the block from the appropriate list
	if chapterID != "" {
		// Remove from chapter
		chapter, err := s.GetChapter(docID, chapterID)
		if err != nil {
			return err
		}
		
		if blockIndex < 0 || blockIndex >= len(chapter.Blocks) {
			return fmt.Errorf("block index out of range")
		}
		
		// Delete the block file
		blockPath := filepath.Join(s.config.GetDocumentFolder(docID), chapter.Blocks[blockIndex].File)
		os.Remove(blockPath) // Ignore error if file doesn't exist
		
		// Remove from blocks list
		chapter.Blocks = append(chapter.Blocks[:blockIndex], chapter.Blocks[blockIndex+1:]...)
		
		// Save updated chapter
		if err := s.SaveChapter(docID, chapterID, chapter); err != nil {
			return err
		}
	} else {
		// Remove from flat document
		if blockIndex < 0 || blockIndex >= len(doc.Blocks) {
			return fmt.Errorf("block index out of range")
		}
		
		// Delete the block file
		blockPath := filepath.Join(s.config.GetDocumentFolder(docID), doc.Blocks[blockIndex].File)
		os.Remove(blockPath) // Ignore error if file doesn't exist
		
		// Remove from blocks list
		doc.Blocks = append(doc.Blocks[:blockIndex], doc.Blocks[blockIndex+1:]...)
	}
	
	// Update document timestamp
	return s.SaveDocument(docID, doc)
}

// MoveBlock moves a block to a new position
func (s *Storage) MoveBlock(docID, blockID string, newPosition document.Position) error {
	// Find the block's current location
	chapterID, blockIndex, err := s.FindBlockLocation(docID, blockID)
	if err != nil {
		return err
	}
	
	// Get the document
	doc, err := s.GetDocument(docID)
	if err != nil {
		return err
	}
	
	var blockRef blocks.BlockReference
	
	if chapterID != "" {
		// Moving within a chapter
		chapter, err := s.GetChapter(docID, chapterID)
		if err != nil {
			return err
		}
		
		if blockIndex < 0 || blockIndex >= len(chapter.Blocks) {
			return fmt.Errorf("block index out of range")
		}
		
		// Get the block reference
		blockRef = chapter.Blocks[blockIndex]
		
		// Remove block from current position
		remainingBlocks := append(chapter.Blocks[:blockIndex], chapter.Blocks[blockIndex+1:]...)
		
		// Insert at new position
		chapter.Blocks = s.insertBlockAtPosition(remainingBlocks, blockRef, newPosition)
		
		// Save updated chapter
		if err := s.SaveChapter(docID, chapterID, chapter); err != nil {
			return err
		}
	} else {
		// Moving within flat document
		if blockIndex < 0 || blockIndex >= len(doc.Blocks) {
			return fmt.Errorf("block index out of range")
		}
		
		// Get the block reference
		blockRef = doc.Blocks[blockIndex]
		
		// Remove block from current position
		remainingBlocks := append(doc.Blocks[:blockIndex], doc.Blocks[blockIndex+1:]...)
		
		// Insert at new position
		doc.Blocks = s.insertBlockAtPosition(remainingBlocks, blockRef, newPosition)
	}
	
	// Update document timestamp
	return s.SaveDocument(docID, doc)
}

// FindBlockLocation finds the location of a block (which chapter it's in)
func (s *Storage) FindBlockLocation(docID, blockID string) (chapterID string, blockIndex int, err error) {
	doc, err := s.GetDocument(docID)
	if err != nil {
		return "", -1, err
	}
	
	if doc.HasChapters {
		// Search in chapters
		for _, chapterRef := range doc.Chapters {
			chapter, err := s.GetChapter(docID, chapterRef.ID)
			if err != nil {
				continue
			}
			
			for i, ref := range chapter.Blocks {
				if ref.ID == blockID {
					return chapterRef.ID, i, nil
				}
			}
		}
	} else {
		// Search in flat document
		for i, ref := range doc.Blocks {
			if ref.ID == blockID {
				return "", i, nil
			}
		}
	}
	
	return "", -1, fmt.Errorf("block not found: %s", blockID)
}