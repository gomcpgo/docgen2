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
		createDoc    string
		withChapters bool
		listDocs     bool
		overview     string
		addHeading   string
		headingLevel int
		headingText  string
		addMarkdown  string
		markdownContent string
		exportDoc    string
		exportFormat string
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