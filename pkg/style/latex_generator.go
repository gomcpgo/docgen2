package style

import (
	"fmt"
	"strings"
)

// LaTeXTemplateGenerator generates LaTeX templates with custom styling
type LaTeXTemplateGenerator struct{}

// NewLaTeXTemplateGenerator creates a new LaTeX template generator
func NewLaTeXTemplateGenerator() *LaTeXTemplateGenerator {
	return &LaTeXTemplateGenerator{}
}

// GenerateTemplate generates a complete LaTeX template with the given style
func (g *LaTeXTemplateGenerator) GenerateTemplate(style StyleConfig, title, author string) string {
	var template strings.Builder
	
	// Document class and packages
	template.WriteString("\\documentclass[")
	template.WriteString(fmt.Sprintf("%dpt", style.Fonts.BodySize))
	template.WriteString(",")
	template.WriteString(style.Page.Size)
	template.WriteString("paper")
	if style.Page.Orientation == "landscape" {
		template.WriteString(",landscape")
	}
	template.WriteString("]{article}\n\n")
	
	// Essential packages
	template.WriteString("% Essential packages\n")
	template.WriteString("\\usepackage[utf8]{inputenc}\n")
	template.WriteString("\\usepackage[T1]{fontenc}\n")
	template.WriteString("\\usepackage{fontspec}\n")
	template.WriteString("\\usepackage{xcolor}\n")
	template.WriteString("\\usepackage{graphicx}\n")
	template.WriteString("\\usepackage{fancyhdr}\n")
	template.WriteString("\\usepackage{geometry}\n")
	template.WriteString("\\usepackage{setspace}\n")
	template.WriteString("\\usepackage{titlesec}\n\n")
	
	// Page geometry
	template.WriteString("% Page geometry\n")
	template.WriteString("\\geometry{\n")
	template.WriteString(fmt.Sprintf("  top=%dpt,\n", style.Page.Margins.Top))
	template.WriteString(fmt.Sprintf("  bottom=%dpt,\n", style.Page.Margins.Bottom))
	template.WriteString(fmt.Sprintf("  left=%dpt,\n", style.Page.Margins.Left))
	template.WriteString(fmt.Sprintf("  right=%dpt\n", style.Page.Margins.Right))
	template.WriteString("}\n\n")
	
	// Font settings
	template.WriteString("% Font settings\n")
	template.WriteString(fmt.Sprintf("\\setmainfont{%s}\n", style.Fonts.BodyFamily))
	template.WriteString(fmt.Sprintf("\\setsansfont{%s}\n", style.Fonts.HeadingFamily))
	template.WriteString(fmt.Sprintf("\\setmonofont{%s}\n", style.Fonts.MonospaceFamily))
	template.WriteString("\n")
	
	// Colors
	template.WriteString("% Color definitions\n")
	bodyRGB := g.parseRGBColor(style.Colors.BodyText)
	headingRGB := g.parseRGBColor(style.Colors.HeadingText)
	template.WriteString(fmt.Sprintf("\\definecolor{bodycolor}{RGB}{%s}\n", bodyRGB))
	template.WriteString(fmt.Sprintf("\\definecolor{headingcolor}{RGB}{%s}\n", headingRGB))
	template.WriteString("\\color{bodycolor}\n\n")
	
	// Line spacing
	template.WriteString("% Line spacing\n")
	template.WriteString(fmt.Sprintf("\\setstretch{%.1f}\n\n", style.Spacing.LineSpacing))
	
	// Paragraph spacing
	template.WriteString("% Paragraph spacing\n")
	template.WriteString(fmt.Sprintf("\\setlength{\\parskip}{%dpt}\n", style.Spacing.ParagraphSpacing))
	template.WriteString("\\setlength{\\parindent}{0pt}\n\n")
	
	// Title formatting
	template.WriteString("% Heading styles\n")
	processedSections := make(map[string]bool)
	for level := 1; level <= 6; level++ {
		headingKey := fmt.Sprintf("h%d", level)
		if size, exists := style.Fonts.HeadingSizes[headingKey]; exists {
			sectionName := g.getSectionName(level)
			// Avoid duplicate section formatting
			if !processedSections[sectionName] {
				template.WriteString(fmt.Sprintf("\\titleformat{\\%s}\n", sectionName))
				template.WriteString("  {\\sffamily\\bfseries\\color{headingcolor}}\n")
				template.WriteString(fmt.Sprintf("  {\\the%s}\n", sectionName))
				template.WriteString("  {1em}\n")
				template.WriteString(fmt.Sprintf("  {\\fontsize{%d}{%d}\\selectfont}\n", size, int(float64(size)*1.2)))
				template.WriteString("\n")
				processedSections[sectionName] = true
			}
		}
	}
	
	// Header and footer
	template.WriteString("% Header and footer\n")
	template.WriteString("\\pagestyle{fancy}\n")
	template.WriteString("\\fancyhf{}\n")
	
	if style.Header.Enabled {
		headerContent := g.processTemplateVariables(style.Header.Content, title, author)
		headerPosition := g.getHeaderAlignment(style.Header.Align)
		template.WriteString(fmt.Sprintf("\\%s{\\fontsize{%d}{%d}\\selectfont %s}\n", 
			headerPosition, style.Header.FontSize, int(float64(style.Header.FontSize)*1.2), headerContent))
	}
	
	if style.Footer.Enabled {
		footerContent := g.processTemplateVariables(style.Footer.Content, title, author)
		footerPosition := g.getFooterAlignment(style.Footer.Align)
		template.WriteString(fmt.Sprintf("\\%s{\\fontsize{%d}{%d}\\selectfont %s}\n", 
			footerPosition, style.Footer.FontSize, int(float64(style.Footer.FontSize)*1.2), footerContent))
	}
	
	template.WriteString("\\renewcommand{\\headrulewidth}{0pt}\n")
	template.WriteString("\\renewcommand{\\footrulewidth}{0pt}\n\n")
	
	// Document metadata
	template.WriteString("% Document metadata\n")
	template.WriteString(fmt.Sprintf("\\title{%s}\n", g.escapeLatex(title)))
	if author != "" {
		template.WriteString(fmt.Sprintf("\\author{%s}\n", g.escapeLatex(author)))
	}
	template.WriteString("\\date{}\n\n")
	
	// Document begin
	template.WriteString("\\begin{document}\n")
	template.WriteString("\n")
	template.WriteString("$if(has-frontmatter)$\n")
	template.WriteString("\\frontmatter\n")
	template.WriteString("$endif$\n")
	template.WriteString("$if(title)$\n")
	template.WriteString("\\maketitle\n")
	template.WriteString("$endif$\n")
	template.WriteString("$if(has-frontmatter)$\n")
	template.WriteString("\\mainmatter\n")
	template.WriteString("$endif$\n")
	template.WriteString("\n")
	template.WriteString("$body$\n")
	template.WriteString("\n")
	template.WriteString("$if(has-frontmatter)$\n")
	template.WriteString("\\backmatter\n")
	template.WriteString("$endif$\n")
	template.WriteString("\\end{document}\n")
	
	return template.String()
}

