package export

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/savant/mcp-servers/docgen2/pkg/style"
)

// TestPDFStylingWithPageNumbers tests that page numbers are properly rendered in PDFs
func TestPDFStylingWithPageNumbers(t *testing.T) {
	// Create temp directory for test files
	tempDir, err := ioutil.TempDir("", "docgen2-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test markdown content with multiple pages
	markdownContent := `# Test Document

## Introduction
This is a test document to verify PDF styling and page numbers work correctly.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor 
incididunt ut labore et dolore magna aliqua.

---

## Chapter 1: Testing Headers and Footers

This page should show proper header and footer content with page numbers.

Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut 
aliquip ex ea commodo consequat.

---

## Chapter 2: Verifying Page Count

This is page 3 of the document. The footer should show "Page 3 of X" where X 
is the total page count.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore 
eu fugiat nulla pariatur.

---

## Conclusion

Final page to test that total page count is correct.

Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia 
deserunt mollit anim id est laborum.
`

	testCases := []struct {
		name        string
		styleConfig style.StyleConfig
		title       string
		author      string
		description string
	}{
		{
			name: "basic_page_numbers",
			styleConfig: style.StyleConfig{
				Fonts: style.FontConfig{
					BodyFamily:      "Times Roman",
					HeadingFamily:   "Helvetica",
					MonospaceFamily: "Courier",
					BodySize:        12,
				},
				Colors: style.ColorConfig{
					BodyText:    "40,40,40",
					HeadingText: "150,50,100",
				},
				Page: style.PageConfig{
					Margins: style.MarginConfig{
						Top:    72,
						Bottom: 72,
						Left:   72,
						Right:  72,
					},
				},
				Footer: style.FooterConfig{
					Enabled:  true,
					Content:  "Page {page}",
					Align:    "center",
					FontSize: 10,
				},
			},
			title:       "Test Document",
			author:      "Test Author",
			description: "Basic page numbering",
		},
		{
			name: "page_x_of_y",
			styleConfig: style.StyleConfig{
				Fonts: style.FontConfig{
					BodyFamily:      "Times Roman",
					HeadingFamily:   "Arial",
					MonospaceFamily: "Courier",
					BodySize:        11,
				},
				Colors: style.ColorConfig{
					BodyText:    "0,0,0",
					HeadingText: "0,0,128",
				},
				Page: style.PageConfig{
					Margins: style.MarginConfig{
						Top:    50,
						Bottom: 50,
						Left:   50,
						Right:  50,
					},
				},
				Header: style.HeaderConfig{
					Enabled:  true,
					Content:  "{title}",
					Align:    "center",
					FontSize: 12,
				},
				Footer: style.FooterConfig{
					Enabled:  true,
					Content:  "Page {page} of {total_pages}",
					Align:    "center",
					FontSize: 10,
				},
			},
			title:       "Multi-Page Document",
			author:      "Test Author",
			description: "Page X of Y format",
		},
		{
			name: "emoji_header_footer",
			styleConfig: style.StyleConfig{
				Fonts: style.FontConfig{
					BodyFamily:      "Times Roman",
					HeadingFamily:   "Helvetica",
					MonospaceFamily: "Courier",
					BodySize:        14,
				},
				Colors: style.ColorConfig{
					BodyText:    "40,40,40",
					HeadingText: "150,50,100",
				},
				Page: style.PageConfig{
					Margins: style.MarginConfig{
						Top:    100,
						Bottom: 100,
						Left:   100,
						Right:  100,
					},
				},
				Header: style.HeaderConfig{
					Enabled:  true,
					Content:  "📖 {title} - {author}",
					Align:    "left",
					FontSize: 12,
				},
				Footer: style.FooterConfig{
					Enabled:  true,
					Content:  "Created by AI Storybook Creator • Page {page}",
					Align:    "right",
					FontSize: 9,
				},
			},
			title:       "Book Test Document",
			author:      "AI Storybook Creator",
			description: "Emoji handling in headers/footers",
		},
	}

	// Create Pandoc wrapper
	wrapper := NewPandocWrapper()

	// Check if Pandoc is installed
	if err := wrapper.CheckPandocInstalled(); err != nil {
		t.Skip("Pandoc not installed, skipping PDF generation tests")
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, fmt.Sprintf("%s.pdf", tc.name))

			// Generate PDF with styling
			err := wrapper.ConvertMarkdownToFormatWithStyle(
				markdownContent,
				outputPath,
				"pdf",
				tempDir,
				tc.styleConfig,
				tc.title,
				tc.author,
			)

			if err != nil {
				t.Errorf("Failed to generate PDF for %s: %v", tc.description, err)
				return
			}

			// Verify PDF was created
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("PDF file not created for %s", tc.description)
				return
			}

			// Check file size (should be reasonable)
			info, err := os.Stat(outputPath)
			if err != nil {
				t.Errorf("Failed to stat PDF file: %v", err)
				return
			}

			if info.Size() < 1000 {
				t.Errorf("PDF file too small (%d bytes), likely corrupted", info.Size())
				return
			}

			t.Logf("✅ Successfully generated PDF for %s: %s (size: %d bytes)",
				tc.description, outputPath, info.Size())
		})
	}
}

