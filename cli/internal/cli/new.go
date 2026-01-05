package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new page",
		Long: `Creates a new page with frontmatter template.
Supports nested paths like 'projects/my-project'.`,
		Args: cobra.ExactArgs(1),
		RunE: runNew,
	}
}

func runNew(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Convert name to slug
	slug := slugify(name)
	if slug == "" {
		return fmt.Errorf("invalid page name: %s", name)
	}

	// Determine file path
	filePath := slug + ".md"
	if !strings.HasSuffix(name, ".md") {
		filePath = slug + ".md"
	}

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file already exists: %s", filePath)
	}

	// Create parent directories if needed
	dir := filepath.Dir(filePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate title from name
	title := generateTitle(filepath.Base(slug))

	// Create frontmatter
	content := fmt.Sprintf(`---
title: "%s"
date: %s
tags: []
draft: true
growth: "seedling"
---

`, title, time.Now().Format("2006-01-02"))

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Created %s\n", filePath)
	return nil
}

// slugify converts a name to a URL-safe slug
func slugify(name string) string {
	// Remove .md extension if present
	name = strings.TrimSuffix(name, ".md")

	var result strings.Builder
	lastWasHyphen := false

	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
			lastWasHyphen = false
		} else if r == '/' {
			result.WriteRune('/')
			lastWasHyphen = false
		} else if !lastWasHyphen {
			result.WriteRune('-')
			lastWasHyphen = true
		}
	}

	s := result.String()
	s = strings.Trim(s, "-")
	return s
}

// generateTitle converts a slug to a human-readable title
func generateTitle(slug string) string {
	words := strings.Split(slug, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
