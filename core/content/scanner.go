package content

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	rootDir     string
	ignorePaths map[string]bool
}

// NewScanner creates a new content scanner
func NewScanner(rootDir string, ignore []string) *Scanner {
	ignorePaths := make(map[string]bool)
	for _, path := range ignore {
		ignorePaths[path] = true
	}
	return &Scanner{rootDir: rootDir, ignorePaths: ignorePaths}
}

// fileEntry holds info needed to parse a file
type fileEntry struct {
	absPath string
	relPath string
	info    os.FileInfo
}

// Scan walks the directory tree and returns all markdown files
func (s *Scanner) Scan() ([]*Page, error) {
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

		// Check if this is a reserved path
		topLevel := strings.Split(relPath, string(filepath.Separator))[0]
		if ReservedPaths[topLevel] {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this path should be ignored (from config)
		if s.ignorePaths[topLevel] {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process markdown files
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
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
		Tags:                fm.Tags,
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
