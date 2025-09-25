package style

import (
	"fmt"
	"strings"
)

// HTMLCSSGenerator generates embedded CSS for HTML exports
type HTMLCSSGenerator struct{}

// NewHTMLCSSGenerator creates a new HTML CSS generator
func NewHTMLCSSGenerator() *HTMLCSSGenerator {
	return &HTMLCSSGenerator{}
}

// GenerateCSS generates embedded CSS styles for HTML export
func (g *HTMLCSSGenerator) GenerateCSS(style StyleConfig) string {
	var css strings.Builder
	
	// CSS reset and base styles
	css.WriteString("/* Document styling */\n")
	css.WriteString("body {\n")
	css.WriteString(fmt.Sprintf("  font-family: '%s', serif;\n", style.Fonts.BodyFamily))
	css.WriteString(fmt.Sprintf("  font-size: %dpt;\n", style.Fonts.BodySize))
	css.WriteString(fmt.Sprintf("  line-height: %.1f;\n", style.Spacing.LineSpacing))
	css.WriteString(fmt.Sprintf("  color: rgb(%s);\n", style.Colors.BodyText))
	css.WriteString(fmt.Sprintf("  margin: %dpt %dpt %dpt %dpt;\n", 
		style.Page.Margins.Top, style.Page.Margins.Right, 
		style.Page.Margins.Bottom, style.Page.Margins.Left))
	css.WriteString("  max-width: 8.5in;\n") // Standard page width
	css.WriteString("  background-color: white;\n")
	css.WriteString("}\n\n")
	
	// Heading styles
	css.WriteString("/* Heading styles */\n")
	for level := 1; level <= 6; level++ {
		headingKey := fmt.Sprintf("h%d", level)
		if size, exists := style.Fonts.HeadingSizes[headingKey]; exists {
			css.WriteString(fmt.Sprintf("h%d {\n", level))
			css.WriteString(fmt.Sprintf("  font-family: '%s', sans-serif;\n", style.Fonts.HeadingFamily))
			css.WriteString(fmt.Sprintf("  font-size: %dpt;\n", size))
			css.WriteString(fmt.Sprintf("  color: rgb(%s);\n", style.Colors.HeadingText))
			css.WriteString("  font-weight: bold;\n")
			css.WriteString(fmt.Sprintf("  margin-top: %dpt;\n", style.Spacing.ParagraphSpacing*2))
			css.WriteString(fmt.Sprintf("  margin-bottom: %dpt;\n", style.Spacing.ParagraphSpacing))
			css.WriteString("}\n\n")
		}
	}
	
	// Paragraph styles
	css.WriteString("/* Paragraph styles */\n")
	css.WriteString("p {\n")
	css.WriteString(fmt.Sprintf("  margin-top: %dpt;\n", style.Spacing.ParagraphSpacing))
	css.WriteString(fmt.Sprintf("  margin-bottom: %dpt;\n", style.Spacing.ParagraphSpacing))
	css.WriteString("}\n\n")
	
	// Code and monospace styles
	css.WriteString("/* Code styles */\n")
	css.WriteString("code, pre, tt {\n")
	css.WriteString(fmt.Sprintf("  font-family: '%s', monospace;\n", style.Fonts.MonospaceFamily))
	css.WriteString("}\n\n")
	
	// Image styles
	css.WriteString("/* Image styles */\n")
	css.WriteString("img {\n")
	css.WriteString("  max-width: 100%;\n")
	css.WriteString("  height: auto;\n")
	css.WriteString("  display: block;\n")
	css.WriteString("  margin: 1em auto;\n")
	css.WriteString("}\n\n")
	
	// Table styles
	css.WriteString("/* Table styles */\n")
	css.WriteString("table {\n")
	css.WriteString("  border-collapse: collapse;\n")
	css.WriteString("  width: 100%;\n")
	css.WriteString("  margin: 1em 0;\n")
	css.WriteString("}\n\n")
	css.WriteString("th, td {\n")
	css.WriteString("  border: 1px solid #ddd;\n")
	css.WriteString("  padding: 0.5em;\n")
	css.WriteString("  text-align: left;\n")
	css.WriteString("}\n\n")
	css.WriteString("th {\n")
	css.WriteString(fmt.Sprintf("  font-family: '%s', sans-serif;\n", style.Fonts.HeadingFamily))
	css.WriteString("  font-weight: bold;\n")
	css.WriteString("  background-color: #f5f5f5;\n")
	css.WriteString("}\n\n")
	
	// List styles
	css.WriteString("/* List styles */\n")
	css.WriteString("ul, ol {\n")
	css.WriteString(fmt.Sprintf("  margin-top: %dpt;\n", style.Spacing.ParagraphSpacing))
	css.WriteString(fmt.Sprintf("  margin-bottom: %dpt;\n", style.Spacing.ParagraphSpacing))
	css.WriteString("}\n\n")
	
	// Blockquote styles
	css.WriteString("/* Blockquote styles */\n")
	css.WriteString("blockquote {\n")
	css.WriteString("  margin: 1em 2em;\n")
	css.WriteString("  padding: 0.5em 1em;\n")
	css.WriteString("  border-left: 3px solid #ccc;\n")
	css.WriteString("  background-color: #f9f9f9;\n")
	css.WriteString("  font-style: italic;\n")
	css.WriteString("}\n\n")
	
	// Print styles for PDF generation from HTML
	css.WriteString("/* Print styles */\n")
	css.WriteString("@media print {\n")
	css.WriteString("  body {\n")
	css.WriteString("    margin: 0;\n")
	css.WriteString("    padding: 0;\n")
	css.WriteString("  }\n")
	css.WriteString("  \n")
	css.WriteString("  @page {\n")
	css.WriteString(fmt.Sprintf("    size: %s", style.Page.Size))
	if style.Page.Orientation == "landscape" {
		css.WriteString(" landscape")
	}
	css.WriteString(";\n")
	css.WriteString(fmt.Sprintf("    margin: %dpt %dpt %dpt %dpt;\n", 
		style.Page.Margins.Top, style.Page.Margins.Right, 
		style.Page.Margins.Bottom, style.Page.Margins.Left))
	css.WriteString("  }\n")
	css.WriteString("}\n\n")
	
	return css.String()
}

