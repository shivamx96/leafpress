package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withConfigFlag sets the persistent --config value for one test.
func withConfigFlag(t *testing.T, value string) {
	t.Helper()
	previous := cfgFile
	cfgFile = value
	t.Cleanup(func() { cfgFile = previous })
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

func TestIsOfflineReadsFlagAndEnvironment(t *testing.T) {
	previous := offline
	t.Cleanup(func() { offline = previous })

	offline = false
	for value, want := range map[string]bool{"": false, "0": false, "false": false, "FALSE": false, "1": true, "true": true, "yes": true} {
		t.Setenv("LEAFPRESS_OFFLINE", value)
		if got := isOffline(); got != want {
			t.Errorf("LEAFPRESS_OFFLINE=%q: isOffline() = %v, want %v", value, got, want)
		}
	}
	t.Setenv("LEAFPRESS_OFFLINE", "")
	offline = true
	if !isOffline() {
		t.Error("--offline should disable downloads")
	}
}

func TestLoadConfigRejectsMissingFlaggedFile(t *testing.T) {
	t.Chdir(t.TempDir())
	withConfigFlag(t, "leafpres.json")

	_, err := loadConfig()
	if err == nil || !strings.Contains(err.Error(), "leafpres.json") {
		t.Fatalf("missing --config file should fail naming it, got: %v", err)
	}
}

func TestLoadConfigDefaultsWhenUnflaggedFileMissing(t *testing.T) {
	t.Chdir(t.TempDir())
	withConfigFlag(t, "")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("missing default leafpress.json should use defaults: %v", err)
	}
	if cfg.Build.OutputDir != "_site" {
		t.Errorf("OutputDir = %q, want default _site", cfg.Build.OutputDir)
	}
}

func TestLoadConfigReadsFlaggedFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	withConfigFlag(t, "custom.json")
	if err := os.WriteFile(filepath.Join(dir, "custom.json"),
		[]byte(`{"build":{"outputDir":"out"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Build.OutputDir != "out" {
		t.Errorf("OutputDir = %q, want out", cfg.Build.OutputDir)
	}
}
