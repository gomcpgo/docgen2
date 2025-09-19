# DocGen2 MCP Server

A Model Context Protocol (MCP) server for creating and managing structured documents. DocGen2 enables LLMs to create books, reports, resumes, and other documents with a flexible block-based architecture.

## Features

- **Block-Based Documents**: Documents are composed of discrete content blocks (headings, markdown, images, tables, page breaks)
- **Chapter Support**: Create books and large documents with chapter organization
- **Document Search**: Search within documents to find and update specific content
- **Multiple Export Formats**: Export to PDF, DOCX, and HTML (Pandoc integration planned)
- **Terminal Mode**: CLI interface for testing and direct document manipulation
- **File-Based Storage**: All documents stored as YAML and markdown files for easy version control

## Installation

### Prerequisites

- Go 1.23 or higher
- Pandoc (for export functionality - coming soon)

### Building from Source

```bash
# Clone the repository
cd mcp_servers/docgen2

# Build the server
./run.sh build

# Or directly with go
go build -o bin/docgen2 ./cmd
```

## Usage

### MCP Server Mode

Run as an MCP server for integration with Claude or other LLM clients:

```bash
./bin/docgen2
# or
./run.sh run
```

### Terminal Mode

DocGen2 includes a comprehensive terminal mode for testing and direct manipulation:

```bash
# Create a new document
./bin/docgen2 -create "My Document Title"
./bin/docgen2 -create "My Book" -chapters  # With chapter support

# List all documents
./bin/docgen2 -list

# Show document structure
./bin/docgen2 -overview <document-id>

# Add content
./bin/docgen2 -add-heading <doc-id> -level 1 -text "Introduction"
./bin/docgen2 -add-markdown <doc-id> -content "Your markdown content here"

# Export document (coming soon)
./bin/docgen2 -export <doc-id> -format pdf
```

## Document Structure

### Storage Layout

Documents are stored in a hierarchical file structure:

```
docgen_data/
└── documents/
    └── my-document/
        ├── manifest.yaml          # Document metadata
        ├── blocks/                # Content blocks
        │   ├── hd-001-heading.yaml
        │   ├── md-001.md
        │   └── img-001-image.yaml
        └── assets/                # Images and attachments
            └── image.png
```

### Block Types

1. **Heading**: Section titles with levels h1-h6
2. **Markdown**: General formatted text, lists, code blocks, quotes
3. **Image**: Images with optional captions and alt text
4. **Table**: Structured data in CSV-like format
5. **Page Break**: Force page breaks in PDF/DOCX output

## MCP Tools

The server provides the following MCP tools:

### Document Operations
- `create_document` - Create a new document
- `list_documents` - List all documents
- `get_document_overview` - Get document structure
- `delete_document` - Delete a document
- `search_blocks` - Search within documents

### Block Operations
- `add_heading` - Add a heading block
- `add_markdown` - Add markdown content
- `add_image` - Add an image with metadata
- `add_table` - Add a structured table
- `add_page_break` - Add a page break
- `add_multiple_blocks` - Add multiple blocks at once
- `get_block` - Get specific block content
- `update_block` - Update existing block (planned)
- `delete_block` - Delete a block (planned)
- `move_block` - Reorder blocks (planned)

### Chapter Operations
- `add_chapter` - Add a chapter to chaptered documents
- `update_chapter` - Update chapter title (planned)
- `delete_chapter` - Delete a chapter (planned)
- `move_chapter` - Reorder chapters (planned)

### Export Operations
- `export_document` - Export to PDF/DOCX/HTML (planned)

## Configuration

Set the document storage location using environment variable:

```bash
export DOCGEN_ROOT=/path/to/documents
```

Default location is `./docgen_data` in the current directory.

## Development

### Running Tests

```bash
# Unit tests
./run.sh test

# Integration tests
./run.sh integration

# All tests
go test ./... -v
```

### Project Structure

```
docgen2/
├── cmd/           # Main entry point
├── pkg/
│   ├── blocks/    # Block types and interfaces
│   ├── config/    # Configuration management
│   ├── document/  # Document types and operations
│   ├── handler/   # MCP protocol handlers
│   └── storage/   # File system operations
└── test/          # Integration tests
```

## Design Principles

1. **Block-Based Architecture**: Enables granular editing without regenerating entire documents
2. **File Storage**: No database required, easy backup and version control
3. **Minimal Complexity**: Focus on current requirements, no over-engineering
4. **Fail Fast**: Clear error messages, no silent failures

See [docs/docgen2-design-document.md](../../docs/docgen2-design-document.md) for detailed design documentation.

## License
MIT License
