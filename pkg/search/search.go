package search

import (
	"strings"
	
	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/document"
	"github.com/savant/mcp-servers/docgen2/pkg/storage"
)

// Searcher handles document search operations
type Searcher struct {
	storage *storage.Storage
}

// NewSearcher creates a new searcher instance
func NewSearcher(storage *storage.Storage) *Searcher {
	return &Searcher{storage: storage}
}

// SearchDocument searches for a query within a document
func (s *Searcher) SearchDocument(docID, query, chapterID string) ([]document.SearchResult, error) {
	doc, err := s.storage.GetDocument(docID)
	if err != nil {
		return nil, err
	}
	
	var results []document.SearchResult
	queryLower := strings.ToLower(query)
	
	// Search in document-level blocks first (if any)
	if len(doc.Blocks) > 0 {
		results = append(results, s.SearchInBlocks(docID, doc.Blocks, queryLower, "")...)
	}

	// Then search in chapters (if any and if no specific chapter requested, or if the specific chapter is being searched)
	if len(doc.Chapters) > 0 {
		for _, chapterRef := range doc.Chapters {
			// Skip if specific chapter requested and this isn't it
			if chapterID != "" && chapterRef.ID != chapterID {
				continue
			}
			
			chapter, err := s.storage.GetChapter(docID, chapterRef.ID)
			if err != nil {
				continue
			}
			
			results = append(results, s.SearchInBlocks(docID, chapter.Blocks, queryLower, chapterRef.ID)...)
		}
	}
	
	return results, nil
}

// SearchInBlocks searches for query in a list of blocks
func (s *Searcher) SearchInBlocks(docID string, blockRefs []blocks.BlockReference, query string, chapterID string) []document.SearchResult {
	var results []document.SearchResult
	
	for i, ref := range blockRefs {
		block, err := s.storage.LoadBlock(docID, ref)
		if err != nil {
			continue
		}
		
		content := s.GetBlockContent(block)
		if strings.Contains(strings.ToLower(content), query) {
			// Find snippet around the match
			snippet := s.ExtractSnippet(content, query)
			
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

// GetBlockContent extracts searchable text from a block
func (s *Searcher) GetBlockContent(block blocks.Block) string {
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

// ExtractSnippet extracts a snippet around the query match
func (s *Searcher) ExtractSnippet(content, query string) string {
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

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}