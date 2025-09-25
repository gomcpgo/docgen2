package export

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/savant/mcp-servers/docgen2/pkg/style"
)

// PandocWrapper handles Pandoc command execution
type PandocWrapper struct {
	timeout time.Duration
}

// NewPandocWrapper creates a new Pandoc wrapper
func NewPandocWrapper() *PandocWrapper {
	return &PandocWrapper{
		timeout: 30 * time.Second,
	}
}

// CheckPandocInstalled checks if Pandoc is available
func (p *PandocWrapper) CheckPandocInstalled() error {
	cmd := exec.Command("pandoc", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pandoc not found: please install pandoc to enable document export")
	}
	return nil
}

// ConvertMarkdownToFormat converts markdown to specified format using Pandoc
func (p *PandocWrapper) ConvertMarkdownToFormat(markdownContent string, outputPath string, format string) error {
	// Check if Pandoc is installed
	if err := p.CheckPandocInstalled(); err != nil {
		return err
	}

	// Build Pandoc command arguments
	args := []string{
		"-f", "markdown",
		"-o", outputPath,
	}

	// Add format-specific options
	switch format {
	case "pdf":
		// Use xelatex for Unicode support (emojis), images should work now
		args = append(args, "--pdf-engine=xelatex")
	case "docx":
		// DOCX needs no special options
	case "html":
		// Standalone HTML with embedded CSS
		args = append(args, "--standalone")
		args = append(args, "--self-contained")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Create command
	cmd := exec.Command("pandoc", args...)

	// Provide markdown content via stdin
	cmd.Stdin = bytes.NewBufferString(markdownContent)

	// Capture stderr for error messages
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			if stderr.Len() > 0 {
				return fmt.Errorf("pandoc conversion failed: %s", stderr.String())
			}
			return fmt.Errorf("pandoc conversion failed: %w", err)
		}
		return nil
	case <-time.After(p.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return fmt.Errorf("pandoc conversion timed out after %v", p.timeout)
	}
}

// GetPandocVersion returns the installed Pandoc version
func (p *PandocWrapper) GetPandocVersion() (string, error) {
	cmd := exec.Command("pandoc", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get pandoc version: %w", err)
	}
	
	// Return first line which contains version
	lines := bytes.Split(output, []byte("\n"))
	if len(lines) > 0 {
		return string(lines[0]), nil
	}
	return "unknown", nil
}

// ConvertMarkdownToFormatInDir converts markdown to specified format using Pandoc from a specific working directory
func (p *PandocWrapper) ConvertMarkdownToFormatInDir(markdownContent string, outputPath string, format string, workingDir string) error {
	// Check if Pandoc is installed
	if err := p.CheckPandocInstalled(); err != nil {
		return err
	}

	// Build Pandoc command arguments
	args := []string{
		"-f", "markdown",
		"-o", outputPath,
	}

	// Add format-specific options
	switch format {
	case "pdf":
		// Use xelatex for Unicode support (emojis), images should work now
		args = append(args, "--pdf-engine=xelatex")
	case "docx":
		// DOCX needs no special options
	case "html":
		// Standalone HTML with embedded CSS
		args = append(args, "--standalone")
		args = append(args, "--self-contained")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Create command with working directory
	cmd := exec.Command("pandoc", args...)
	cmd.Dir = workingDir

	// Provide markdown content via stdin
	cmd.Stdin = bytes.NewBufferString(markdownContent)

	// Capture stderr for error messages
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			if stderr.Len() > 0 {
				return fmt.Errorf("pandoc conversion failed: %s", stderr.String())
			}
			return fmt.Errorf("pandoc conversion failed: %w", err)
		}
		return nil
	case <-time.After(p.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return fmt.Errorf("pandoc conversion timed out after %v", p.timeout)
	}
}

