package blocks

// BlockType represents the type of content block
type BlockType string

const (
	TypeHeading   BlockType = "heading"
	TypeMarkdown  BlockType = "markdown"
	TypeImage     BlockType = "image"
	TypeTable     BlockType = "table"
	TypePageBreak BlockType = "page_break"
)

// Block is the interface for all block types
type Block interface {
	GetID() string
	GetType() BlockType
	ToMarkdown() string
}

// BaseBlock contains common fields for all blocks
type BaseBlock struct {
	ID   string    `yaml:"id"`
	Type BlockType `yaml:"type"`
}

// HeadingBlock represents a heading in the document
type HeadingBlock struct {
	BaseBlock `yaml:",inline"`
	Level     int    `yaml:"level"` // 1-6 for h1-h6
	Text      string `yaml:"text"`
}

func (h *HeadingBlock) GetID() string       { return h.ID }
func (h *HeadingBlock) GetType() BlockType  { return TypeHeading }
func (h *HeadingBlock) ToMarkdown() string {
	prefix := ""
	for i := 0; i < h.Level; i++ {
		prefix += "#"
	}
	return prefix + " " + h.Text
}

// MarkdownBlock represents markdown content
type MarkdownBlock struct {
	BaseBlock `yaml:",inline"`
	Content   string `yaml:"-"` // Content stored in separate .md file
}

func (m *MarkdownBlock) GetID() string       { return m.ID }
func (m *MarkdownBlock) GetType() BlockType  { return TypeMarkdown }
func (m *MarkdownBlock) ToMarkdown() string  { return m.Content }

// ImageBlock represents an image with metadata
type ImageBlock struct {
	BaseBlock `yaml:",inline"`
	Path      string `yaml:"path"`     // Relative path to image in assets folder
	Caption   string `yaml:"caption,omitempty"`
	AltText   string `yaml:"alt_text,omitempty"`
}

func (i *ImageBlock) GetID() string       { return i.ID }
func (i *ImageBlock) GetType() BlockType  { return TypeImage }
func (i *ImageBlock) ToMarkdown() string {
	alt := i.AltText
	if alt == "" {
		alt = i.Caption
	}
	markdown := "![" + alt + "](" + i.Path + ")"
	if i.Caption != "" {
		markdown += "\n\n*" + i.Caption + "*"
	}
	return markdown
}

// TableBlock represents a table with structured data
type TableBlock struct {
	BaseBlock `yaml:",inline"`
	Headers   []string   `yaml:"headers"`
	Rows      [][]string `yaml:"rows"`
}

func (t *TableBlock) GetID() string       { return t.ID }
func (t *TableBlock) GetType() BlockType  { return TypeTable }
func (t *TableBlock) ToMarkdown() string {
	if len(t.Headers) == 0 {
		return ""
	}
	
	markdown := "| " 
	for _, h := range t.Headers {
		markdown += h + " | "
	}
	markdown += "\n|"
	for range t.Headers {
		markdown += " --- |"
	}
	markdown += "\n"
	
	for _, row := range t.Rows {
		markdown += "| "
		for i, cell := range row {
			if i < len(t.Headers) {
				markdown += cell + " | "
			}
		}
		markdown += "\n"
	}
	
	return markdown
}

// PageBreakBlock represents a page break
type PageBreakBlock struct {
	BaseBlock `yaml:",inline"`
}

func (p *PageBreakBlock) GetID() string       { return p.ID }
func (p *PageBreakBlock) GetType() BlockType  { return TypePageBreak }
func (p *PageBreakBlock) ToMarkdown() string  { return "\\newpage" }

// BlockReference stores the reference to a block in the manifest
type BlockReference struct {
	ID   string    `yaml:"id"`
	Type BlockType `yaml:"type"`
	File string    `yaml:"file"`
}