// TestLaTeXHeaderGeneration tests the LaTeX header generation
func TestLaTeXHeaderGeneration(t *testing.T) {
	wrapper := NewPandocWrapper()

	styleConfig := style.StyleConfig{
		Colors: style.ColorConfig{
			BodyText:    "40,40,40",
			HeadingText: "150,50,100",
		},
		Header: style.HeaderConfig{
			Enabled:  true,
			Content:  "📖 {title} by {author}",
			Align:    "center",
			FontSize: 12,
		},
		Footer: style.FooterConfig{
			Enabled:  true,
			Content:  "Page {page} of {total_pages}",
			Align:    "center",
			FontSize: 10,
		},
	}

	header := wrapper.generateLaTeXHeader(styleConfig, "Test Title", "Test Author")

	// Check that AtBeginDocument wrapper is present
	if !strings.Contains(header, "\\AtBeginDocument{") {
		t.Error("LaTeX header missing AtBeginDocument wrapper")
	}

	// Check that lastpage package is included
	if !strings.Contains(header, "\\usepackage{lastpage}") {
		t.Error("LaTeX header missing lastpage package")
	}

	// Check that page variables are preserved
	if !strings.Contains(header, "\\thepage") {
		t.Error("LaTeX header missing thepage command")
	}

	if !strings.Contains(header, "\\pageref{LastPage}") {
		t.Error("LaTeX header missing pageref command")
	}

	// Check that emoji was removed (not converted to text)
	if !strings.Contains(header, " Test Title by Test Author") {
		t.Error("LaTeX header missing content with emoji removed")
	}

	t.Logf("✅ LaTeX header generation working correctly")
}

// TestProcessTemplateVariables tests the template variable processing
func TestProcessTemplateVariables(t *testing.T) {
	wrapper := NewPandocWrapper()

	testCases := []struct {
		template string
		title    string
		author   string
		expected []string // strings that should be in the result
	}{
		{
			template: "Page {page}",
			title:    "Test",
			author:   "Author",
			expected: []string{"Page \\thepage"},
		},
		{
			template: "Page {page} of {total_pages}",
			title:    "Test",
			author:   "Author",
			expected: []string{"Page \\thepage of \\pageref{LastPage}"},
		},
		{
			template: "{title} by {author} - {date}",
			title:    "My Book",
			author:   "John Doe",
			expected: []string{"My Book by John Doe", "\\today"},
		},
		{
			template: "Copyright {year}",
			title:    "Test",
			author:   "Author",
			expected: []string{"Copyright 2025"}, // Current year
		},
	}

	for _, tc := range testCases {
		t.Run(tc.template, func(t *testing.T) {
			result := wrapper.processTemplateVariables(tc.template, tc.title, tc.author)

			for _, expected := range tc.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected '%s' to contain '%s', got: %s", tc.template, expected, result)
				}
			}
		})
	}
}

// TestUnicodeHandling tests the generic Unicode handling functionality
func TestUnicodeHandling(t *testing.T) {
	wrapper := NewPandocWrapper()

	testCases := []struct {
		input       string
		expected    string
		description string
	}{
		// Emoji removal tests
		{"📖 Book Title", " Book Title", "Remove book emoji"},
		{"✨ Sparkle", " Sparkle", "Remove sparkle emoji"},
		{"🚀 Launch", " Launch", "Remove rocket emoji"},
		{"Hello 🌟 World 🎉", "Hello  World ", "Remove multiple emojis"},
		
		// Language preservation tests  
		{"中文测试 Chinese", "中文测试 Chinese", "Preserve Chinese characters"},
		{"اللغة العربية Arabic", "اللغة العربية Arabic", "Preserve Arabic characters"},
		{"Русский текст", "Русский текст", "Preserve Russian characters"},
		{"Français Español", "Français Español", "Preserve accented characters"},
		
		// Mixed content tests
		{"Hello 🌍 中文 🚀", "Hello  中文 ", "Remove emojis, keep Chinese"},
		{"Math ∑∫ 📊 Chart", "Math ∑∫  Chart", "Keep math symbols, remove emoji"},
		
		// LaTeX preservation
		{"Page \\thepage with 📖", "Page \\thepage with ", "LaTeX preserved, emoji removed"},
		
		// Basic symbols (should be preserved)
		{"Basic arrows: ← → ↑ ↓", "Basic arrows: ← → ↑ ↓", "Keep basic arrows"},
		
		// No problematic characters
		{"Regular English text", "Regular English text", "Unchanged regular text"},
		{"", "", "Empty string"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := wrapper.escapeLatex(tc.input)
			if result != tc.expected {
				t.Errorf("Test: %s\nInput: '%s'\nExpected: '%s'\nGot: '%s'", 
					tc.description, tc.input, tc.expected, result)
			}
		})
	}
}