package cli

import (
	"fmt"
	"time"

	"github.com/shivamx96/leafpress/internal/build"
	"github.com/shivamx96/leafpress/internal/config"
	"github.com/spf13/cobra"
)

var includeDrafts bool

func buildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the static site",
		Long:  `Generates static site into _site/ directory.`,
		RunE:  runBuild,
	}

	cmd.Flags().BoolVarP(&includeDrafts, "drafts", "d", false, "include draft pages")

	return cmd
}

func runBuild(cmd *cobra.Command, args []string) error {
	start := time.Now()

	// Load config
	cfg, err := config.Load(getConfigPath())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create builder
	builder := build.New(cfg, build.Options{
		IncludeDrafts: includeDrafts,
		Verbose:       isVerbose(),
	})

	// Run build
	stats, err := builder.Build()
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("Built %d pages in %s\n", stats.PageCount, elapsed.Round(time.Millisecond))

	if stats.WarningCount > 0 {
		fmt.Printf("Warnings: %d\n", stats.WarningCount)
	}

	return nil
}
