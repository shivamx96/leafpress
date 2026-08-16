package cli

import (
	"errors"
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

	// Keep the destination lexical path beneath the project. os.Root below
	// enforces the same boundary while following existing symlinks.
	filePath := filepath.FromSlash(slug + ".md")
	if !filepath.IsLocal(filePath) {
		return fmt.Errorf("invalid page name %q: destination must stay inside the project", name)
	}

	project, err := os.OpenRoot(".")
	if err != nil {
		return fmt.Errorf("failed to open project directory: %w", err)
	}
	defer project.Close()

	// Create parent directories if needed
	dir := filepath.Dir(filePath)
	if dir != "." {
		if err := project.MkdirAll(dir, 0755); err != nil {
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

	// Create exclusively so a file appearing between validation and creation
	// is never overwritten.
	file, err := project.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("file already exists: %s", filePath)
	}
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	_, writeErr := file.Write([]byte(content))
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		_ = project.Remove(filePath)
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
		runes := []rune(word)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}
