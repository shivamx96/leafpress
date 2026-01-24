package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/shivamx96/leafpress/cli/internal/build"
	"github.com/shivamx96/leafpress/cli/internal/config"
	"github.com/shivamx96/leafpress/cli/internal/deploy"
	"github.com/spf13/cobra"
)

func deployCmd() *cobra.Command {
	var (
		providerFlag string
		skipBuild    bool
		reconfigure  bool
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy your site to a hosting provider",
		Long: `Deploy your leafpress site to GitHub Pages, Netlify, or Vercel.

First-time setup will guide you through authentication and configuration.
Subsequent deploys are a single command.

Examples:
  leafpress deploy              # Deploy using saved configuration
  leafpress deploy --dry-run    # Validate without deploying
  leafpress deploy --skip-build # Deploy existing build
  leafpress deploy --reconfigure # Re-run setup wizard`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(providerFlag, skipBuild, reconfigure, dryRun)
		},
	}

	cmd.Flags().StringVar(&providerFlag, "provider", "", "Deploy provider (github-pages|netlify|vercel)")
	cmd.Flags().BoolVar(&skipBuild, "skip-build", false, "Deploy existing build without rebuilding")
	cmd.Flags().BoolVar(&reconfigure, "reconfigure", false, "Re-run provider setup wizard")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Build and validate without deploying")

	return cmd
}

func runDeploy(providerFlag string, skipBuild, reconfigure, dryRun bool) error {
	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt - goroutine exits when context is cancelled
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer signal.Stop(sigChan) // Stop receiving signals on function exit
	go func() {
		select {
		case <-sigChan:
			fmt.Println("\nCancelled")
			cancel()
		case <-ctx.Done():
			// Context cancelled, exit goroutine cleanly
		}
	}()

	// Load config
	cfg, err := config.Load("leafpress.json")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize credentials store
	store, err := deploy.NewCredentialsStore()
	if err != nil {
		return fmt.Errorf("failed to initialize credentials store: %w", err)
	}

	// Determine provider and config
	var providerConfig *deploy.ProviderConfig
	var creds *deploy.Credentials

	needsSetup := reconfigure || cfg.Deploy.Provider == ""

	// Check if we need to run the wizard
	if needsSetup {
		if !deploy.IsInteractive() {
			return fmt.Errorf("no deploy configuration found and running in non-interactive mode\n" +
				"Run 'leafpress deploy' interactively first, or set LEAFPRESS_GITHUB_TOKEN")
		}

		wizard := deploy.NewWizard(store)
		providerConfig, creds, err = wizard.Run(ctx)
		if err != nil {
			return err
		}

		// Save config to leafpress.json
		if err := saveDeployConfig(cfg, providerConfig); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Println()
		fmt.Println("  Configuration saved to leafpress.json")
	} else {
		// Use existing config
		providerConfig = &deploy.ProviderConfig{
			Provider: cfg.Deploy.Provider,
			Settings: cfg.Deploy.Settings,
		}

		// Get credentials
		if envCreds := deploy.GetFromEnv(cfg.Deploy.Provider); envCreds != nil {
			creds = envCreds
		} else if storedCreds, ok := store.Get(cfg.Deploy.Provider); ok {
			creds = storedCreds
		} else {
			return fmt.Errorf("no credentials found for %s\n"+
				"Run 'leafpress deploy --reconfigure' to set up authentication", cfg.Deploy.Provider)
		}
	}

	// Override provider if flag is set
	if providerFlag != "" {
		providerConfig.Provider = providerFlag
	}

	// Get provider
	provider, ok := deploy.Get(providerConfig.Provider)
	if !ok {
		return fmt.Errorf("unknown provider: %s", providerConfig.Provider)
	}

	// Validate credentials
	if provider.NeedsAuth() {
		if err := provider.ValidateCredentials(ctx, creds); err != nil {
			return fmt.Errorf("invalid credentials: %w\nRun 'leafpress deploy --reconfigure' to re-authenticate", err)
		}
	}

	// Build site (unless skipped)
	if !skipBuild {
		fmt.Println()
		fmt.Println("Building site...")
		start := time.Now()

		builder := build.New(cfg, build.Options{})
		stats, err := builder.Build()
		if err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		fmt.Printf("  Built %d pages in %s\n", stats.PageCount, time.Since(start).Round(time.Millisecond))
	}

	// Check build directory exists
	buildDir := cfg.OutputDir
	if buildDir == "" {
		buildDir = "_site"
	}

	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		return fmt.Errorf("build directory '%s' not found - run 'leafpress build' first", buildDir)
	}

	// Deploy
	fmt.Println()
	if dryRun {
		fmt.Println("Validating deployment (dry run)...")
	} else {
		fmt.Printf("Deploying to %s...\n", provider.DisplayName())
	}

	deployCtx := &deploy.DeployContext{
		BuildDir: buildDir,
		Config:   providerConfig,
		Creds:    creds,
		DryRun:   dryRun,
	}

	result, err := provider.Deploy(ctx, deployCtx)
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	fmt.Println()
	if dryRun {
		fmt.Printf("  Dry run complete. Would deploy to: %s\n", result.URL)
	} else {
		fmt.Printf("  Deployed! Live at %s\n", result.URL)
	}

	return nil
}

// saveDeployConfig updates leafpress.json with deploy configuration
func saveDeployConfig(cfg *config.Config, deployConfig *deploy.ProviderConfig) error {
	cfg.Deploy = config.DeployConfig{
		Provider: deployConfig.Provider,
		Settings: deployConfig.Settings,
	}

	// Read existing file to preserve formatting
	data, err := os.ReadFile("leafpress.json")
	if err != nil {
		return err
	}

	// Parse into map to preserve unknown fields
	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return err
	}

	// Update deploy section
	rawConfig["deploy"] = map[string]interface{}{
		"provider": deployConfig.Provider,
		"settings": deployConfig.Settings,
	}

	// Set baseURL for GitHub Pages if not already set
	if deployConfig.Provider == "github-pages" {
		if currentBaseURL, _ := rawConfig["baseURL"].(string); currentBaseURL == "" {
			baseURL := deploy.BuildGitHubPagesURL(deployConfig.Settings[deploy.SettingRepo])
			if baseURL != "" {
				rawConfig["baseURL"] = baseURL
				fmt.Printf("  Setting baseURL to %s\n", baseURL)
			}
		}
	}

	// Write back with indentation
	newData, err := json.MarshalIndent(rawConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("leafpress.json", newData, 0644)
}
