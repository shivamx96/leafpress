package content

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents the YAML frontmatter of a page
type Frontmatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"` // SEO meta description
	Date        string   `yaml:"date"`
	Tags        []string `yaml:"tags"`
	Draft       bool     `yaml:"draft"`
	Growth      string   `yaml:"growth"`
	Sort        string   `yaml:"sort"`     // For _index.md files
	TOC         *bool    `yaml:"toc"`      // Override site-wide TOC setting (nil = use site default)
	ShowList    *bool    `yaml:"showList"` // Show page list on section index (nil = true)
	Image       string   `yaml:"image"`    // OG image override for this page

	// Obsidian-compatible date aliases
	Created   string `yaml:"created"`   // Alias for date (creation date)
	CreatedAt string `yaml:"createdAt"` // Alias for date (creation date)
	Modified  string `yaml:"modified"`  // Last modified date
	Updated   string `yaml:"updated"`   // Alias for modified
	UpdatedAt string `yaml:"updatedAt"` // Alias for modified

	// Reading time override
	ReadingTime *int `yaml:"readingTime"` // Manual override for reading time in minutes
}

// ParseFrontmatter extracts frontmatter and content from markdown
func ParseFrontmatter(content string) (*Frontmatter, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	// Scanner's default 64 KiB token limit otherwise truncates a long Markdown
	// line without the body collection loop noticing. The complete input is
	// already resident in memory, so allow a token as large as that input and
	// still check Scanner.Err at each termination point.
	scanner.Buffer(make([]byte, min(len(content)+1, bufio.MaxScanTokenSize)), len(content)+1)

	// Check for frontmatter delimiter
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, "", fmt.Errorf("read first content line: %w", err)
		}
		return &Frontmatter{}, content, nil
	}

	firstLine := scanner.Text()
	if firstLine != "---" {
		// No frontmatter
		return &Frontmatter{}, content, nil
	}

	// Read frontmatter lines
	var fmLines []string
	foundEnd := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			foundEnd = true
			break
		}
		fmLines = append(fmLines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("read frontmatter: %w", err)
	}

	if !foundEnd {
		return nil, "", fmt.Errorf("unclosed frontmatter: missing closing ---")
	}

	// Parse YAML
	fm := &Frontmatter{}
	fmContent := strings.Join(fmLines, "\n")
	if err := yaml.Unmarshal([]byte(fmContent), fm); err != nil {
		return nil, "", fmt.Errorf("invalid frontmatter YAML: %w", err)
	}

	// Validate growth value
	if fm.Growth != "" && fm.Growth != "seedling" && fm.Growth != "budding" && fm.Growth != "evergreen" {
		return nil, "", fmt.Errorf("invalid growth value: %s (must be seedling, budding, or evergreen)", fm.Growth)
	}

	// Collect remaining content
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("read Markdown body: %w", err)
	}

	body := strings.Join(bodyLines, "\n")
	// Trim leading newlines from body
	body = strings.TrimLeft(body, "\n")

	return fm, body, nil
}

// ParseDate parses the date string from frontmatter
func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"January 2, 2006",
		"Jan 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized date format: %s", dateStr)
}

// GetCreatedDate returns the creation date with priority: date > created > createdAt
func (fm *Frontmatter) GetCreatedDate() string {
	if fm.Date != "" {
		return fm.Date
	}
	if fm.Created != "" {
		return fm.Created
	}
	if fm.CreatedAt != "" {
		return fm.CreatedAt
	}
	return ""
}

// GetModifiedDate returns the modified date with priority: modified > updated > updatedAt
func (fm *Frontmatter) GetModifiedDate() string {
	if fm.Modified != "" {
		return fm.Modified
	}
	if fm.Updated != "" {
		return fm.Updated
	}
	if fm.UpdatedAt != "" {
		return fm.UpdatedAt
	}
	return ""
}
