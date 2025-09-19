package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	
	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/config"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
	"gopkg.in/yaml.v3"
)

// Storage handles all file operations for documents
type Storage struct {
	config *config.Config
}

// NewStorage creates a new storage instance
func NewStorage(cfg *config.Config) *Storage {
	return &Storage{config: cfg}
}

// CreateDocument creates a new document
func (s *Storage) CreateDocument(title string, hasChapters bool, author string) (string, error) {
	// Generate document ID from title
	docID := s.generateDocumentID(title)
	
	// Create document folder
	docPath := s.config.GetDocumentFolder(docID)
	if err := os.MkdirAll(docPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create document folder: %w", err)
	}
	
	// Create assets folder
	assetsPath := filepath.Join(docPath, "assets")
	if err := os.MkdirAll(assetsPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create assets folder: %w", err)
	}
	
	// Create blocks folder for flat documents or chapters folder for chaptered documents
	if !hasChapters {
		blocksPath := filepath.Join(docPath, "blocks")
		if err := os.MkdirAll(blocksPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create blocks folder: %w", err)
		}
	} else {
		chaptersPath := filepath.Join(docPath, "chapters")
		if err := os.MkdirAll(chaptersPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create chapters folder: %w", err)
		}
	}
	
	// Create manifest
	doc := &document.Document{
		Title:       title,
		Author:      author,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		HasChapters: hasChapters,
		Blocks:      []blocks.BlockReference{},
		Chapters:    []document.ChapterReference{},
	}
	
	// Save manifest
	if err := s.SaveDocument(docID, doc); err != nil {
		// Clean up on error
		os.RemoveAll(docPath)
		return "", fmt.Errorf("failed to save manifest: %w", err)
	}
	
	return docID, nil
}

// generateDocumentID generates a unique document ID from the title
func (s *Storage) generateDocumentID(title string) string {
	// Clean title for filesystem
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)
	cleaned := reg.ReplaceAllString(strings.ToLower(title), "-")
	cleaned = strings.Trim(cleaned, "-")
	
	// Truncate if too long
	if len(cleaned) > 50 {
		cleaned = cleaned[:50]
	}
	
	// Check for uniqueness and add number if needed
	baseID := cleaned
	counter := 1
	for {
		docPath := s.config.GetDocumentFolder(cleaned)
		if _, err := os.Stat(docPath); os.IsNotExist(err) {
			break
		}
		cleaned = fmt.Sprintf("%s-%d", baseID, counter)
		counter++
	}
	
	return cleaned
}

// GetDocument loads a document manifest
func (s *Storage) GetDocument(docID string) (*document.Document, error) {
	manifestPath := filepath.Join(s.config.GetDocumentFolder(docID), "manifest.yaml")
	
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("document not found: %s", docID)
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	
	var doc document.Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	
	return &doc, nil
}

// SaveDocument saves a document manifest
func (s *Storage) SaveDocument(docID string, doc *document.Document) error {
	doc.UpdatedAt = time.Now()
	
	data, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}
	
	manifestPath := filepath.Join(s.config.GetDocumentFolder(docID), "manifest.yaml")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}
	
	return nil
}

// ListDocuments returns a list of all document IDs
func (s *Storage) ListDocuments() ([]string, error) {
	docsPath := s.config.GetDocumentsFolder()
	
	entries, err := os.ReadDir(docsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read documents folder: %w", err)
	}
	
	var docIDs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it has a manifest
			manifestPath := filepath.Join(docsPath, entry.Name(), "manifest.yaml")
			if _, err := os.Stat(manifestPath); err == nil {
				docIDs = append(docIDs, entry.Name())
			}
		}
	}
	
	return docIDs, nil
}

