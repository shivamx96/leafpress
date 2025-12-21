package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

func Execute(version string) error {
	rootCmd := &cobra.Command{
		Use:   "leafpress",
		Short: "A CLI-driven static site generator for digital gardens",
		Long: `LeafPress transforms a folder of Markdown files into a clean,
interlinked website with minimal configuration.

Your garden folder IS the product. LeafPress is invisible infrastructure.`,
		Version: version,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ./leafpress.json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(buildCmd())
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(newCmd())

	// Custom version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("LeafPress %s\n", version))

	return rootCmd.Execute()
}

func getConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}
	return "leafpress.json"
}

func isVerbose() bool {
	return verbose
}
