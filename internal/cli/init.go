package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shivamx96/leafpress/internal/config"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new leafpress site in the current directory",
		Long: `Scaffolds leafpress.json and optional style.css in current directory.
If no markdown files exist, creates a sample index.md.`,
		RunE: runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if config already exists
	configPath := filepath.Join(cwd, "leafpress.json")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("leafpress.json already exists. Remove it first to reinitialize")
	}

	// Create default config
	cfg := config.Default()
	if err := config.Write(configPath, cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	fmt.Println("Created leafpress.json")

	// Create style.css
	stylePath := filepath.Join(cwd, "style.css")
	if _, err := os.Stat(stylePath); os.IsNotExist(err) {
		styleContent := `/* leafpress Custom Styles
 * Override CSS variables or add custom rules below.
 * See: https://leafpress.dev/docs/theming
 *
 * Available variables:
 * --lp-font, --lp-font-mono, --lp-accent, --lp-bg, --lp-text,
 * --lp-text-muted, --lp-border, --lp-code-bg, --lp-max-width
 */
`
		if err := os.WriteFile(stylePath, []byte(styleContent), 0644); err != nil {
			return fmt.Errorf("failed to write style.css: %w", err)
		}
		fmt.Println("Created style.css")
	}

	// Create static directory
	staticDir := filepath.Join(cwd, "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if err := os.MkdirAll(staticDir, 0755); err != nil {
			return fmt.Errorf("failed to create static directory: %w", err)
		}
		// Create images subdirectory
		imagesDir := filepath.Join(staticDir, "images")
		if err := os.MkdirAll(imagesDir, 0755); err != nil {
			return fmt.Errorf("failed to create static/images directory: %w", err)
		}
		fmt.Println("Created static/images/")
	}

	// Update .gitignore
	gitignorePath := filepath.Join(cwd, ".gitignore")
	gitignoreEntries := "\n# leafpress\n_site/\n.leafpress/\n"

	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte(gitignoreEntries[1:]), 0644); err != nil {
			return fmt.Errorf("failed to write .gitignore: %w", err)
		}
		fmt.Println("Created .gitignore")
	} else {
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open .gitignore: %w", err)
		}
		defer f.Close()
		if _, err := f.WriteString(gitignoreEntries); err != nil {
			return fmt.Errorf("failed to append to .gitignore: %w", err)
		}
		fmt.Println("Updated .gitignore")
	}

	// Check if any markdown files exist, if not create index.md
	hasMarkdown := false
	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			// Skip docs directory
			rel, _ := filepath.Rel(cwd, path)
			if len(rel) > 4 && rel[:4] == "docs" {
				return nil
			}
			hasMarkdown = true
			return filepath.SkipAll
		}
		return nil
	})

	if !hasMarkdown {
		indexPath := filepath.Join(cwd, "index.md")
		indexContent := `---
title: "Welcome to My Garden"
date: ` + fmt.Sprintf("%s", "2025-01-15") + `
growth: "seedling"
---

# Welcome

This is your digital garden. Start writing!

## Getting Started

1. Edit this file or create new ones with ` + "`leafpress new <name>`" + `
2. Use [[wiki-links]] to connect your thoughts
3. Run ` + "`leafpress serve`" + ` to preview your garden
4. Run ` + "`leafpress build`" + ` to generate your static site
`
		if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
			return fmt.Errorf("failed to write index.md: %w", err)
		}
		fmt.Println("Created index.md")
	}

	fmt.Println("\nleafpress initialized! Run 'leafpress serve' to start the dev server.")
	return nil
}