// DeleteDocument deletes a document and all its files
func (s *Storage) DeleteDocument(docID string) error {
	docPath := s.config.GetDocumentFolder(docID)
	
	// Check if document exists
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		return fmt.Errorf("document not found: %s", docID)
	}
	
	// Remove entire document folder
	if err := os.RemoveAll(docPath); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	
	return nil
}

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
	
	// Get existing blocks to find next number
	var existingBlocks []blocks.BlockReference
	
	doc, _ := s.GetDocument(docID)
	if doc != nil {
		if doc.HasChapters && chapterID != "" {
			if chapter, err := s.GetChapter(docID, chapterID); err == nil {
				existingBlocks = chapter.Blocks
			}
		} else {
			existingBlocks = doc.Blocks
		}
	}
	
	// Find highest number for this type
	maxNum := 0
	for _, ref := range existingBlocks {
		if ref.Type == blockType {
			// Extract number from ID like "md-001"
			parts := strings.Split(ref.ID, "-")
			if len(parts) == 2 {
				var num int
				fmt.Sscanf(parts[1], "%d", &num)
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
		// If block not found, add to end
		return append(blockList, newBlock)
	
	default: // PositionEnd
		return append(blockList, newBlock)
	}
}

// saveBlockFile saves a block to a file and returns the relative file path
func (s *Storage) saveBlockFile(docID, chapterID string, block blocks.Block) (string, error) {
	var basePath string
	var relPath string
	
	if chapterID != "" {
		basePath = filepath.Join(s.config.GetDocumentFolder(docID), "chapters", chapterID, "blocks")
		relPath = filepath.Join("chapters", chapterID, "blocks")
	} else {
		basePath = filepath.Join(s.config.GetDocumentFolder(docID), "blocks")
		relPath = "blocks"
	}
	
	// Ensure blocks directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create blocks directory: %w", err)
	}
	
	switch b := block.(type) {
	case *blocks.MarkdownBlock:
		// Save markdown content to .md file
		filename := fmt.Sprintf("%s.md", b.ID)
		filePath := filepath.Join(basePath, filename)
		if err := os.WriteFile(filePath, []byte(b.Content), 0644); err != nil {
			return "", fmt.Errorf("failed to write markdown file: %w", err)
		}
		return filepath.Join(relPath, filename), nil
	
	case *blocks.ImageBlock:
		// Save image metadata to YAML
		filename := fmt.Sprintf("%s-image.yaml", b.ID)
		filePath := filepath.Join(basePath, filename)
		data, err := yaml.Marshal(b)
		if err != nil {
			return "", fmt.Errorf("failed to marshal image block: %w", err)
		}
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return "", fmt.Errorf("failed to write image metadata: %w", err)
		}
		return filepath.Join(relPath, filename), nil
	
	default:
		// Save other blocks as YAML
		var filename string
		switch block.GetType() {
		case blocks.TypeHeading:
			filename = fmt.Sprintf("%s-heading.yaml", block.GetID())
		case blocks.TypeTable:
			filename = fmt.Sprintf("%s-table.yaml", block.GetID())
		case blocks.TypePageBreak:
			filename = fmt.Sprintf("%s-pagebreak.yaml", block.GetID())
		default:
			filename = fmt.Sprintf("%s.yaml", block.GetID())
		}
		
		filePath := filepath.Join(basePath, filename)
		data, err := yaml.Marshal(block)
		if err != nil {
			return "", fmt.Errorf("failed to marshal block: %w", err)
		}
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return "", fmt.Errorf("failed to write block file: %w", err)
		}
		return filepath.Join(relPath, filename), nil
	}
}

// GetChapter loads a chapter
func (s *Storage) GetChapter(docID, chapterID string) (*document.Chapter, error) {
	doc, err := s.GetDocument(docID)
	if err != nil {
		return nil, err
	}
	
	if !doc.HasChapters {
		return nil, fmt.Errorf("document does not have chapters")
	}
	
	var chapterRef *document.ChapterReference
	for _, ch := range doc.Chapters {
		if ch.ID == chapterID {
			chapterRef = &ch
			break
		}
	}
	
	if chapterRef == nil {
		return nil, fmt.Errorf("chapter not found: %s", chapterID)
	}
	
	// Load chapter manifest
	chapterPath := filepath.Join(s.config.GetDocumentFolder(docID), chapterRef.Folder, "chapter.yaml")
	data, err := os.ReadFile(chapterPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chapter manifest: %w", err)
	}
	
	var chapter document.Chapter
	if err := yaml.Unmarshal(data, &chapter); err != nil {
		return nil, fmt.Errorf("failed to parse chapter manifest: %w", err)
	}
	
	return &chapter, nil
}

// SaveChapter saves a chapter
func (s *Storage) SaveChapter(docID, chapterID string, chapter *document.Chapter) error {
	doc, err := s.GetDocument(docID)
	if err != nil {
		return err
	}
	
	var chapterRef *document.ChapterReference
	for _, ch := range doc.Chapters {
		if ch.ID == chapterID {
			chapterRef = &ch
			break
		}
	}
	
	if chapterRef == nil {
		return fmt.Errorf("chapter not found: %s", chapterID)
	}
	
	data, err := yaml.Marshal(chapter)
	if err != nil {
		return fmt.Errorf("failed to marshal chapter: %w", err)
	}
	
	chapterPath := filepath.Join(s.config.GetDocumentFolder(docID), chapterRef.Folder, "chapter.yaml")
	if err := os.WriteFile(chapterPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write chapter manifest: %w", err)
	}
	
	return nil
}

