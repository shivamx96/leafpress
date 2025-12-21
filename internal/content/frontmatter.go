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
	Title  string   `yaml:"title"`
	Date   string   `yaml:"date"`
	Tags   []string `yaml:"tags"`
	Draft  bool     `yaml:"draft"`
	Growth string   `yaml:"growth"`
	Sort   string   `yaml:"sort"` // For _index.md files
	TOC    *bool    `yaml:"toc"`  // Override site-wide TOC setting (nil = use site default)
}

// ParseFrontmatter extracts frontmatter and content from markdown
func ParseFrontmatter(content string) (*Frontmatter, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Check for frontmatter delimiter
	if !scanner.Scan() {
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
