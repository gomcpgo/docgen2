package style

import (
	"os"
	"strings"
	"testing"
)

func TestGetDefaultStyle(t *testing.T) {
	defaultStyle := GetDefaultStyle()
	
	// Test basic structure
	if defaultStyle.Fonts.BodyFamily == "" {
		t.Error("Expected body font family to be set")
	}
	if defaultStyle.Fonts.HeadingFamily == "" {
		t.Error("Expected heading font family to be set")
	}
	if defaultStyle.Fonts.BodySize == 0 {
		t.Error("Expected body font size to be set")
	}
	if len(defaultStyle.Fonts.HeadingSizes) == 0 {
		t.Error("Expected heading sizes to be set")
	}
	
	// Test specific values
	if defaultStyle.Page.Size != "a4" {
		t.Errorf("Expected page size to be 'a4', got '%s'", defaultStyle.Page.Size)
	}
	if defaultStyle.Page.Orientation != "portrait" {
		t.Errorf("Expected orientation to be 'portrait', got '%s'", defaultStyle.Page.Orientation)
	}
}

func TestStyleLoader(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "style_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	loader := NewStyleLoader(tempDir)
	
	// Test loading with no files (should return defaults)
	style := loader.LoadStyleForDocument(nil)
	if style.Fonts.BodyFamily == "" {
		t.Error("Expected default style to be loaded")
	}
}

func TestStyleLoaderWithGlobalDefaults(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "style_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	loader := NewStyleLoader(tempDir)
	
	// Create custom global default style
	customStyle := GetDefaultStyle()
	customStyle.Fonts.BodyFamily = "Custom Font"
	customStyle.Page.Size = "letter"
	
	// Save it to the temp directory
	err = loader.SaveGlobalDefaultStyle(customStyle)
	if err != nil {
		t.Fatalf("Failed to save global default style: %v", err)
	}
	
	// Load style for a document with no specific style
	loadedStyle := loader.LoadStyleForDocument(nil)
	
	// Verify custom values are loaded
	if loadedStyle.Fonts.BodyFamily != "Custom Font" {
		t.Errorf("Expected body font to be 'Custom Font', got '%s'", loadedStyle.Fonts.BodyFamily)
	}
	if loadedStyle.Page.Size != "letter" {
		t.Errorf("Expected page size to be 'letter', got '%s'", loadedStyle.Page.Size)
	}
}

func TestStyleMerging(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "style_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	loader := NewStyleLoader(tempDir)
	
	// Create document-specific style that only overrides some values
	documentStyle := &StyleConfig{
		Fonts: FontConfig{
			BodyFamily: "Document Font",
			// Other font settings not specified, should use defaults
		},
		Colors: ColorConfig{
			BodyText: "255,0,0", // Red
			// HeadingText not specified, should use default
		},
		// Other configs not specified, should use defaults
	}
	
	// Load merged style
	mergedStyle := loader.LoadStyleForDocument(documentStyle)
	
	// Verify document overrides are applied
	if mergedStyle.Fonts.BodyFamily != "Document Font" {
		t.Errorf("Expected body font to be 'Document Font', got '%s'", mergedStyle.Fonts.BodyFamily)
	}
	if mergedStyle.Colors.BodyText != "255,0,0" {
		t.Errorf("Expected body text color to be '255,0,0', got '%s'", mergedStyle.Colors.BodyText)
	}
	
	// Verify defaults are preserved for non-overridden values
	if mergedStyle.Fonts.HeadingFamily == "" {
		t.Error("Expected heading font to use default value")
	}
	if mergedStyle.Colors.HeadingText == "" {
		t.Error("Expected heading text color to use default value")
	}
	if mergedStyle.Page.Size == "" {
		t.Error("Expected page size to use default value")
	}
}

func TestLaTeXTemplateGeneration(t *testing.T) {
	generator := NewLaTeXTemplateGenerator()
	style := GetDefaultStyle()
	
	template := generator.GenerateTemplate(style, "Test Document", "Test Author")
	
	// Basic checks for LaTeX template structure
	if template == "" {
		t.Error("Expected non-empty template")
	}
	if !strings.Contains(template, "\\documentclass") {
		t.Error("Expected LaTeX template to contain \\documentclass")
	}
	if !strings.Contains(template, "\\begin{document}") {
		t.Error("Expected LaTeX template to contain \\begin{document}")
	}
	if !strings.Contains(template, "\\end{document}") {
		t.Error("Expected LaTeX template to contain \\end{document}")
	}
	if !strings.Contains(template, "Test Document") {
		t.Error("Expected template to contain document title")
	}
	if !strings.Contains(template, "Test Author") {
		t.Error("Expected template to contain author name")
	}
}

func TestHTMLCSSGeneration(t *testing.T) {
	generator := NewHTMLCSSGenerator()
	style := GetDefaultStyle()
	
	css := generator.GenerateCSS(style)
	
	// Basic checks for CSS structure
	if css == "" {
		t.Error("Expected non-empty CSS")
	}
	if !strings.Contains(css, "body {") {
		t.Error("Expected CSS to contain body styles")
	}
	if !strings.Contains(css, "h1 {") {
		t.Error("Expected CSS to contain heading styles")
	}
	
	template := generator.GenerateHTMLTemplate(style, "Test Document", "Test Author")
	
	// Basic checks for HTML template structure
	if template == "" {
		t.Error("Expected non-empty HTML template")
	}
	if !strings.Contains(template, "<!DOCTYPE html>") {
		t.Error("Expected HTML template to contain DOCTYPE")
	}
	if !strings.Contains(template, "<style>") {
		t.Error("Expected HTML template to contain embedded styles")
	}
	if !strings.Contains(template, "Test Document") {
		t.Error("Expected template to contain document title")
	}
}