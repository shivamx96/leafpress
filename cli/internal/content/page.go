package content

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// Page represents a content page
type Page struct {
	// Metadata from frontmatter
	Title    string
	Date     time.Time // Primary display date (from date, created, or createdAt)
	Created  time.Time // Creation date (from created, createdAt, or date)
	Modified time.Time // Last modified date (from modified, updated, or updatedAt)
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

	// Reading time
	WordCount           int  // Total word count
	ImageCount          int  // Number of images
	ReadingTime         int  // Estimated reading time in minutes
	ReadingTimeOverride *int // Manual override from frontmatter

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

// FormattedModified returns the modified date in a human-readable format
func (p *Page) FormattedModified() string {
	if p.Modified.IsZero() {
		return ""
	}
	return p.Modified.Format("Jan 2, 2006")
}

// ISOModified returns the modified date in ISO format
func (p *Page) ISOModified() string {
	if p.Modified.IsZero() {
		return ""
	}
	return p.Modified.Format("2006-01-02")
}

// HasModified returns true if the page has a modified date different from created
func (p *Page) HasModified() bool {
	if p.Modified.IsZero() {
		return false
	}
	// Only show modified if it's different from the created/date
	if !p.Created.IsZero() {
		return !p.Modified.Equal(p.Created)
	}
	if !p.Date.IsZero() {
		return !p.Modified.Equal(p.Date)
	}
	return true
}

// DisplayDate returns the most relevant date (modified if exists, otherwise created)
func (p *Page) DisplayDate() string {
	if p.HasModified() {
		return p.Modified.Format("Jan 2006")
	}
	if p.Date.IsZero() {
		return ""
	}
	return p.Date.Format("Jan 2006")
}

// DisplayDateISO returns the most relevant date in ISO format
func (p *Page) DisplayDateISO() string {
	if p.HasModified() {
		return p.Modified.Format("2006-01-02")
	}
	if p.Date.IsZero() {
		return ""
	}
	return p.Date.Format("2006-01-02")
}

// PlainContent returns content with HTML tags stripped for search indexing
func (p *Page) PlainContent() string {
	plain := htmlTagRegex.ReplaceAllString(p.HTMLContent, " ")
	// Normalize whitespace
	plain = strings.Join(strings.Fields(plain), " ")
	// Limit to ~5000 chars for search index size
	if len(plain) > 5000 {
		plain = plain[:5000]
	}
	return plain
}

// ReadingTimeDisplay returns a human-readable reading time string
func (p *Page) ReadingTimeDisplay() string {
	if p.ReadingTime <= 0 {
		return ""
	}
	if p.ReadingTime == 1 {
		return "1 min read"
	}
	return fmt.Sprintf("%d min read", p.ReadingTime)
}
