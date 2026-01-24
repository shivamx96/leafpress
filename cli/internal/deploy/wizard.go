package deploy

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Wizard handles interactive setup for deployment
type Wizard struct {
	reader *bufio.Reader
	store  *CredentialsStore
}

// NewWizard creates a new setup wizard
func NewWizard(store *CredentialsStore) *Wizard {
	return &Wizard{
		reader: bufio.NewReader(os.Stdin),
		store:  store,
	}
}

// Run executes the interactive setup wizard
func (w *Wizard) Run(ctx context.Context) (*ProviderConfig, *Credentials, error) {
	fmt.Println()
	fmt.Println("No deploy configuration found. Let's set one up!")
	fmt.Println()

	// Step 1: Select provider
	provider, err := w.selectProvider()
	if err != nil {
		return nil, nil, err
	}

	// Step 2: Authenticate
	creds, err := w.authenticate(ctx, provider)
	if err != nil {
		return nil, nil, err
	}

	// Step 3: Configure provider-specific settings
	config, err := w.configureProvider(ctx, provider, creds)
	if err != nil {
		return nil, nil, err
	}

	return config, creds, nil
}

// selectProvider shows provider selection menu
func (w *Wizard) selectProvider() (Provider, error) {
	providers := List()
	if len(providers) == 0 {
		return nil, fmt.Errorf("no deploy providers available")
	}

	fmt.Println("Select a deploy provider:")
	fmt.Println()

	for i, p := range providers {
		fmt.Printf("  %d. %s\n", i+1, p.DisplayName())
		fmt.Printf("     %s\n", p.Description())
		fmt.Println()
	}

	for {
		fmt.Print("Enter choice [1]: ")
		input, err := w.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			input = "1"
		}

		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(providers) {
			fmt.Printf("Please enter a number between 1 and %d\n", len(providers))
			continue
		}

		return providers[choice-1], nil
	}
}

// authenticate handles the auth flow for a provider
func (w *Wizard) authenticate(ctx context.Context, provider Provider) (*Credentials, error) {
	// Check if we already have valid credentials
	if existing, ok := w.store.Get(provider.Name()); ok {
		if err := provider.ValidateCredentials(ctx, existing); err == nil {
			fmt.Printf("\n  Already authenticated as %s\n", existing.Username)
			fmt.Print("  Use existing credentials? [Y/n]: ")

			input, _ := w.reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			if input == "" || input == "y" || input == "yes" {
				return existing, nil
			}
		}
	}

	// Check for environment variable
	if envCreds := GetFromEnv(provider.Name()); envCreds != nil {
		if err := provider.ValidateCredentials(ctx, envCreds); err == nil {
			fmt.Println("\n  Using credentials from environment variable")
			return envCreds, nil
		}
	}

	if !provider.NeedsAuth() {
		return &Credentials{Provider: provider.Name()}, nil
	}

	fmt.Printf("\n  %s selected. Let's authenticate.\n", provider.DisplayName())

	creds, err := provider.Authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Save credentials
	if err := w.store.Set(creds); err != nil {
		fmt.Printf("  Warning: couldn't save credentials: %v\n", err)
	}

	fmt.Printf("\n  Authenticated as %s\n", creds.Username)

	return creds, nil
}

// configureProvider sets up provider-specific settings
func (w *Wizard) configureProvider(ctx context.Context, provider Provider, creds *Credentials) (*ProviderConfig, error) {
	switch provider.Name() {
	case "github-pages":
		return w.configureGitHubPages(ctx, provider.(*GitHubPagesProvider), creds)
	case "mock":
		return &ProviderConfig{
			Provider: "mock",
			Settings: map[string]string{
				SettingRepo:   "mock-user/mock-repo",
				SettingBranch: "gh-pages",
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider.Name())
	}
}

// configureGitHubPages handles GitHub Pages specific setup
func (w *Wizard) configureGitHubPages(ctx context.Context, provider *GitHubPagesProvider, creds *Credentials) (*ProviderConfig, error) {
	fmt.Println()
	fmt.Println("  Fetching your repositories...")

	repos, err := provider.ListRepos(ctx, creds.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found - please create one first")
	}

	// Show repo selection
	fmt.Println()
	fmt.Println("  Select a repository:")
	fmt.Println()

	maxShow := 10
	if len(repos) < maxShow {
		maxShow = len(repos)
	}

	for i := 0; i < maxShow; i++ {
		visibility := "public"
		if repos[i].Private {
			visibility = "private"
		}
		fmt.Printf("    %d. %s (%s)\n", i+1, repos[i].FullName, visibility)
	}

	if len(repos) > maxShow {
		fmt.Printf("    ... and %d more\n", len(repos)-maxShow)
		fmt.Println()
		fmt.Println("  Or type a repo name (e.g., username/repo):")
	}

	var selectedRepo string
	for {
		fmt.Print("\n  Enter choice or repo name: ")
		input, err := w.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Check if it's a number
		if choice, err := strconv.Atoi(input); err == nil {
			if choice >= 1 && choice <= len(repos) {
				selectedRepo = repos[choice-1].FullName
				break
			}
			fmt.Printf("  Please enter a number between 1 and %d\n", len(repos))
			continue
		}

		// Check if it looks like a repo name
		if strings.Contains(input, "/") {
			selectedRepo = input
			break
		}

		fmt.Println("  Please enter a valid repo name (e.g., username/repo)")
	}

	// Ask for branch
	fmt.Print("\n  Deploy branch [gh-pages]: ")
	branchInput, _ := w.reader.ReadString('\n')
	branch := strings.TrimSpace(branchInput)
	if branch == "" {
		branch = "gh-pages"
	}

	fmt.Println()
	fmt.Printf("  Repository: %s\n", selectedRepo)
	fmt.Printf("  Branch: %s\n", branch)

	return &ProviderConfig{
		Provider: "github-pages",
		Settings: map[string]string{
			SettingRepo:   selectedRepo,
			SettingBranch: branch,
		},
	}, nil
}

// IsInteractive returns true if running in an interactive terminal
func IsInteractive() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