// AddChapter adds a new chapter to a document
func (s *Storage) AddChapter(docID, title string, position document.Position) (string, error) {
	doc, err := s.GetDocument(docID)
	if err != nil {
		return "", err
	}
	
	if !doc.HasChapters {
		return "", fmt.Errorf("document does not support chapters")
	}
	
	// Generate chapter ID
	chapterID := fmt.Sprintf("ch-%03d", len(doc.Chapters)+1)
	chapterFolder := fmt.Sprintf("chapters/%s-%s", chapterID, s.sanitizeForPath(title))
	
	// Create chapter folder
	chapterPath := filepath.Join(s.config.GetDocumentFolder(docID), chapterFolder)
	if err := os.MkdirAll(chapterPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create chapter folder: %w", err)
	}
	
	// Create blocks folder for chapter
	blocksPath := filepath.Join(chapterPath, "blocks")
	if err := os.MkdirAll(blocksPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create chapter blocks folder: %w", err)
	}
	
	// Create chapter manifest
	chapter := &document.Chapter{
		ID:     chapterID,
		Title:  title,
		Blocks: []blocks.BlockReference{},
	}
	
	data, err := yaml.Marshal(chapter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal chapter: %w", err)
	}
	
	chapterManifest := filepath.Join(chapterPath, "chapter.yaml")
	if err := os.WriteFile(chapterManifest, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write chapter manifest: %w", err)
	}
	
	// Add to document manifest
	chapterRef := document.ChapterReference{
		ID:     chapterID,
		Title:  title,
		Folder: chapterFolder,
	}
	
	// Insert at position
	doc.Chapters = s.insertChapterAtPosition(doc.Chapters, chapterRef, position)
	
	if err := s.SaveDocument(docID, doc); err != nil {
		return "", err
	}
	
	return chapterID, nil
}

// sanitizeForPath cleans a string for use in a file path
func (s *Storage) sanitizeForPath(str string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)
	cleaned := reg.ReplaceAllString(strings.ToLower(str), "-")
	cleaned = strings.Trim(cleaned, "-")
	if len(cleaned) > 30 {
		cleaned = cleaned[:30]
	}
	return cleaned
}

// insertChapterAtPosition inserts a chapter at the specified position
func (s *Storage) insertChapterAtPosition(chapters []document.ChapterReference, newChapter document.ChapterReference, position document.Position) []document.ChapterReference {
	switch position.Type {
	case document.PositionStart:
		return append([]document.ChapterReference{newChapter}, chapters...)
	
	case document.PositionAfter:
		for i, chapter := range chapters {
			if chapter.ID == position.BlockID {
				result := make([]document.ChapterReference, 0, len(chapters)+1)
				result = append(result, chapters[:i+1]...)
				result = append(result, newChapter)
				result = append(result, chapters[i+1:]...)
				return result
			}
		}
		return append(chapters, newChapter)
	
	default: // PositionEnd
		return append(chapters, newChapter)
	}
}

// LoadBlock loads a block from file
func (s *Storage) LoadBlock(docID string, blockRef blocks.BlockReference) (blocks.Block, error) {
	filePath := filepath.Join(s.config.GetDocumentFolder(docID), blockRef.File)
	
	switch blockRef.Type {
	case blocks.TypeMarkdown:
		// Load markdown content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read markdown file: %w", err)
		}
		return &blocks.MarkdownBlock{
			BaseBlock: blocks.BaseBlock{ID: blockRef.ID, Type: blockRef.Type},
			Content:   string(content),
		}, nil
	
	case blocks.TypeHeading:
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read heading file: %w", err)
		}
		var heading blocks.HeadingBlock
		if err := yaml.Unmarshal(data, &heading); err != nil {
			return nil, fmt.Errorf("failed to parse heading: %w", err)
		}
		return &heading, nil
	
	case blocks.TypeImage:
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read image file: %w", err)
		}
		var image blocks.ImageBlock
		if err := yaml.Unmarshal(data, &image); err != nil {
			return nil, fmt.Errorf("failed to parse image: %w", err)
		}
		return &image, nil
	
	case blocks.TypeTable:
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read table file: %w", err)
		}
		var table blocks.TableBlock
		if err := yaml.Unmarshal(data, &table); err != nil {
			return nil, fmt.Errorf("failed to parse table: %w", err)
		}
		return &table, nil
	
	case blocks.TypePageBreak:
		return &blocks.PageBreakBlock{
			BaseBlock: blocks.BaseBlock{ID: blockRef.ID, Type: blockRef.Type},
		}, nil
	
	default:
		return nil, fmt.Errorf("unknown block type: %s", blockRef.Type)
	}
}

// CopyImageToAssets copies an image to the document's assets folder
func (s *Storage) CopyImageToAssets(docID, sourcePath string) (string, error) {
	// Open source file
	source, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source image: %w", err)
	}
	defer source.Close()
	
	// Generate asset filename
	ext := filepath.Ext(sourcePath)
	assetName := fmt.Sprintf("image_%d%s", time.Now().Unix(), ext)
	assetPath := filepath.Join(s.config.GetDocumentFolder(docID), "assets", assetName)
	
	// Create destination file
	dest, err := os.Create(assetPath)
	if err != nil {
		return "", fmt.Errorf("failed to create asset file: %w", err)
	}
	defer dest.Close()
	
	// Copy file
	if _, err := io.Copy(dest, source); err != nil {
		return "", fmt.Errorf("failed to copy image: %w", err)
	}
	
	// Return relative path
	return filepath.Join("assets", assetName), nil
}