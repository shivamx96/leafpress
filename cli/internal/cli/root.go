package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/shivamx96/leafpress/core/config"
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

// loadConfig loads the selected config file. Without --config, a missing
// leafpress.json means "use defaults". An explicit --config naming a file
// that does not exist is an error: falling back to defaults there would
// silently drop build.ignore and outputDir, publishing notes the author
// meant to keep out of the site.
func loadConfig() (*config.Config, error) {
	path := getConfigPath()
	if cfgFile != "" {
		if _, err := os.Stat(path); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("config file %q does not exist", path)
			}
			return nil, fmt.Errorf("config file %q: %w", path, err)
		}
	}
	return config.Load(path)
}

func isVerbose() bool {
	return verbose
}

// offline is set by --offline on build and serve.
var offline bool

// isOffline reports whether font downloads are disabled, by flag or by a
// LEAFPRESS_OFFLINE value other than "", "0", or "false".
func isOffline() bool {
	if offline {
		return true
	}
	switch strings.ToLower(os.Getenv("LEAFPRESS_OFFLINE")) {
	case "", "0", "false":
		return false
	}
	return true
}
