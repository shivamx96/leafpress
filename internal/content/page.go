package content

import (
	"time"
)

// Page represents a content page
type Page struct {
	// Metadata from frontmatter
	Title    string
	Date     time.Time
	Tags     []string
	Draft    bool
	Growth   string // seedling | budding | evergreen
	TOC      *bool  // Override site-wide TOC setting (nil = use site default)
	ShowList *bool  // Show page list on section index (nil = true)

	// Paths
	SourcePath string // Relative path to .md file (e.g., "projects/leafpress.md")
	Slug       string // URL slug (e.g., "projects/leafpress")
	OutputPath string // Path in _site/ (e.g., "projects/leafpress/index.html")
	Permalink  string // Full URL path (e.g., "/projects/leafpress/")

	// Content
	RawContent  string // Original markdown (without frontmatter)
	HTMLContent string // Rendered HTML

	// Relationships
	Backlinks []*Page  // Pages that link to this page
	OutLinks  []string // Wiki-link targets (slugs)

	// Section
	IsIndex     bool   // Is this a section index (_index.md)?
	SectionSort string // Sort order for section pages (date|title|growth)
}

// GrowthEmoji returns the emoji for the growth stage
func (p *Page) GrowthEmoji() string {
	switch p.Growth {
	case "seedling":
		return "🌱"
	case "budding":
		return "🌿"
	case "evergreen":
		return "🌳"
	default:
		return ""
	}
}

// FormattedDate returns the date in a human-readable format
func (p *Page) FormattedDate() string {
	if p.Date.IsZero() {
		return ""
	}
	return p.Date.Format("Jan 2, 2006")
}

// ShortDate returns the date in short format
func (p *Page) ShortDate() string {
	if p.Date.IsZero() {
		return ""
	}
	return p.Date.Format("Jan 2006")
}

// ISODate returns the date in ISO format for datetime attribute
func (p *Page) ISODate() string {
	if p.Date.IsZero() {
		return ""
	}
	return p.Date.Format("2006-01-02")
}
