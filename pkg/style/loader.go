package style

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// StyleLoader handles loading and resolving style configurations
type StyleLoader struct {
	rootPath string
}

// NewStyleLoader creates a new style loader
func NewStyleLoader(rootPath string) *StyleLoader {
	return &StyleLoader{
		rootPath: rootPath,
	}
}

// LoadStyleForDocument loads the effective style configuration for a document
// Resolution priority:
// 1. Document manifest.yaml style section
// 2. Global default_style.yaml
// 3. Hard-coded defaults
func (sl *StyleLoader) LoadStyleForDocument(documentStyle *StyleConfig) StyleConfig {
	// Start with hard-coded defaults
	effectiveStyle := GetDefaultStyle()
	
	// Try to load global defaults
	if globalStyle, err := sl.loadGlobalDefaultStyle(); err == nil {
		effectiveStyle = sl.mergeStyles(effectiveStyle, globalStyle)
	}
	
	// Apply document-specific style if provided
	if documentStyle != nil {
		effectiveStyle = sl.mergeStyles(effectiveStyle, *documentStyle)
	}
	
	return effectiveStyle
}

// LoadGlobalDefaultStyle loads the global default style from default_style.yaml
func (sl *StyleLoader) LoadGlobalDefaultStyle() (StyleConfig, error) {
	return sl.loadGlobalDefaultStyle()
}

// SaveGlobalDefaultStyle saves the global default style to default_style.yaml
func (sl *StyleLoader) SaveGlobalDefaultStyle(style StyleConfig) error {
	defaultStylePath := filepath.Join(sl.rootPath, "default_style.yaml")
	
	// Create a wrapper structure for YAML output
	wrapper := struct {
		Style StyleConfig `yaml:"style"`
	}{
		Style: style,
	}
	
	data, err := yaml.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal default style: %w", err)
	}
	
	err = os.WriteFile(defaultStylePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write default style file: %w", err)
	}
	
	return nil
}

// loadGlobalDefaultStyle loads from default_style.yaml file
func (sl *StyleLoader) loadGlobalDefaultStyle() (StyleConfig, error) {
	defaultStylePath := filepath.Join(sl.rootPath, "default_style.yaml")
	
	data, err := os.ReadFile(defaultStylePath)
	if err != nil {
		return StyleConfig{}, fmt.Errorf("failed to read default style file: %w", err)
	}
	
	// Parse the wrapper structure
	var wrapper struct {
		Style StyleConfig `yaml:"style"`
	}
	
	err = yaml.Unmarshal(data, &wrapper)
	if err != nil {
		return StyleConfig{}, fmt.Errorf("failed to parse default style file: %w", err)
	}
	
	return wrapper.Style, nil
}

// mergeStyles merges two style configurations, with override taking precedence
func (sl *StyleLoader) mergeStyles(base, override StyleConfig) StyleConfig {
	result := base
	
	// Merge font config
	if override.Fonts.BodyFamily != "" {
		result.Fonts.BodyFamily = override.Fonts.BodyFamily
	}
	if override.Fonts.HeadingFamily != "" {
		result.Fonts.HeadingFamily = override.Fonts.HeadingFamily
	}
	if override.Fonts.MonospaceFamily != "" {
		result.Fonts.MonospaceFamily = override.Fonts.MonospaceFamily
	}
	if override.Fonts.BodySize != 0 {
		result.Fonts.BodySize = override.Fonts.BodySize
	}
	if override.Fonts.HeadingSizes != nil {
		if result.Fonts.HeadingSizes == nil {
			result.Fonts.HeadingSizes = make(map[string]int)
		}
		for k, v := range override.Fonts.HeadingSizes {
			result.Fonts.HeadingSizes[k] = v
		}
	}
	
	// Merge color config
	if override.Colors.BodyText != "" {
		result.Colors.BodyText = override.Colors.BodyText
	}
	if override.Colors.HeadingText != "" {
		result.Colors.HeadingText = override.Colors.HeadingText
	}
	
	// Merge page config
	if override.Page.Size != "" {
		result.Page.Size = override.Page.Size
	}
	if override.Page.Orientation != "" {
		result.Page.Orientation = override.Page.Orientation
	}
	if override.Page.Margins.Top != 0 {
		result.Page.Margins.Top = override.Page.Margins.Top
	}
	if override.Page.Margins.Bottom != 0 {
		result.Page.Margins.Bottom = override.Page.Margins.Bottom
	}
	if override.Page.Margins.Left != 0 {
		result.Page.Margins.Left = override.Page.Margins.Left
	}
	if override.Page.Margins.Right != 0 {
		result.Page.Margins.Right = override.Page.Margins.Right
	}
	
	// Merge spacing config
	if override.Spacing.LineSpacing != 0 {
		result.Spacing.LineSpacing = override.Spacing.LineSpacing
	}
	if override.Spacing.ParagraphSpacing != 0 {
		result.Spacing.ParagraphSpacing = override.Spacing.ParagraphSpacing
	}
	
	// Merge header config
	if override.Header.Content != "" {
		result.Header.Content = override.Header.Content
	}
	if override.Header.Align != "" {
		result.Header.Align = override.Header.Align
	}
	if override.Header.FontSize != 0 {
		result.Header.FontSize = override.Header.FontSize
	}
	// Note: Enabled is merged even if false, as it's a valid override
	result.Header.Enabled = override.Header.Enabled || base.Header.Enabled
	
	// Merge footer config
	if override.Footer.Content != "" {
		result.Footer.Content = override.Footer.Content
	}
	if override.Footer.Align != "" {
		result.Footer.Align = override.Footer.Align
	}
	if override.Footer.FontSize != 0 {
		result.Footer.FontSize = override.Footer.FontSize
	}
	// Note: Enabled is merged even if false, as it's a valid override
	result.Footer.Enabled = override.Footer.Enabled || base.Footer.Enabled
	
	return result
}