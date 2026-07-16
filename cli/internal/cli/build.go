package cli

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/shivamx96/leafpress/cli/internal/build"
	"github.com/shivamx96/leafpress/core/config"
	"github.com/spf13/cobra"
)

var includeDrafts bool
var cpuProfile string
var memProfile string

func buildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the static site",
		Long:  `Generates static site into _site/ directory.`,
		RunE:  runBuild,
	}

	cmd.Flags().BoolVarP(&includeDrafts, "drafts", "d", false, "include draft pages")
	cmd.Flags().StringVar(&cpuProfile, "cpuprofile", "", "write CPU profile to file")
	cmd.Flags().StringVar(&memProfile, "memprofile", "", "write memory profile to file")

	return cmd
}

func runBuild(cmd *cobra.Command, args []string) error {
	// Start CPU profiling
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			return fmt.Errorf("could not create CPU profile: %w", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return fmt.Errorf("could not start CPU profile: %w", err)
		}
		defer pprof.StopCPUProfile()
	}

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

	// Write memory profile
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			return fmt.Errorf("could not create memory profile: %w", err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			return fmt.Errorf("could not write memory profile: %w", err)
		}
	}

	return nil
}
