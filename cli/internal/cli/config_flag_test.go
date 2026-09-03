package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/cli/internal/deploy"
	"github.com/shivamx96/leafpress/core/config"
)

var deployProviderFixture = deploy.ProviderConfig{
	Provider: "netlify",
	Settings: map[string]string{deploy.SettingSiteID: "abc123"},
}

// withConfigFlag sets the persistent --config value for one test.
func withConfigFlag(t *testing.T, value string) {
	t.Helper()
	previous := cfgFile
	cfgFile = value
	t.Cleanup(func() { cfgFile = previous })
}

// The wizard writes the provider back. Writing it to leafpress.json when
// --config named another file would strand the settings in a file nothing
// reads.
func TestSaveDeployConfigWritesFlaggedFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := config.Write(filepath.Join(dir, "custom.json"), config.Default()); err != nil {
		t.Fatal(err)
	}
	withConfigFlag(t, "custom.json")

	cfg, err := config.Load(getConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := saveDeployConfig(cfg, &deployProviderFixture); err != nil {
		t.Fatal(err)
	}

	written, err := os.ReadFile(filepath.Join(dir, "custom.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "netlify") {
		t.Errorf("deploy settings were not saved to custom.json:\n%s", written)
	}
	if _, err := os.Stat(filepath.Join(dir, "leafpress.json")); err == nil {
		t.Error("saveDeployConfig created leafpress.json instead of the flagged file")
	}
}

func TestInitScaffoldsFlaggedConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	withConfigFlag(t, "custom.json")

	if err := runInit(nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "custom.json")); err != nil {
		t.Fatalf("init should scaffold the flagged config: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "leafpress.json")); err == nil {
		t.Error("init wrote leafpress.json despite --config custom.json")
	}

	// Re-running must refuse by the flagged name, not a hardcoded one.
	err := runInit(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "custom.json") {
		t.Fatalf("re-init should refuse naming custom.json, got: %v", err)
	}
}
