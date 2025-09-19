package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	
	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/mcp/pkg/server"
	"github.com/savant/mcp-servers/docgen2/pkg/config"
	docgenHandler "github.com/savant/mcp-servers/docgen2/pkg/handler"
)

func main() {
	// Terminal mode flags
	var (
		createDoc       string
		withChapters    bool
		listDocs        bool
		overview        string
		addHeading      string
		headingLevel    int
		headingText     string
		addMarkdown     string
		markdownContent string
		exportDoc       string
		exportFormat    string
		
		// New block operations
		updateBlock     string
		deleteBlock     string
		moveBlock       string
		getBlock        string
		blockID         string
		newPosition     string
		newContent      string
	)
	
	flag.StringVar(&createDoc, "create", "", "Create a new document with the given title")
	flag.BoolVar(&withChapters, "chapters", false, "Create document with chapters support")
	flag.BoolVar(&listDocs, "list", false, "List all documents")
	flag.StringVar(&overview, "overview", "", "Show overview of a document")
	flag.StringVar(&addHeading, "add-heading", "", "Add heading to document (specify doc ID)")
	flag.IntVar(&headingLevel, "level", 1, "Heading level (1-6)")
	flag.StringVar(&headingText, "text", "", "Heading text")
	flag.StringVar(&addMarkdown, "add-markdown", "", "Add markdown to document (specify doc ID)")
	flag.StringVar(&markdownContent, "content", "", "Markdown content")
	flag.StringVar(&exportDoc, "export", "", "Export document (specify doc ID)")
	flag.StringVar(&exportFormat, "format", "html", "Export format (pdf, docx, html)")
	
	// New block operation flags
	flag.StringVar(&updateBlock, "update-block", "", "Update block in document (specify doc ID)")
	flag.StringVar(&deleteBlock, "delete-block", "", "Delete block from document (specify doc ID)")
	flag.StringVar(&moveBlock, "move-block", "", "Move block in document (specify doc ID)")
	flag.StringVar(&getBlock, "get-block", "", "Get block from document (specify doc ID)")
	flag.StringVar(&blockID, "block-id", "", "Block ID for block operations")
	flag.StringVar(&newPosition, "new-position", "", "New position for move operation (start, end, after:block-id)")
	flag.StringVar(&newContent, "new-content", "", "New content for update operation (JSON format)")
	
	flag.Parse()
	
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Create handler
	h := docgenHandler.NewHandler(cfg)
	ctx := context.Background()
	
	// Terminal mode operations
	if createDoc != "" {
		runTerminalCommand(ctx, h, "create_document", map[string]interface{}{
			"title":        createDoc,
			"has_chapters": withChapters,
		})
		return
	}
	
	if listDocs {
		runTerminalCommand(ctx, h, "list_documents", map[string]interface{}{})
		return
	}
	
	if overview != "" {
		runTerminalCommand(ctx, h, "get_document_overview", map[string]interface{}{
			"document_id": overview,
		})
		return
	}
	
	if addHeading != "" {
		if headingText == "" {
			log.Fatal("Heading text is required (use -text flag)")
		}
		runTerminalCommand(ctx, h, "add_heading", map[string]interface{}{
			"document_id": addHeading,
			"level":       headingLevel,
			"text":        headingText,
		})
		return
	}
	
	if addMarkdown != "" {
		if markdownContent == "" {
			log.Fatal("Markdown content is required (use -content flag)")
		}
		runTerminalCommand(ctx, h, "add_markdown", map[string]interface{}{
			"document_id": addMarkdown,
			"content":     markdownContent,
		})
		return
	}
	
	if exportDoc != "" {
		runTerminalCommand(ctx, h, "export_document", map[string]interface{}{
			"document_id": exportDoc,
			"format":      exportFormat,
		})
		return
	}
	
	if updateBlock != "" {
		if blockID == "" {
			log.Fatal("Block ID is required for update operation (use -block-id flag)")
		}
		if newContent == "" {
			log.Fatal("New content is required for update operation (use -new-content flag)")
		}
		
		// Parse JSON content
		var contentData map[string]interface{}
		if err := json.Unmarshal([]byte(newContent), &contentData); err != nil {
			log.Fatalf("Invalid JSON in new-content: %v", err)
		}
		
		runTerminalCommand(ctx, h, "update_block", map[string]interface{}{
			"document_id": updateBlock,
			"block_id":    blockID,
			"new_content": contentData,
		})
		return
	}
	
	if deleteBlock != "" {
		if blockID == "" {
			log.Fatal("Block ID is required for delete operation (use -block-id flag)")
		}
		runTerminalCommand(ctx, h, "delete_block", map[string]interface{}{
			"document_id": deleteBlock,
			"block_id":    blockID,
		})
		return
	}
	
	if moveBlock != "" {
		if blockID == "" {
			log.Fatal("Block ID is required for move operation (use -block-id flag)")
		}
		if newPosition == "" {
			log.Fatal("New position is required for move operation (use -new-position flag)")
		}
		runTerminalCommand(ctx, h, "move_block", map[string]interface{}{
			"document_id":  moveBlock,
			"block_id":     blockID,
			"new_position": newPosition,
		})
		return
	}
	
	if getBlock != "" {
		if blockID == "" {
			log.Fatal("Block ID is required for get operation (use -block-id flag)")
		}
		runTerminalCommand(ctx, h, "get_block", map[string]interface{}{
			"document_id": getBlock,
			"block_id":    blockID,
		})
		return
	}
	
	// MCP Server mode
	fmt.Println("Starting DocGen2 MCP Server...")
	fmt.Printf("Root folder: %s\n", cfg.RootFolder)
	
	// Create handler registry
	registry := handler.NewHandlerRegistry()
	registry.RegisterToolHandler(h)
	
	// Create and run server
	srv := server.New(server.Options{
		Name:     "docgen2",
		Version:  "1.0.0",
		Registry: registry,
	})
	
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// runTerminalCommand executes a tool command in terminal mode
func runTerminalCommand(ctx context.Context, h *docgenHandler.Handler, toolName string, args map[string]interface{}) {
	req := &protocol.CallToolRequest{
		Name:      toolName,
		Arguments: args,
	}
	
	resp, err := h.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	// Print response
	for _, content := range resp.Content {
		if content.Type == "text" {
			// Try to pretty print JSON
			var jsonData interface{}
			if err := json.Unmarshal([]byte(content.Text), &jsonData); err == nil {
				prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")
				fmt.Println(string(prettyJSON))
			} else {
				fmt.Println(content.Text)
			}
		}
	}
}