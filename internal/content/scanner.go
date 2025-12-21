package content

import (
	"os"
	"path/filepath"
	"strings"
)

// ReservedPaths contains paths that should be ignored during content scanning
var ReservedPaths = map[string]bool{
	"leafpress.json": true,
	"style.css":      true,
	"static":         true,
	"_site":          true,
	".leafpress":     true,
	".git":           true,
	".gitignore":     true,
	".obsidian":      true,
	"node_modules":   true,
	"docs":           true, // Ignore docs folder
}

// Scanner scans the content directory for markdown files
type Scanner struct {
	rootDir string
}

// NewScanner creates a new content scanner
func NewScanner(rootDir string) *Scanner {
	return &Scanner{rootDir: rootDir}
}

// Scan walks the directory tree and returns all markdown files
func (s *Scanner) Scan() ([]*Page, error) {
	var pages []*Page

	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(s.rootDir, path)
		if err != nil {
			return err
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		// Check if this is a reserved path
		topLevel := strings.Split(relPath, string(filepath.Separator))[0]
		if ReservedPaths[topLevel] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process markdown files
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		// Read and parse the file
		page, err := s.parsePage(path, relPath, info)
		if err != nil {
			return err
		}

		pages = append(pages, page)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return pages, nil
}

// parsePage reads and parses a markdown file into a Page
func (s *Scanner) parsePage(absPath, relPath string, info os.FileInfo) (*Page, error) {
	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// Parse frontmatter
	fm, body, err := ParseFrontmatter(string(content))
	if err != nil {
		return nil, err
	}

	// Parse date
	date, err := ParseDate(fm.Date)
	if err != nil {
		// Use file modification time as fallback
		date = info.ModTime()
	}
	if date.IsZero() {
		date = info.ModTime()
	}

	// Generate slug
	slug := generateSlug(relPath)

	// Generate title from filename if not set
	title := fm.Title
	if title == "" {
		title = generateTitleFromSlug(filepath.Base(slug))
	}

	// Check if this is a section index
	isIndex := filepath.Base(relPath) == "_index.md"

	// Generate output path and permalink
	outputPath := generateOutputPath(slug, isIndex)
	permalink := generatePermalink(slug, isIndex)

	page := &Page{
		Title:       title,
		Date:        date,
		Tags:        fm.Tags,
		Draft:       fm.Draft,
		Growth:      fm.Growth,
		TOC:         fm.TOC,
		ShowList:    fm.ShowList,
		SourcePath:  relPath,
		Slug:        slug,
		OutputPath:  outputPath,
		Permalink:   permalink,
		RawContent:  body,
		IsIndex:     isIndex,
		SectionSort: fm.Sort,
	}

	return page, nil
}

// generateSlug creates a URL slug from a file path
func generateSlug(relPath string) string {
	// Remove .md extension
	slug := strings.TrimSuffix(relPath, ".md")

	// Convert to forward slashes (for Windows compatibility)
	slug = filepath.ToSlash(slug)

	// Handle _index.md -> use parent directory
	if strings.HasSuffix(slug, "_index") {
		slug = filepath.Dir(slug)
		if slug == "." {
			slug = ""
		}
	}

	// Handle index.md at root
	if slug == "index" {
		slug = ""
	}

	return slug
}

// generateOutputPath creates the output file path
func generateOutputPath(slug string, isIndex bool) string {
	if slug == "" {
		return "index.html"
	}
	return filepath.Join(slug, "index.html")
}

// generatePermalink creates the URL permalink
func generatePermalink(slug string, isIndex bool) string {
	if slug == "" {
		return "/"
	}
	return "/" + slug + "/"
}

// generateTitleFromSlug creates a title from a slug
func generateTitleFromSlug(slug string) string {
	// Remove _index suffix
	slug = strings.TrimSuffix(slug, "_index")

	// Replace hyphens with spaces
	title := strings.ReplaceAll(slug, "-", " ")

	// Capitalize first letter of each word
	words := strings.Fields(title)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}

	return strings.Join(words, " ")
}
