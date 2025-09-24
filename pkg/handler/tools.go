package handler

import (
	"encoding/json"
	
	"github.com/gomcpgo/mcp/pkg/protocol"
)

func getAllTools() []protocol.Tool {
	return []protocol.Tool{
		// Document operations
		{
			Name:        "create_document",
			Description: "Create a new document with optional chapters support",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"title": {
						"type": "string",
						"description": "The document title"
					},
					"has_chapters": {
						"type": "boolean",
						"description": "Whether the document should support chapters (default: false)"
					},
					"author": {
						"type": "string",
						"description": "The document author (optional)"
					}
				},
				"required": ["title"]
			}`),
		},
		{
			Name:        "list_documents",
			Description: "List all available documents",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {}
			}`),
		},
		{
			Name:        "get_document_overview",
			Description: "Get a hierarchical overview of a document's structure",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					}
				},
				"required": ["document_id"]
			}`),
		},
		{
			Name:        "delete_document",
			Description: "Delete a document and all its content",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID to delete"
					}
				},
				"required": ["document_id"]
			}`),
		},
		{
			Name:        "search_blocks",
			Description: "Search for blocks containing specific text within a document",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"query": {
						"type": "string",
						"description": "The search query"
					},
					"chapter_id": {
						"type": "string",
						"description": "Optional: limit search to a specific chapter"
					}
				},
				"required": ["document_id", "query"]
			}`),
		},
		
		// Block operations
		{
			Name:        "add_heading",
			Description: "Add a heading block to a document",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"chapter_id": {
						"type": "string",
						"description": "Optional: chapter ID for chaptered documents"
					},
					"level": {
						"type": "integer",
						"description": "Heading level (1-6 for h1-h6)",
						"minimum": 1,
						"maximum": 6
					},
					"text": {
						"type": "string",
						"description": "The heading text"
					},
					"position": {
						"type": "string",
						"description": "Where to add the block: 'start', 'end', or 'after:block-id' (default: 'end')"
					}
				},
				"required": ["document_id", "level", "text"]
			}`),
		},
		{
			Name:        "add_markdown",
			Description: "Add a markdown content block to a document",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"chapter_id": {
						"type": "string",
						"description": "Optional: chapter ID for chaptered documents"
					},
					"content": {
						"type": "string",
						"description": "The markdown content"
					},
					"position": {
						"type": "string",
						"description": "Where to add the block: 'start', 'end', or 'after:block-id' (default: 'end')"
					}
				},
				"required": ["document_id", "content"]
			}`),
		},
		{
			Name:        "add_image",
			Description: "Add an image block to a document",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"chapter_id": {
						"type": "string",
						"description": "Optional: chapter ID for chaptered documents"
					},
					"image_path": {
						"type": "string",
						"description": "Path to the image file (will be copied to document assets)"
					},
					"caption": {
						"type": "string",
						"description": "Optional image caption"
					},
					"alt_text": {
						"type": "string",
						"description": "Optional alt text for accessibility"
					},
					"position": {
						"type": "string",
						"description": "Where to add the block: 'start', 'end', or 'after:block-id' (default: 'end')"
					}
				},
				"required": ["document_id", "image_path"]
			}`),
		},
		{
			Name:        "add_table",
			Description: "Add a table block to a document",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"chapter_id": {
						"type": "string",
						"description": "Optional: chapter ID for chaptered documents"
					},
					"headers": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Table column headers"
					},
					"rows": {
						"type": "array",
						"items": {
							"type": "array",
							"items": {"type": "string"}
						},
						"description": "Table rows (2D array of strings)"
					},
					"position": {
						"type": "string",
						"description": "Where to add the block: 'start', 'end', or 'after:block-id' (default: 'end')"
					}
				},
				"required": ["document_id", "headers", "rows"]
			}`),
		},
		{
			Name:        "add_page_break",
			Description: "Add a page break block to a document (only affects PDF/DOCX export)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"chapter_id": {
						"type": "string",
						"description": "Optional: chapter ID for chaptered documents"
					},
					"position": {
						"type": "string",
						"description": "Where to add the block: 'start', 'end', or 'after:block-id' (default: 'end')"
					}
				},
				"required": ["document_id"]
			}`),
		},
		{
			Name:        "add_multiple_blocks",
			Description: "Add multiple blocks to a document in one operation",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"chapter_id": {
						"type": "string",
						"description": "Optional: chapter ID for chaptered documents"
					},
					"blocks": {
						"type": "array",
						"description": "Array of blocks to add",
						"items": {
							"type": "object",
							"properties": {
								"type": {
									"type": "string",
									"enum": ["heading", "markdown", "image", "table", "page_break"],
									"description": "Block type"
								},
								"data": {
									"type": "object",
									"description": "Block data (varies by type)"
								}
							},
							"required": ["type", "data"]
						}
					},
					"position": {
						"type": "string",
						"description": "Where to add the blocks: 'start', 'end', or 'after:block-id' (default: 'end')"
					}
				},
				"required": ["document_id", "blocks"]
			}`),
		},
		{
			Name:        "update_block",
			Description: "Update the content of an existing block",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"block_id": {
						"type": "string",
						"description": "The block ID to update"
					},
					"new_content": {
						"type": "object",
						"description": "New content for the block (structure depends on block type)"
					}
				},
				"required": ["document_id", "block_id", "new_content"]
			}`),
		},
		{
			Name:        "delete_block",
			Description: "Delete a block from a document",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"block_id": {
						"type": "string",
						"description": "The block ID to delete"
					}
				},
				"required": ["document_id", "block_id"]
			}`),
		},
		{
			Name:        "move_block",
			Description: "Move a block to a new position within the document",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"block_id": {
						"type": "string",
						"description": "The block ID to move"
					},
					"new_position": {
						"type": "string",
						"description": "New position: 'start', 'end', or 'after:block-id'"
					}
				},
				"required": ["document_id", "block_id", "new_position"]
			}`),
		},
		{
			Name:        "get_block",
			Description: "Get the content of a specific block",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"block_id": {
						"type": "string",
						"description": "The block ID"
					}
				},
				"required": ["document_id", "block_id"]
			}`),
		},
		{
			Name:        "get_blocks",
			Description: "Get the content of multiple blocks in a single request",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"block_ids": {
						"type": "array",
						"items": {
							"type": "string"
						},
						"description": "Array of block IDs to fetch"
					}
				},
				"required": ["document_id", "block_ids"]
			}`),
		},
		
		// Chapter operations
		{
			Name:        "add_chapter",
			Description: "Add a new chapter to a chaptered document",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"title": {
						"type": "string",
						"description": "The chapter title"
					},
					"position": {
						"type": "string",
						"description": "Where to add the chapter: 'start', 'end', or 'after:chapter-id' (default: 'end')"
					}
				},
				"required": ["document_id", "title"]
			}`),
		},
		{
			Name:        "update_chapter",
			Description: "Update a chapter's title",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"chapter_id": {
						"type": "string",
						"description": "The chapter ID"
					},
					"title": {
						"type": "string",
						"description": "New chapter title"
					}
				},
				"required": ["document_id", "chapter_id", "title"]
			}`),
		},
		{
			Name:        "delete_chapter",
			Description: "Delete a chapter and all its blocks",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"chapter_id": {
						"type": "string",
						"description": "The chapter ID to delete"
					}
				},
				"required": ["document_id", "chapter_id"]
			}`),
		},
		{
			Name:        "move_chapter",
			Description: "Move a chapter to a new position",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"chapter_id": {
						"type": "string",
						"description": "The chapter ID to move"
					},
					"new_position": {
						"type": "string",
						"description": "New position: 'start', 'end', or 'after:chapter-id'"
					}
				},
				"required": ["document_id", "chapter_id", "new_position"]
			}`),
		},
		
		// Export operations
		{
			Name:        "export_document",
			Description: "Export a document to PDF, DOCX, or HTML format",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "The document ID"
					},
					"format": {
						"type": "string",
						"enum": ["pdf", "docx", "html"],
						"description": "Export format"
					}
				},
				"required": ["document_id", "format"]
			}`),
		},
	}
}