// parseRGBColor converts "R,G,B" string to LaTeX RGB format
func (g *LaTeXTemplateGenerator) parseRGBColor(rgbString string) string {
	// rgbString format: "255,128,0"
	return rgbString
}

// getSectionName returns the LaTeX section command name for the given heading level
func (g *LaTeXTemplateGenerator) getSectionName(level int) string {
	switch level {
	case 1:
		return "section"
	case 2:
		return "subsection"
	case 3:
		return "subsubsection"
	case 4:
		return "paragraph"
	case 5:
		return "subparagraph"
	default:
		return "subparagraph"
	}
}

// getHeaderAlignment converts alignment string to LaTeX header command
func (g *LaTeXTemplateGenerator) getHeaderAlignment(align string) string {
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

// getFooterAlignment converts alignment string to LaTeX footer command  
func (g *LaTeXTemplateGenerator) getFooterAlignment(align string) string {
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

// processTemplateVariables replaces template variables with actual values
func (g *LaTeXTemplateGenerator) processTemplateVariables(template, title, author string) string {
	result := template
	result = strings.ReplaceAll(result, "{title}", g.escapeLatex(title))
	result = strings.ReplaceAll(result, "{author}", g.escapeLatex(author))
	result = strings.ReplaceAll(result, "{page}", "\\thepage")
	return result
}

// escapeLatex escapes special LaTeX characters
func (g *LaTeXTemplateGenerator) escapeLatex(text string) string {
	replacer := strings.NewReplacer(
		"\\", "\\textbackslash{}",
		"{", "\\{",
		"}", "\\}",
		"$", "\\$",
		"&", "\\&",
		"%", "\\%",
		"#", "\\#",
		"^", "\\textasciicircum{}",
		"_", "\\_",
		"~", "\\textasciitilde{}",
	)
	return replacer.Replace(text)
}