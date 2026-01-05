package cli

import (
	"fmt"
	"time"

	"github.com/shivamx96/leafpress/cli/internal/build"
	"github.com/shivamx96/leafpress/cli/internal/config"
	"github.com/shivamx96/leafpress/cli/internal/server"
	"github.com/spf13/cobra"
)

var servePort int

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start development server with live reload",
		Long:  `Starts a local development server with live reload on file changes.`,
		RunE:  runServe,
	}

	cmd.Flags().IntVarP(&servePort, "port", "p", 0, "override server port")
	cmd.Flags().BoolVarP(&includeDrafts, "drafts", "d", false, "include draft pages")

	return cmd
}

func runServe(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(getConfigPath())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override port if specified
	if servePort > 0 {
		cfg.Port = servePort
	}

	// Create builder
	builder := build.New(cfg, build.Options{
		IncludeDrafts: includeDrafts,
		Verbose:       isVerbose(),
	})

	// Initial build
	fmt.Println("Building site...")
	start := time.Now()
	stats, err := builder.Build()
	if err != nil {
		return fmt.Errorf("initial build failed: %w", err)
	}
	elapsed := time.Since(start)
	fmt.Printf("Built %d pages in %s\n", stats.PageCount, elapsed.Round(time.Millisecond))

	// Start server
	srv := server.New(cfg, builder, server.Options{
		Verbose: isVerbose(),
	})

	return srv.Start()
}
