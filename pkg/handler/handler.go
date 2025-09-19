package handler

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/savant/mcp-servers/docgen2/pkg/config"
	"github.com/savant/mcp-servers/docgen2/pkg/search"
	"github.com/savant/mcp-servers/docgen2/pkg/storage"
)

// Handler implements the MCP tool handlers for DocGen2
type Handler struct {
	storage  *storage.Storage
	searcher *search.Searcher
	config   *config.Config
}

// NewHandler creates a new Handler instance
func NewHandler(cfg *config.Config) *Handler {
	stor := storage.NewStorage(cfg)
	return &Handler{
		storage:  stor,
		searcher: search.NewSearcher(stor),
		config:   cfg,
	}
}

// ListTools returns the list of available tools
func (h *Handler) ListTools(ctx context.Context) (*protocol.ListToolsResponse, error) {
	return &protocol.ListToolsResponse{
		Tools: getAllTools(),
	}, nil
}

// CallTool handles tool invocations
func (h *Handler) CallTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResponse, error) {
	switch req.Name {
	// Document operations
	case "create_document":
		return h.handleCreateDocument(ctx, req.Arguments)
	case "list_documents":
		return h.handleListDocuments(ctx, req.Arguments)
	case "get_document_overview":
		return h.handleGetDocumentOverview(ctx, req.Arguments)
	case "delete_document":
		return h.handleDeleteDocument(ctx, req.Arguments)
	case "search_blocks":
		return h.handleSearchBlocks(ctx, req.Arguments)
		
	// Block operations
	case "add_heading":
		return h.handleAddHeading(ctx, req.Arguments)
	case "add_markdown":
		return h.handleAddMarkdown(ctx, req.Arguments)
	case "add_image":
		return h.handleAddImage(ctx, req.Arguments)
	case "add_table":
		return h.handleAddTable(ctx, req.Arguments)
	case "add_page_break":
		return h.handleAddPageBreak(ctx, req.Arguments)
	case "add_multiple_blocks":
		return h.handleAddMultipleBlocks(ctx, req.Arguments)
	case "update_block":
		return h.handleUpdateBlock(ctx, req.Arguments)
	case "delete_block":
		return h.handleDeleteBlock(ctx, req.Arguments)
	case "move_block":
		return h.handleMoveBlock(ctx, req.Arguments)
	case "get_block":
		return h.handleGetBlock(ctx, req.Arguments)
		
	// Chapter operations
	case "add_chapter":
		return h.handleAddChapter(ctx, req.Arguments)
	case "update_chapter":
		return h.handleUpdateChapter(ctx, req.Arguments)
	case "delete_chapter":
		return h.handleDeleteChapter(ctx, req.Arguments)
	case "move_chapter":
		return h.handleMoveChapter(ctx, req.Arguments)
		
	// Export operations
	case "export_document":
		return h.handleExportDocument(ctx, req.Arguments)
		
	default:
		return nil, fmt.Errorf("unknown tool: %s", req.Name)
	}
}

// Helper function to extract string from arguments
func getString(args map[string]interface{}, key string, required bool) (string, error) {
	val, ok := args[key]
	if !ok {
		if required {
			return "", fmt.Errorf("%s parameter is required", key)
		}
		return "", nil
	}
	
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", key)
	}
	
	if required && str == "" {
		return "", fmt.Errorf("%s cannot be empty", key)
	}
	
	return str, nil
}

// Helper function to extract int from arguments
func getInt(args map[string]interface{}, key string, defaultVal int) (int, error) {
	val, ok := args[key]
	if !ok {
		return defaultVal, nil
	}
	
	// JSON numbers come as float64
	if fVal, ok := val.(float64); ok {
		return int(fVal), nil
	}
	
	// Try direct int
	if iVal, ok := val.(int); ok {
		return iVal, nil
	}
	
	return defaultVal, fmt.Errorf("%s must be a number", key)
}

// Helper function to extract bool from arguments
func getBool(args map[string]interface{}, key string, defaultVal bool) bool {
	val, ok := args[key]
	if !ok {
		return defaultVal
	}
	
	bVal, ok := val.(bool)
	if !ok {
		return defaultVal
	}
	
	return bVal
}

// Helper function to extract string array from arguments
func getStringArray(args map[string]interface{}, key string, required bool) ([]string, error) {
	val, ok := args[key]
	if !ok {
		if required {
			return nil, fmt.Errorf("%s parameter is required", key)
		}
		return []string{}, nil
	}
	
	// Handle []interface{} from JSON
	if arr, ok := val.([]interface{}); ok {
		result := make([]string, len(arr))
		for i, item := range arr {
			if str, ok := item.(string); ok {
				result[i] = str
			} else {
				return nil, fmt.Errorf("%s array must contain strings", key)
			}
		}
		return result, nil
	}
	
	// Handle []string directly
	if arr, ok := val.([]string); ok {
		return arr, nil
	}
	
	return nil, fmt.Errorf("%s must be an array of strings", key)
}

// Helper function to extract 2D string array from arguments
func getStringArray2D(args map[string]interface{}, key string, required bool) ([][]string, error) {
	val, ok := args[key]
	if !ok {
		if required {
			return nil, fmt.Errorf("%s parameter is required", key)
		}
		return [][]string{}, nil
	}
	
	// Handle []interface{} from JSON
	if arr, ok := val.([]interface{}); ok {
		result := make([][]string, len(arr))
		for i, row := range arr {
			if rowArr, ok := row.([]interface{}); ok {
				rowStrings := make([]string, len(rowArr))
				for j, item := range rowArr {
					if str, ok := item.(string); ok {
						rowStrings[j] = str
					} else {
						return nil, fmt.Errorf("%s array must contain strings", key)
					}
				}
				result[i] = rowStrings
			} else {
				return nil, fmt.Errorf("%s must be a 2D array", key)
			}
		}
		return result, nil
	}
	
	return nil, fmt.Errorf("%s must be a 2D array of strings", key)
}

// Helper to build success response
func successResponse(message string) *protocol.CallToolResponse {
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: message,
			},
		},
	}
}

// Helper to build JSON response
func jsonResponse(data interface{}) (*protocol.CallToolResponse, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}