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
