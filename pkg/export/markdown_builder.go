package export

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/savant/mcp-servers/docgen2/pkg/blocks"
	"github.com/savant/mcp-servers/docgen2/pkg/storage"
)

// MarkdownBuilder converts document blocks to markdown
type MarkdownBuilder struct {
	storage *storage.Storage
}

// NewMarkdownBuilder creates a new markdown builder
func NewMarkdownBuilder(storage *storage.Storage) *MarkdownBuilder {
	return &MarkdownBuilder{storage: storage}
}

// BuildMarkdown converts a document to markdown string
func (mb *MarkdownBuilder) BuildMarkdown(docID string) (string, error) {
	doc, err := mb.storage.GetDocument(docID)
	if err != nil {
		return "", fmt.Errorf("failed to get document: %w", err)
	}

	var markdown strings.Builder

	// Add document title and metadata
	// Quote the title and author to handle special characters like colons
	markdown.WriteString(fmt.Sprintf("---\ntitle: \"%s\"\n", escapeYAMLString(doc.Title)))
	if doc.Author != "" {
		markdown.WriteString(fmt.Sprintf("author: \"%s\"\n", escapeYAMLString(doc.Author)))
	}
	markdown.WriteString("---\n\n")

	// Process document-level blocks first (if any)
	if len(doc.Blocks) > 0 {
		content, err := mb.processBlocks(docID, doc.Blocks)
		if err != nil {
			return "", fmt.Errorf("failed to process document blocks: %w", err)
		}
		markdown.WriteString(content)
	}

	// Then process chapters (if any)
	if len(doc.Chapters) > 0 {
		for _, chapterRef := range doc.Chapters {
			chapter, err := mb.storage.GetChapter(docID, chapterRef.ID)
			if err != nil {
				return "", fmt.Errorf("failed to get chapter %s: %w", chapterRef.ID, err)
			}

			// Add chapter title as H1
			markdown.WriteString(fmt.Sprintf("# %s\n\n", chapter.Title))

			// Process chapter blocks
			chapterContent, err := mb.processBlocks(docID, chapter.Blocks)
			if err != nil {
				return "", fmt.Errorf("failed to process chapter %s blocks: %w", chapterRef.ID, err)
			}
			markdown.WriteString(chapterContent)
		}
	}

	return markdown.String(), nil
}

// processBlocks converts a list of block references to markdown
func (mb *MarkdownBuilder) processBlocks(docID string, blockRefs []blocks.BlockReference) (string, error) {
	var result strings.Builder

	for _, blockRef := range blockRefs {
		block, err := mb.storage.LoadBlock(docID, blockRef)
		if err != nil {
			return "", fmt.Errorf("failed to load block %s: %w", blockRef.ID, err)
		}

		blockMarkdown, err := mb.blockToMarkdown(docID, block)
		if err != nil {
			return "", fmt.Errorf("failed to convert block %s to markdown: %w", blockRef.ID, err)
		}

		result.WriteString(blockMarkdown)
		result.WriteString("\n\n")
	}

	return result.String(), nil
}

// blockToMarkdown converts a single block to markdown
func (mb *MarkdownBuilder) blockToMarkdown(docID string, block blocks.Block) (string, error) {
	switch b := block.(type) {
	case *blocks.HeadingBlock:
		return mb.headingToMarkdown(b), nil

	case *blocks.MarkdownBlock:
		return b.Content, nil

	case *blocks.ImageBlock:
		return mb.imageToMarkdown(docID, b), nil

	case *blocks.TableBlock:
		return mb.tableToMarkdown(b), nil

	case *blocks.PageBreakBlock:
		// Use LaTeX command for page break (works in PDF and DOCX)
		return "\\newpage", nil

	default:
		return "", fmt.Errorf("unsupported block type: %T", block)
	}
}

// headingToMarkdown converts a heading block to markdown
func (mb *MarkdownBuilder) headingToMarkdown(heading *blocks.HeadingBlock) string {
	hashes := strings.Repeat("#", heading.Level)
	return fmt.Sprintf("%s %s", hashes, heading.Text)
}

// imageToMarkdown converts an image block to markdown
func (mb *MarkdownBuilder) imageToMarkdown(docID string, img *blocks.ImageBlock) string {
	// Build absolute path to the image file
	// The image path is relative to the document folder
	config := mb.storage.GetConfig()
	docFolder := config.GetDocumentFolder(docID)
	imagePath := filepath.Join(docFolder, img.Path)
	
	// For PDF generation with xelatex, we need to quote paths with spaces
	// Markdown image syntax allows quotes around the path
	if img.Caption != "" {
		return fmt.Sprintf(`![%s]("%s")`, img.Caption, imagePath)
	}
	return fmt.Sprintf(`![]("%s")`, imagePath)
}

// tableToMarkdown converts a table block to markdown
func (mb *MarkdownBuilder) tableToMarkdown(table *blocks.TableBlock) string {
	if len(table.Headers) == 0 {
		return ""
	}

	var result strings.Builder

	// Write headers
	result.WriteString("|")
	for _, header := range table.Headers {
		result.WriteString(fmt.Sprintf(" %s |", header))
	}
	result.WriteString("\n")

	// Write separator
	result.WriteString("|")
	for range table.Headers {
		result.WriteString(" --- |")
	}
	result.WriteString("\n")

	// Write rows
	for _, row := range table.Rows {
		result.WriteString("|")
		for i, cell := range row {
			if i < len(table.Headers) {
				result.WriteString(fmt.Sprintf(" %s |", cell))
			}
		}
		// Fill empty cells if row is shorter than headers
		for i := len(row); i < len(table.Headers); i++ {
			result.WriteString(" |")
		}
		result.WriteString("\n")
	}

	return result.String()
}

// escapeYAMLString escapes special characters in YAML strings
func escapeYAMLString(s string) string {
	// Escape backslashes and double quotes for YAML quoted strings
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// BuildChapterMarkdown builds markdown for a specific chapter
func (mb *MarkdownBuilder) BuildChapterMarkdown(docID, chapterID string) (string, error) {
	chapter, err := mb.storage.GetChapter(docID, chapterID)
	if err != nil {
		return "", fmt.Errorf("failed to get chapter: %w", err)
	}

	var markdown strings.Builder
	markdown.WriteString(fmt.Sprintf("# %s\n\n", chapter.Title))

	content, err := mb.processBlocks(docID, chapter.Blocks)
	if err != nil {
		return "", fmt.Errorf("failed to process chapter blocks: %w", err)
	}

	markdown.WriteString(content)
	return markdown.String(), nil
}