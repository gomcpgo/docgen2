package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
	"gopkg.in/yaml.v3"
)

// GetChapter retrieves a chapter from storage
func (s *Storage) GetChapter(docID, chapterID string) (*document.Chapter, error) {
	doc, err := s.GetDocument(docID)
	if err != nil {
		return nil, err
	}
	
	if !doc.HasChapters {
		return nil, fmt.Errorf("document %s does not have chapters", docID)
	}
	
	// Find chapter reference
	var chapterRef *document.ChapterReference
	for _, ref := range doc.Chapters {
		if ref.ID == chapterID {
			chapterRef = &ref
			break
		}
	}
	
	if chapterRef == nil {
		return nil, fmt.Errorf("chapter %s not found in document %s", chapterID, docID)
	}
	
	// Load chapter file
	chapterPath := filepath.Join(s.config.GetDocumentFolder(docID), chapterRef.Folder, "chapter.yaml")
	data, err := os.ReadFile(chapterPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chapter file: %w", err)
	}
	
	var chapter document.Chapter
	if err := yaml.Unmarshal(data, &chapter); err != nil {
		return nil, fmt.Errorf("failed to parse chapter file: %w", err)
	}
	
	return &chapter, nil
}

// SaveChapter saves a chapter to storage
func (s *Storage) SaveChapter(docID, chapterID string, chapter *document.Chapter) error {
	doc, err := s.GetDocument(docID)
	if err != nil {
		return err
	}
	
	if !doc.HasChapters {
		return fmt.Errorf("document %s does not have chapters", docID)
	}
	
	// Find chapter reference
	var chapterRef *document.ChapterReference
	for _, ref := range doc.Chapters {
		if ref.ID == chapterID {
			chapterRef = &ref
			break
		}
	}
	
	if chapterRef == nil {
		return fmt.Errorf("chapter %s not found in document %s", chapterID, docID)
	}
	
	// Save chapter file
	chapterPath := filepath.Join(s.config.GetDocumentFolder(docID), chapterRef.Folder, "chapter.yaml")
	data, err := yaml.Marshal(chapter)
	if err != nil {
		return fmt.Errorf("failed to marshal chapter: %w", err)
	}
	
	if err := os.WriteFile(chapterPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write chapter file: %w", err)
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
	chapterNum := len(doc.Chapters) + 1
	chapterID := fmt.Sprintf("ch-%03d", chapterNum)
	
	// Ensure unique ID
	for {
		exists := false
		for _, ch := range doc.Chapters {
			if ch.ID == chapterID {
				exists = true
				break
			}
		}
		if !exists {
			break
		}
		chapterNum++
		chapterID = fmt.Sprintf("ch-%03d", chapterNum)
	}
	
	// Create chapter folder name
	sanitizedTitle := s.sanitizeForPath(title)
	chapterFolder := fmt.Sprintf("chapters/%s-%s", chapterID, sanitizedTitle)
	
	// Create chapter folder structure
	chapterPath := filepath.Join(s.config.GetDocumentFolder(docID), chapterFolder)
	if err := os.MkdirAll(chapterPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create chapter folder: %w", err)
	}
	
	// Create blocks folder for the chapter
	blocksPath := filepath.Join(chapterPath, "blocks")
	if err := os.MkdirAll(blocksPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create chapter blocks folder: %w", err)
	}
	
	// Create chapter reference
	chapterRef := document.ChapterReference{
		ID:     chapterID,
		Title:  title,
		Folder: chapterFolder,
	}
	
	// Insert chapter at position
	doc.Chapters = s.insertChapterAtPosition(doc.Chapters, chapterRef, position)
	
	// Save document manifest
	if err := s.SaveDocument(docID, doc); err != nil {
		// Clean up on error
		os.RemoveAll(chapterPath)
		return "", err
	}
	
	// Create chapter file
	chapter := &document.Chapter{
		ID:     chapterID,
		Title:  title,
		Blocks: []blocks.BlockReference{},
	}
	
	if err := s.SaveChapter(docID, chapterID, chapter); err != nil {
		// Clean up on error
		os.RemoveAll(chapterPath)
		return "", err
	}
	
	return chapterID, nil
}

// sanitizeForPath converts a string to be safe for use in filesystem paths
func (s *Storage) sanitizeForPath(str string) string {
	// Convert to lowercase and replace spaces with hyphens
	str = strings.ToLower(str)
	str = strings.ReplaceAll(str, " ", "-")
	
	// Remove any characters that aren't alphanumeric or hyphens
	var result strings.Builder
	for _, r := range str {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	
	// Truncate if too long
	sanitized := result.String()
	if len(sanitized) > 30 {
		sanitized = sanitized[:30]
	}
	
	// Remove any leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")
	
	return sanitized
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
		// If chapter not found, append to end
		return append(chapters, newChapter)
	
	case document.PositionEnd:
		fallthrough
	default:
		return append(chapters, newChapter)
	}
}

// UpdateChapter updates a chapter's title
func (s *Storage) UpdateChapter(docID, chapterID, newTitle string) error {
	doc, err := s.GetDocument(docID)
	if err != nil {
		return err
	}
	
	if !doc.HasChapters {
		return fmt.Errorf("document does not have chapters")
	}
	
	// Find and update chapter in document manifest
	found := false
	for i, chapterRef := range doc.Chapters {
		if chapterRef.ID == chapterID {
			doc.Chapters[i].Title = newTitle
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("chapter not found: %s", chapterID)
	}
	
	// Update chapter file
	chapter, err := s.GetChapter(docID, chapterID)
	if err != nil {
		return err
	}
	
	chapter.Title = newTitle
	if err := s.SaveChapter(docID, chapterID, chapter); err != nil {
		return err
	}
	
	// Save document manifest
	return s.SaveDocument(docID, doc)
}

// DeleteChapter deletes a chapter and all its contents
func (s *Storage) DeleteChapter(docID, chapterID string) error {
	doc, err := s.GetDocument(docID)
	if err != nil {
		return err
	}
	
	if !doc.HasChapters {
		return fmt.Errorf("document does not have chapters")
	}
	
	// Find chapter index
	chapterIndex := -1
	for i, chapterRef := range doc.Chapters {
		if chapterRef.ID == chapterID {
			chapterIndex = i
			break
		}
	}
	
	if chapterIndex == -1 {
		return fmt.Errorf("chapter not found: %s", chapterID)
	}
	
	// Remove chapter from manifest
	doc.Chapters = append(doc.Chapters[:chapterIndex], doc.Chapters[chapterIndex+1:]...)
	
	// Delete chapter directory and all its contents
	chapterDir := filepath.Join(s.config.GetDocumentFolder(docID), "chapters", chapterID)
	if err := os.RemoveAll(chapterDir); err != nil {
		return fmt.Errorf("failed to delete chapter directory: %w", err)
	}
	
	// Save updated document manifest
	return s.SaveDocument(docID, doc)
}

// MoveChapter moves a chapter to a new position in the document
func (s *Storage) MoveChapter(docID, chapterID string, newPosition document.Position) error {
	doc, err := s.GetDocument(docID)
	if err != nil {
		return err
	}
	
	if !doc.HasChapters {
		return fmt.Errorf("document does not have chapters")
	}
	
	// Find current chapter
	var targetChapter document.ChapterReference
	currentIndex := -1
	for i, chapterRef := range doc.Chapters {
		if chapterRef.ID == chapterID {
			targetChapter = chapterRef
			currentIndex = i
			break
		}
	}
	
	if currentIndex == -1 {
		return fmt.Errorf("chapter not found: %s", chapterID)
	}
	
	// Remove chapter from current position
	remainingChapters := append(doc.Chapters[:currentIndex], doc.Chapters[currentIndex+1:]...)
	
	// Insert at new position
	doc.Chapters = s.insertChapterAtPosition(remainingChapters, targetChapter, newPosition)
	
	// Save document manifest
	return s.SaveDocument(docID, doc)
}