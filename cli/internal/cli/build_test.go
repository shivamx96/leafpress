package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStrictBuildPreservesPublishedOutput(t *testing.T) {
	t.Chdir(t.TempDir())
	withConfigFlag(t, "leafpress.json")
	if err := os.WriteFile("index.md", []byte("# Published\n"), 0644); err != nil {
		t.Fatal(err)
	}
	build := func(args ...string) error {
		cmd := buildCmd()
		cmd.SetArgs(append([]string{}, args...))
		return cmd.Execute()
	}
	if err := build("--strict"); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join("_site", "index.html")
	before, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("index.md", []byte("# Changed\n\n[[missing-page]]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := build("--strict"); err == nil || !strings.Contains(err.Error(), "strict build failed") {
		t.Fatalf("strict build must reject broken links: %v", err)
	}
	after, err := os.ReadFile(output)
	if err != nil || string(after) != string(before) {
		t.Fatalf("strict failure replaced output: %v", err)
	}
	if err := build(); err != nil {
		t.Fatalf("normal build should allow warnings: %v", err)
	}
	after, err = os.ReadFile(output)
	if err != nil || !strings.Contains(string(after), "Changed") {
		t.Fatalf("normal build did not publish changed content: %v", err)
	}
}

func TestBuildRejectsTrailingConfigBeforePublishing(t *testing.T) {
	t.Chdir(t.TempDir())
	withConfigFlag(t, "custom.json")
	if err := os.WriteFile("custom.json", []byte("{} trailing"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := buildCmd()
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "custom.json") {
		t.Fatalf("expected a config error naming the file: %v", err)
	}
	if _, err := os.Stat("_site"); !os.IsNotExist(err) {
		t.Fatalf("invalid config published output: %v", err)
	}
}