// GenerateHTMLTemplate generates a complete HTML template with embedded CSS
func (g *HTMLCSSGenerator) GenerateHTMLTemplate(style StyleConfig, title, author string) string {
	var html strings.Builder
	
	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html lang=\"en\">\n")
	html.WriteString("<head>\n")
	html.WriteString("  <meta charset=\"UTF-8\">\n")
	html.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	html.WriteString(fmt.Sprintf("  <title>%s</title>\n", g.escapeHTML(title)))
	if author != "" {
		html.WriteString(fmt.Sprintf("  <meta name=\"author\" content=\"%s\">\n", g.escapeHTML(author)))
	}
	html.WriteString("  <style>\n")
	html.WriteString(g.GenerateCSS(style))
	html.WriteString("  </style>\n")
	html.WriteString("</head>\n")
	html.WriteString("<body>\n")
	
	// Add header if enabled
	if style.Header.Enabled {
		headerContent := g.processHTMLTemplateVariables(style.Header.Content, title, author)
		headerAlign := g.getHTMLAlignment(style.Header.Align)
		html.WriteString(fmt.Sprintf("  <header style=\"text-align: %s; font-size: %dpt; margin-bottom: 2em;\">\n", 
			headerAlign, style.Header.FontSize))
		html.WriteString(fmt.Sprintf("    <div>%s</div>\n", headerContent))
		html.WriteString("  </header>\n")
	}
	
	html.WriteString("$body$\n")
	
	// Add footer if enabled
	if style.Footer.Enabled {
		footerContent := g.processHTMLTemplateVariables(style.Footer.Content, title, author)
		footerAlign := g.getHTMLAlignment(style.Footer.Align)
		html.WriteString(fmt.Sprintf("  <footer style=\"text-align: %s; font-size: %dpt; margin-top: 2em;\">\n", 
			footerAlign, style.Footer.FontSize))
		html.WriteString(fmt.Sprintf("    <div>%s</div>\n", footerContent))
		html.WriteString("  </footer>\n")
	}
	
	html.WriteString("</body>\n")
	html.WriteString("</html>\n")
	
	return html.String()
}

// getHTMLAlignment converts alignment string to CSS text-align value
func (g *HTMLCSSGenerator) getHTMLAlignment(align string) string {
	switch strings.ToLower(align) {
	case "left":
		return "left"
	case "right":
		return "right"
	case "center":
		return "center"
	default:
		return "center"
	}
}

// processHTMLTemplateVariables replaces template variables with HTML values
func (g *HTMLCSSGenerator) processHTMLTemplateVariables(template, title, author string) string {
	result := template
	result = strings.ReplaceAll(result, "{title}", g.escapeHTML(title))
	result = strings.ReplaceAll(result, "{author}", g.escapeHTML(author))
	result = strings.ReplaceAll(result, "{page}", "<span class=\"page-number\">1</span>") // Simplified for HTML
	return result
}

// escapeHTML escapes special HTML characters
func (g *HTMLCSSGenerator) escapeHTML(text string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(text)
}