// ConvertMarkdownToFormatWithStyle converts markdown to specified format with custom styling
func (p *PandocWrapper) ConvertMarkdownToFormatWithStyle(markdownContent string, outputPath string, format string, workingDir string, styleConfig style.StyleConfig, title, author string) error {
	fmt.Printf("DEBUG: Starting ConvertMarkdownToFormatWithStyle - format: %s, title: %s, author: %s\n", format, title, author)
	
	// Check if Pandoc is installed
	if err := p.CheckPandocInstalled(); err != nil {
		return err
	}

	// Build Pandoc command arguments
	args := []string{
		"-f", "markdown",
		"-o", outputPath,
	}

	// Add format-specific options and templates
	switch format {
	case "pdf":
		// Hybrid approach: Use Pandoc variables for basic styling + minimal LaTeX template for advanced features
		
		args = append(args, "--pdf-engine=xelatex")
		
		// Add basic styling through Pandoc variables (fonts, size, margins)
		args = append(args, "-V", fmt.Sprintf("mainfont=%s", styleConfig.Fonts.BodyFamily))
		args = append(args, "-V", fmt.Sprintf("sansfont=%s", styleConfig.Fonts.HeadingFamily))  
		args = append(args, "-V", fmt.Sprintf("monofont=%s", styleConfig.Fonts.MonospaceFamily))
		args = append(args, "-V", fmt.Sprintf("fontsize=%dpt", styleConfig.Fonts.BodySize))
		
		// Geometry for margins
		marginGeometry := fmt.Sprintf("margin=%dpt", styleConfig.Page.Margins.Top)
		args = append(args, "-V", fmt.Sprintf("geometry=%s", marginGeometry))
		
		// Generate minimal LaTeX header for advanced styling (colors, headers/footers)
		latexHeader := p.generateLaTeXHeader(styleConfig, title, author)
		headerPath := filepath.Join(workingDir, "header.tex")
		if err := os.WriteFile(headerPath, []byte(latexHeader), 0644); err != nil {
			return fmt.Errorf("failed to write LaTeX header: %w", err)
		}
		
		args = append(args, "--include-in-header="+headerPath)
		
		fmt.Printf("DEBUG: Using hybrid Pandoc variables + LaTeX header for PDF styling\n")
		
	case "html":
		// Generate HTML template with embedded CSS
		htmlGen := style.NewHTMLCSSGenerator()
		htmlTemplate := htmlGen.GenerateHTMLTemplate(styleConfig, title, author)
		
		// Write template to temp file
		templatePath := filepath.Join(workingDir, "template.html")
		if err := os.WriteFile(templatePath, []byte(htmlTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write HTML template: %w", err)
		}
		
		args = append(args, "--standalone")
		args = append(args, "--self-contained")
		args = append(args, "--template="+templatePath)
		
	case "docx":
		// DOCX doesn't support custom templates easily, use default for now
		// TODO: Consider generating a reference.docx with styles
		
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Debug: Print Pandoc command
	fmt.Printf("DEBUG: Pandoc command: pandoc %s\n", strings.Join(args, " "))
	fmt.Printf("DEBUG: Working directory: %s\n", workingDir)

	// Create command with working directory
	cmd := exec.Command("pandoc", args...)
	cmd.Dir = workingDir

	// Provide markdown content via stdin
	cmd.Stdin = bytes.NewBufferString(markdownContent)

	// Capture stderr for error messages
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			if stderr.Len() > 0 {
				return fmt.Errorf("pandoc conversion failed: %s", stderr.String())
			}
			return fmt.Errorf("pandoc conversion failed: %w", err)
		}
		return nil
	case <-time.After(p.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return fmt.Errorf("pandoc conversion timed out after %v", p.timeout)
	}
}
// generateLaTeXHeader creates a minimal LaTeX header for advanced styling
// FIXED VERSION: Uses \AtBeginDocument to ensure proper page number processing
func (p *PandocWrapper) generateLaTeXHeader(styleConfig style.StyleConfig, title, author string) string {
	var header strings.Builder
	
	// Color definitions
	header.WriteString("% Custom colors\n")
	headingRGB := strings.ReplaceAll(styleConfig.Colors.HeadingText, ",", ",")
	bodyRGB := strings.ReplaceAll(styleConfig.Colors.BodyText, ",", ",")
	header.WriteString(fmt.Sprintf("\\definecolor{headingcolor}{RGB}{%s}\n", headingRGB))
	header.WriteString(fmt.Sprintf("\\definecolor{bodycolor}{RGB}{%s}\n", bodyRGB))
	header.WriteString("\\color{bodycolor}\n\n")
	
	// Heading colors using sectsty package
	header.WriteString("% Heading colors\n")
	header.WriteString("\\usepackage{sectsty}\n")
	header.WriteString("\\sectionfont{\\color{headingcolor}}\n")
	header.WriteString("\\subsectionfont{\\color{headingcolor}}\n")
	header.WriteString("\\subsubsectionfont{\\color{headingcolor}}\n")
	header.WriteString("\\paragraphfont{\\color{headingcolor}}\n")
	header.WriteString("\\subparagraphfont{\\color{headingcolor}}\n\n")
	
	// Headers and footers using fancyhdr - FIXED with AtBeginDocument
	if styleConfig.Header.Enabled || styleConfig.Footer.Enabled {
		header.WriteString("% Headers and footers\n")
		header.WriteString("\\usepackage{fancyhdr}\n")
		header.WriteString("\\usepackage{lastpage}\n") // For total page count support
		
		// Use AtBeginDocument to ensure page numbering is initialized
		header.WriteString("\\AtBeginDocument{\n")
		header.WriteString("  \\pagestyle{fancy}\n")
		header.WriteString("  \\fancyhf{}\n")
		
		if styleConfig.Header.Enabled {
			headerContent := p.processTemplateVariables(styleConfig.Header.Content, title, author)
			headerPos := p.getHeaderAlignment(styleConfig.Header.Align)
			header.WriteString(fmt.Sprintf("  \\%s{\\fontsize{%d}{%d}\\selectfont %s}\n", 
				headerPos, styleConfig.Header.FontSize, int(float64(styleConfig.Header.FontSize)*1.2), 
				p.escapeLatex(headerContent)))
		}
		
		if styleConfig.Footer.Enabled {
			footerContent := p.processTemplateVariables(styleConfig.Footer.Content, title, author)
			footerPos := p.getFooterAlignment(styleConfig.Footer.Align)
			header.WriteString(fmt.Sprintf("  \\%s{\\fontsize{%d}{%d}\\selectfont %s}\n", 
				footerPos, styleConfig.Footer.FontSize, int(float64(styleConfig.Footer.FontSize)*1.2),
				p.escapeLatex(footerContent)))
		}
		
		header.WriteString("  \\renewcommand{\\headrulewidth}{0pt}\n")
		header.WriteString("  \\renewcommand{\\footrulewidth}{0pt}\n")
		header.WriteString("}\n\n")
	}
	
	return header.String()
}

// Helper methods for LaTeX generation
func (p *PandocWrapper) getHeaderAlignment(align string) string {
	switch strings.ToLower(align) {
	case "left":
		return "lhead"
	case "right":
		return "rhead"
	case "center":
		return "chead"
	default:
		return "chead"
	}
}

func (p *PandocWrapper) getFooterAlignment(align string) string {
	switch strings.ToLower(align) {
	case "left":
		return "lfoot"
	case "right":
		return "rfoot"
	case "center":
		return "cfoot"
	default:
		return "cfoot"
	}
}

// Enhanced processTemplateVariables with more options 
// Note: Escaping is handled separately to avoid double-escaping
func (p *PandocWrapper) processTemplateVariables(template, title, author string) string {
	now := time.Now()
	
	// Build replacement map with enhanced variables (no escaping here)
	replacements := map[string]string{
		"{title}":       title,
		"{author}":      author,
		"{page}":        "\\thepage",
		"{total_pages}": "\\pageref{LastPage}",
		"{date}":        "\\today",
		"{year}":        fmt.Sprintf("%d", now.Year()),
		"{month}":       now.Format("January"),
		"{day}":         fmt.Sprintf("%d", now.Day()),
	}
	
	result := template
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	
	// Special handling for combined patterns like "Page X of Y"
	result = strings.ReplaceAll(result, "Page {page} of {total_pages}", "Page \\thepage\\ of \\pageref{LastPage}")
	
	return result
}

// Enhanced escapeLatex with better special character and emoji handling
func (p *PandocWrapper) escapeLatex(text string) string {
	// Skip escaping for text that contains LaTeX commands
	if strings.Contains(text, "\\thepage") || strings.Contains(text, "\\pageref") || strings.Contains(text, "\\today") {
		// This text contains LaTeX commands, only escape emojis and special chars but preserve LaTeX
		result := text
		
		// Handle emojis first
		emojiReplacements := map[string]string{
			"📖": "Book",
			"📚": "Books", 
			"✨": "*",
			"🎯": "[target]",
			"💡": "[idea]",
			"❤️": "love",
			"⭐": "*",
			"🚀": "[rocket]",
			"📝": "[note]",
			"✅": "[done]",
			"❌": "[x]",
			"⚠️": "Warning:",
			"📌": "[pin]",
			"🔍": "[search]",
			"💻": "[computer]",
			"📊": "[chart]",
			"📈": "[graph]",
			"🎨": "[art]",
			"🔧": "[tools]",
			"📧": "[email]",
			"📅": "[calendar]",
			"⏰": "[clock]",
			"🌟": "*",
			"👍": "+1",
			"👎": "-1", 
			"➡️": "->",
			"⬅️": "<-",
			"⬆️": "^",
			"⬇️": "v",
		}
		
		for emoji, replacement := range emojiReplacements {
			result = strings.ReplaceAll(result, emoji, replacement)
		}
		
		return result
	}
	
	// Handle special LaTeX characters
	replacements := []struct{ from, to string }{
		// Order matters - do backslash first
		{"\\", "\\textbackslash{}"},
		{"{", "\\{"},
		{"}", "\\}"},
		{"$", "\\$"},
		{"&", "\\&"},
		{"%", "\\%"},
		{"#", "\\#"},
		{"^", "\\textasciicircum{}"},
		{"_", "\\_"},
		{"~", "\\textasciitilde{}"},
		{"<", "\\textless{}"},
		{">", "\\textgreater{}"},
		{"|", "\\textbar{}"},
	}
	
	result := text
	for _, r := range replacements {
		result = strings.ReplaceAll(result, r.from, r.to)
	}
	
	// Handle common emojis - replace with text equivalents
	emojiReplacements := map[string]string{
		"📖": "Book",
		"📚": "Books",
		"✨": "*",
		"🎯": "[target]",
		"💡": "[idea]",
		"❤️": "love",
		"⭐": "*",
		"🚀": "[rocket]",
		"📝": "[note]",
		"✅": "[done]",
		"❌": "[x]",
		"⚠️": "Warning:",
		"📌": "[pin]",
		"🔍": "[search]",
		"💻": "[computer]",
		"📊": "[chart]",
		"📈": "[graph]",
		"🎨": "[art]",
		"🔧": "[tools]",
		"📧": "[email]",
		"📅": "[calendar]",
		"⏰": "[clock]",
		"🌟": "*",
		"👍": "+1",
		"👎": "-1",
		"➡️": "->",
		"⬅️": "<-",
		"⬆️": "^",
		"⬇️": "v",
	}
	
	for emoji, replacement := range emojiReplacements {
		result = strings.ReplaceAll(result, emoji, replacement)
	}
	
	return result
}
