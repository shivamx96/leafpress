package cli

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shivamx96/leafpress/cli/internal/config"
	"github.com/shivamx96/leafpress/cli/internal/deploy"
	"github.com/spf13/cobra"
)

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show deployment status and pending changes",
		Long: `Show what files have changed since the last deployment.

This helps you see which files need to be deployed.

Examples:
  leafpress status              # Show pending files
  leafpress status --verbose    # Show detailed information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}

	return cmd
}

func runStatus() error {
	// Load config
	cfg, err := config.Load("leafpress.json")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if deployment is configured
	if cfg.Deploy.Provider == "" {
		fmt.Println("Deployment Status")
		fmt.Println("=================")
		fmt.Println()
		fmt.Println("No deployment configured yet.")
		fmt.Println("Run 'leafpress deploy' to set up deployment.")
		return nil
	}

	// Load deployment manifest
	manifest, err := deploy.LoadDeploymentManifest(".")
	if err != nil {
		return fmt.Errorf("failed to load deployment manifest: %w", err)
	}

	// Get build directory
	buildDir := cfg.OutputDir
	if buildDir == "" {
		buildDir = "_site"
	}

	// Check if build directory exists
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		return fmt.Errorf("build directory '%s' not found - run 'leafpress build' first", buildDir)
	}

	// Collect current source files with hashes (respecting ignore patterns and excluding output dir)
	currentFiles, err := CollectSourceFilesWithHashes(buildDir, cfg.Ignore)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	// Get pending files
	pendingFiles := manifest.GetPendingFiles(currentFiles)

	// Print status
	fmt.Println("Deployment Status")
	fmt.Println("=================")
	fmt.Println()

	if manifest.LastDeploy == nil {
		fmt.Println("Provider: ", cfg.Deploy.Provider)
		fmt.Println("Status:   Never deployed")
		fmt.Println()
		fmt.Printf("Ready to deploy %d files.\n", len(currentFiles))
		fmt.Println()
		fmt.Println("Run 'leafpress deploy' to deploy.")
	} else {
		fmt.Println("Provider:     ", cfg.Deploy.Provider)
		fmt.Println("Last Deploy:  ", manifest.TimeSinceLastDeploy())
		fmt.Println("Live URL:     ", manifest.LastDeploy.URL)
		fmt.Println("Deployed:     ", manifest.LastDeploy.FileCount, "files")
		fmt.Println()

		if len(pendingFiles) == 0 {
			fmt.Println("✓ Everything is deployed!")
		} else {
			fmt.Printf("⚠ %d file(s) pending deployment:\n", len(pendingFiles))
			fmt.Println()

			// Sort pending files for consistent output
			var pendingPaths []string
			for path := range pendingFiles {
				pendingPaths = append(pendingPaths, path)
			}
			sort.Strings(pendingPaths)

			for _, path := range pendingPaths {
				status := "modified"
				if pendingFiles[path] == "deleted" {
					status = "deleted"
				} else if _, exists := manifest.LastDeploy.FilesDeployed[path]; !exists {
					status = "new"
				}

				fmt.Printf("  %s %s\n", statusIcon(status), path)
			}

			fmt.Println()
			fmt.Println("Run 'leafpress deploy' to deploy these changes.")
		}
	}

	fmt.Println()
	return nil
}

// CollectSourceFilesWithHashes walks the source directory and returns file paths with SHA1 hashes
// Tracks all source files (notes, config, assets) that have changed
// Excludes the output directory (generated files), ignored directories, .obsidian, and system/metadata files
func CollectSourceFilesWithHashes(buildDir string, ignorePatterns []string) (map[string]string, error) {
	files := make(map[string]string)

	// Convert buildDir to absolute path for comparison
	absBuildDir, err := filepath.Abs(buildDir)
	if err != nil {
		return nil, err
	}

	// System directories to always skip
	skipDirs := map[string]bool{
		".obsidian":    true,
		".git":         true,
		"node_modules": true,
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			absPath, _ := filepath.Abs(path)

			// Skip output directory
			if absPath == absBuildDir {
				return filepath.SkipDir
			}

			// Skip system directories
			dirName := filepath.Base(path)
			if skipDirs[dirName] {
				return filepath.SkipDir
			}

			// Skip ignored directories
			for _, pattern := range ignorePatterns {
				if dirName == pattern {
					return filepath.SkipDir
				}
			}
			return nil
		}

		filename := filepath.Base(path)

		// Skip metadata/system files
		skipFiles := []string{
			deploy.ManifestFile, // .leafpress-deploy-state.json
			".DS_Store",         // macOS
			"Thumbs.db",         // Windows
		}
		for _, pattern := range skipFiles {
			if filename == pattern {
				return nil
			}
		}

		// Calculate SHA1
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hash := sha1.Sum(data)
		hashStr := hex.EncodeToString(hash[:])

		// Use forward slashes for consistency, remove leading ./
		relPath := filepath.ToSlash(path)
		if strings.HasPrefix(relPath, "./") {
			relPath = relPath[2:]
		}
		files["/"+relPath] = hashStr

		return nil
	})

	return files, err
}

// statusIcon returns a visual indicator for file status
func statusIcon(status string) string {
	switch status {
	case "new":
		return "+"
	case "modified":
		return "~"
	case "deleted":
		return "-"
	default:
		return " "
	}
}
