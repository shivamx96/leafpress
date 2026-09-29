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
		Long: `leafpress transforms a folder of Markdown files into a clean,
interlinked website with minimal configuration.

Your garden folder IS the product. leafpress is invisible infrastructure.`,
		Version: version,
		// Cobra validates flags and arguments before this hook, so usage
		// mistakes still print help. Failures after that point, such as a bad
		// config or a strict-build warning, print only the error.
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ./leafpress.json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(buildCmd())
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(newCmd())
	rootCmd.AddCommand(versionCmd(version))
	rootCmd.AddCommand(updateCmd(version))

	// Custom version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("leafpress %s\n", version))

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
