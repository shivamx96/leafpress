package content

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// Page represents a content page
type Page struct {
	// Metadata from frontmatter
	Title       string
	Description string    // SEO meta description (from frontmatter or auto-generated)
	Date        time.Time // Primary display date (from date, created, or createdAt)
	Created     time.Time // Creation date (from created, createdAt, or date)
	Modified    time.Time // Last modified date (from modified, updated, or updatedAt)
	Tags        []string
	Draft       bool
	Growth      string // seedling | budding | evergreen
	TOC         *bool  // Override site-wide TOC setting (nil = use site default)
	ShowList    *bool  // Show page list on section index (nil = true)
	Image       string // OG image for this page (from frontmatter)

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
	// Decode HTML entities (e.g., &amp; -> &, &#34; -> ")
	plain = html.UnescapeString(plain)
	// Normalize whitespace
	plain = strings.Join(strings.Fields(plain), " ")
	// Limit to ~5000 chars for search index size
	plain, _ = truncateRunes(plain, 5000)
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

// SEODescription returns the description for meta tags
// Uses frontmatter description if set, otherwise auto-generates from content
func (p *Page) SEODescription() string {
	if p.Description != "" {
		return p.Description
	}
	// Auto-generate from content (first ~155 chars)
	plain := p.PlainContent()
	if truncated, didTruncate := truncateRunes(plain, 155); didTruncate {
		// Try to break at word boundary
		plain = truncated
		if lastSpace := strings.LastIndex(plain, " "); lastSpace > 100 {
			plain = plain[:lastSpace]
		}
		plain += "..."
	}
	return plain
}

func truncateRunes(value string, limit int) (string, bool) {
	if limit < 0 || utf8.RuneCountInString(value) <= limit {
		return value, false
	}
	return string([]rune(value)[:limit]), true
}

// NormalizeTags preserves the first spelling of each tag while removing
// case-variant duplicates from a single page.
func NormalizeTags(tags []string) []string {
	if len(tags) < 2 {
		return tags
	}
	seen := make(map[string]struct{}, len(tags))
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		key := strings.ToLower(tag)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, tag)
	}
	return normalized
}
