package document

import (
	"time"
	
	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/style"
)

// Document represents a document with its metadata
type Document struct {
	Title       string    `yaml:"title"`
	Author      string    `yaml:"author,omitempty"`
	CreatedAt   time.Time `yaml:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at"`
	HasChapters bool      `yaml:"has_chapters"`
	
	// Styling configuration
	Style *style.StyleConfig `yaml:"style,omitempty"`
	
	// For flat documents
	Blocks []blocks.BlockReference `yaml:"blocks,omitempty"`
	
	// For chaptered documents
	Chapters []ChapterReference `yaml:"chapters,omitempty"`
}

// Chapter represents a chapter in a document
type Chapter struct {
	ID     string                  `yaml:"id"`
	Title  string                  `yaml:"title"`
	Blocks []blocks.BlockReference `yaml:"blocks"`
}

// ChapterReference stores the reference to a chapter in the manifest
type ChapterReference struct {
	ID     string `yaml:"id"`
	Title  string `yaml:"title"`
	Folder string `yaml:"folder"`
}

// DocumentOverview provides a tree structure of the document
type DocumentOverview struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Author      string             `json:"author,omitempty"`
	HasChapters bool               `json:"has_chapters"`
	Blocks      []BlockOverview    `json:"blocks"`
	Chapters    []ChapterOverview  `json:"chapters"`
}

// ChapterOverview provides overview of a chapter
type ChapterOverview struct {
	ID     string          `json:"id"`
	Title  string          `json:"title"`
	Blocks []BlockOverview `json:"blocks"`
}

// BlockOverview provides overview of a block
type BlockOverview struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Preview string `json:"preview"` // First 100 chars or summary
}

// Position represents where to add a new block/chapter
type Position struct {
	Type    PositionType `json:"type"`
	BlockID string       `json:"block_id,omitempty"` // For "after" type
}

// PositionType represents the type of position
type PositionType string

const (
	PositionStart PositionType = "start"
	PositionEnd   PositionType = "end"
	PositionAfter PositionType = "after"
)

// ParsePosition parses a position string
func ParsePosition(pos string) Position {
	if pos == "" || pos == "end" {
		return Position{Type: PositionEnd}
	}
	if pos == "start" {
		return Position{Type: PositionStart}
	}
	if len(pos) > 6 && pos[:6] == "after:" {
		return Position{
			Type:    PositionAfter,
			BlockID: pos[6:],
		}
	}
	return Position{Type: PositionEnd}
}

// SearchResult represents a search result
type SearchResult struct {
	BlockID   string `json:"block_id"`
	BlockType string `json:"block_type"`
	ChapterID string `json:"chapter_id,omitempty"`
	Snippet   string `json:"snippet"`
	Position  int    `json:"position"` // Position in document/chapter
}