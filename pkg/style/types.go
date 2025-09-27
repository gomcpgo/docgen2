package style

// StyleConfig represents the complete styling configuration for a document
type StyleConfig struct {
	Fonts   FontConfig   `yaml:"fonts"`
	Colors  ColorConfig  `yaml:"colors"`
	Page    PageConfig   `yaml:"page"`
	Spacing SpacingConfig `yaml:"spacing"`
	Header  HeaderConfig `yaml:"header"`
	Footer  FooterConfig `yaml:"footer"`
}

// FontConfig defines font settings
type FontConfig struct {
	BodyFamily      string               `yaml:"body_family"`
	HeadingFamily   string               `yaml:"heading_family"`
	MonospaceFamily string               `yaml:"monospace_family"`
	BodySize        int                  `yaml:"body_size"`
	HeadingSizes    map[string]int       `yaml:"heading_sizes"`
}

// ColorConfig defines color settings (RGB format)
type ColorConfig struct {
	BodyText    string `yaml:"body_text"`
	HeadingText string `yaml:"heading_text"`
}

// PageConfig defines page layout settings
type PageConfig struct {
	Size        string        `yaml:"size"`        // a4, letter, legal
	Orientation string        `yaml:"orientation"` // portrait, landscape
	Margins     MarginConfig  `yaml:"margins"`
}

// MarginConfig defines page margins in points
type MarginConfig struct {
	Top    int `yaml:"top"`
	Bottom int `yaml:"bottom"`
	Left   int `yaml:"left"`
	Right  int `yaml:"right"`
}

// SpacingConfig defines spacing settings
type SpacingConfig struct {
	LineSpacing      float64 `yaml:"line_spacing"`      // Multiplier
	ParagraphSpacing int     `yaml:"paragraph_spacing"` // Points before/after paragraphs
}

// HeaderConfig defines header settings
type HeaderConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Content  string `yaml:"content"`  // Template with variables like {title}, {author}, {page}
	Align    string `yaml:"align"`    // left, center, right
	FontSize int    `yaml:"font_size"`
}

// FooterConfig defines footer settings
type FooterConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Content  string `yaml:"content"`  // Template with variables like {title}, {author}, {page}
	Align    string `yaml:"align"`    // left, center, right
	FontSize int    `yaml:"font_size"`
}

// GetDefaultStyle returns the hard-coded default style configuration
func GetDefaultStyle() StyleConfig {
	return StyleConfig{
		Fonts: FontConfig{
			BodyFamily:      "Times New Roman",
			HeadingFamily:   "Arial",
			MonospaceFamily: "Courier New",
			BodySize:        11,
			HeadingSizes: map[string]int{
				"h1": 20,
				"h2": 16,
				"h3": 14,
				"h4": 12,
				"h5": 11,
				"h6": 10,
			},
		},
		Colors: ColorConfig{
			BodyText:    "0,0,0",
			HeadingText: "0,0,0",
		},
		Page: PageConfig{
			Size:        "a4",
			Orientation: "portrait",
			Margins: MarginConfig{
				Top:    72,
				Bottom: 72,
				Left:   72,
				Right:  72,
			},
		},
		Spacing: SpacingConfig{
			LineSpacing:      1.2,
			ParagraphSpacing: 6,
		},
		Header: HeaderConfig{
			Enabled:  true,
			Content:  "{title}",
			Align:    "center",
			FontSize: 10,
		},
		Footer: FooterConfig{
			Enabled:  true,
			Content:  "Page {page}",
			Align:    "center",
			FontSize: 10,
		},
	}
}