#!/bin/bash

# Source .env file if it exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Set default root folder if not specified
export DOCGEN_ROOT="${DOCGEN_ROOT:-./docgen_data}"

case "$1" in
    "build")
        echo "Building DocGen2 MCP server..."
        go build -o bin/docgen2 ./cmd
        ;;
    
    "test")
        echo "Running unit tests..."
        go test -v ./pkg/...
        ;;
    
    "integration")
        echo "Running integration tests..."
        go test -v ./test/...
        ;;
    
    "create")
        # Create a new document
        if [ -z "$2" ]; then
            echo "Usage: ./run.sh create \"Document Title\" [--chapters]"
            exit 1
        fi
        go run ./cmd -create "$2" ${3:-}
        ;;
    
    "list")
        # List all documents
        go run ./cmd -list
        ;;
    
    "overview")
        # Show document overview
        if [ -z "$2" ]; then
            echo "Usage: ./run.sh overview <doc-id>"
            exit 1
        fi
        go run ./cmd -overview "$2"
        ;;
    
    "add-heading")
        # Add heading to document
        if [ -z "$4" ]; then
            echo "Usage: ./run.sh add-heading <doc-id> <level> \"Heading Text\""
            exit 1
        fi
        go run ./cmd -add-heading "$2" -level "$3" -text "$4"
        ;;
    
    "add-markdown")
        # Add markdown to document
        if [ -z "$3" ]; then
            echo "Usage: ./run.sh add-markdown <doc-id> \"Content\""
            exit 1
        fi
        go run ./cmd -add-markdown "$2" -content "$3"
        ;;
    
    "export")
        # Export document
        if [ -z "$3" ]; then
            echo "Usage: ./run.sh export <doc-id> <format>"
            exit 1
        fi
        go run ./cmd -export "$2" -format "$3"
        ;;
    
    "run")
        echo "Running DocGen2 MCP server..."
        go run ./cmd
        ;;
    
    "clean")
        echo "Cleaning build artifacts..."
        rm -rf bin/
        rm -rf docgen_data/
        ;;
    
    *)
        echo "DocGen2 MCP Server"
        echo "Usage: $0 {build|test|integration|create|list|overview|add-heading|add-markdown|export|run|clean}"
        echo ""
        echo "Terminal Mode Commands:"
        echo "  create \"Title\" [--chapters]  - Create new document"
        echo "  list                         - List all documents"
        echo "  overview <doc-id>           - Show document structure"
        echo "  add-heading <doc-id> <level> \"Text\" - Add heading"
        echo "  add-markdown <doc-id> \"Content\" - Add markdown"
        echo "  export <doc-id> <format>    - Export document (pdf/docx/html)"
        echo ""
        echo "Development Commands:"
        echo "  build       - Build the server binary"
        echo "  test        - Run unit tests"
        echo "  integration - Run integration tests"
        echo "  run         - Run MCP server"
        echo "  clean       - Remove build artifacts and test data"
        ;;
esac