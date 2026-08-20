package content

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unicode"
)

// ReservedPaths are top-level names the content scan never treats as content,
// because Leafpress or the surrounding tooling owns them. Anything else is
// content; authors exclude their own folders with build.ignore.
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
}

// IsExcluded reports whether a path relative to the garden root is outside
// the content set: a reserved top-level name, a hidden entry, or an
// ignore-glob match. Directory matches prune the whole subtree.
//
// The content scan and the serve watcher share this predicate. When they
// disagree, `leafpress serve` publishes pages that `leafpress build` drops.
func IsExcluded(relPath string, ignore *IgnoreMatcher) bool {
	if relPath == "" || relPath == "." {
		return false
	}
	segments := strings.Split(filepath.ToSlash(relPath), "/")
	if ReservedPaths[segments[0]] {
		return true
	}
	for _, segment := range segments {
		if strings.HasPrefix(segment, ".") {
			return true
		}
	}
	return ignore.Match(relPath)
}

// Scanner scans the content directory for markdown files
type Scanner struct {
	rootDir string
	ignore  *IgnoreMatcher
	initErr error
}

// NewScanner creates a new content scanner. A malformed ignore pattern is
// held until Scan so the constructor keeps its single-value signature;
// Config.Validate reports the same problem earlier and more clearly.
func NewScanner(rootDir string, ignore []string) *Scanner {
	matcher, err := NewIgnoreMatcher(ignore)
	return &Scanner{rootDir: rootDir, ignore: matcher, initErr: err}
}

// fileEntry holds info needed to parse a file
type fileEntry struct {
	absPath string
	relPath string
	info    os.FileInfo
}

// Scan walks the directory tree and returns all markdown files
func (s *Scanner) Scan() ([]*Page, error) {
	if s.initErr != nil {
		return nil, s.initErr
	}

	// Phase 1: Collect file paths (fast, sequential walk)
	var files []fileEntry

	err := filepath.WalkDir(s.rootDir, func(path string, d os.DirEntry, err error) error {
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

		// Reserved names, hidden entries and configured ignore globs
		if IsExcluded(relPath, s.ignore) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process markdown files
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		// A note reached through a link that leaves the garden is not the
		// author's content to publish.
		if err := CheckSymlinkEscape(s.rootDir, path, d.Type()); err != nil {
			return err
		}

		// Get file info only for markdown files
		info, err := d.Info()
		if err != nil {
			return err
		}

		files = append(files, fileEntry{absPath: path, relPath: relPath, info: info})
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Phase 2: Parse files in parallel
	if len(files) == 0 {
		return nil, nil
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > len(files) {
		numWorkers = len(files)
	}

	pages := make([]*Page, len(files))
	fileChan := make(chan int, len(files))
	var wg sync.WaitGroup
	var parseErr error
	var errOnce sync.Once

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range fileChan {
				f := files[idx]
				page, err := s.parsePage(f.absPath, f.relPath, f.info)
				if err != nil {
					errOnce.Do(func() { parseErr = err })
					return
				}
				pages[idx] = page
			}
		}()
	}

	// Send file indices to workers
	for i := range files {
		fileChan <- i
	}
	close(fileChan)

	wg.Wait()

	if parseErr != nil {
		return nil, parseErr
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

	// Parse created date (priority: date > created > createdAt > file mod time)
	createdStr := fm.GetCreatedDate()
	created, err := ParseDate(createdStr)
	if err != nil || created.IsZero() {
		created = info.ModTime()
	}

	// Parse modified date (priority: modified > updated > updatedAt)
	modifiedStr := fm.GetModifiedDate()
	modified, _ := ParseDate(modifiedStr)
	// Note: modified can be zero if not specified

	// Date is used for display/sorting, same as created
	date := created

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
		Title:               title,
		Description:         fm.Description,
		Date:                date,
		Created:             created,
		Modified:            modified,
		Tags:                MergeTags(fm.Tags, ExtractInlineTags(body)),
		Draft:               fm.Draft,
		Growth:              fm.Growth,
		TOC:                 fm.TOC,
		ShowList:            fm.ShowList,
		Image:               fm.Image,
		SourcePath:          relPath,
		Slug:                slug,
		OutputPath:          outputPath,
		Permalink:           permalink,
		RawContent:          body,
		IsIndex:             isIndex,
		SectionSort:         fm.Sort,
		ReadingTimeOverride: fm.ReadingTime,
	}

	return page, nil
}

// ParseSingleFile parses a single markdown file and returns a Page
func ParseSingleFile(rootDir, relPath string) (*Page, error) {
	absPath := filepath.Join(rootDir, relPath)
	lstat, err := os.Lstat(absPath)
	if err != nil {
		return nil, err
	}
	// The incremental path must apply the same boundary as a full scan, or a
	// serve session would publish what a build refuses.
	if err := CheckSymlinkEscape(rootDir, absPath, lstat.Mode()); err != nil {
		return nil, err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	scanner := &Scanner{rootDir: rootDir}
	return scanner.parsePage(absPath, relPath, info)
}

// generateSlug creates a URL slug from a file path
func generateSlug(relPath string) string {
	// Remove .md extension
	slug := strings.TrimSuffix(relPath, ".md")

	// Convert to forward slashes (for Windows compatibility)
	slug = filepath.ToSlash(slug)

	// Handle the reserved _index.md basename only. A normal filename such as
	// migration_index.md must retain its own route.
	if path.Base(slug) == "_index" {
		slug = path.Dir(slug)
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
		runes := []rune(word)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}

	return strings.Join(words, " ")
}
