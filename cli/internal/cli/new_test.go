package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNewCreatesNestedPage(t *testing.T) {
	project := t.TempDir()
	t.Chdir(project)

	if err := runNew(nil, []string{"Field Notes/Éclair Ideas"}); err != nil {
		t.Fatal(err)
	}
	pagePath := filepath.Join(project, "field-notes", "éclair-ideas.md")
	data, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `title: "Éclair Ideas"`) {
		t.Fatalf("generated page has incorrect Unicode title:\n%s", data)
	}
}

func TestRunNewRejectsAbsoluteDestination(t *testing.T) {
	project := t.TempDir()
	t.Chdir(project)
	outPath := filepath.Join(t.TempDir(), "outside-page.md")

	err := runNew(nil, []string{strings.TrimSuffix(outPath, ".md")})
	if err == nil || !strings.Contains(err.Error(), "must stay inside the project") {
		t.Fatalf("runNew error = %v, want containment error", err)
	}
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		t.Fatalf("page was created outside project: %v", err)
	}
}

func TestRunNewRejectsSymlinkEscape(t *testing.T) {
	project := t.TempDir()
	outside := t.TempDir()
	t.Chdir(project)
	if err := os.Symlink(outside, filepath.Join(project, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	err := runNew(nil, []string{"linked/page"})
	if err == nil {
		t.Fatal("runNew should reject a symlink that leaves the project")
	}
	if _, err := os.Stat(filepath.Join(outside, "page.md")); !os.IsNotExist(err) {
		t.Fatalf("page was created through escaping symlink: %v", err)
	}
}

func TestRunNewDoesNotOverwriteExistingPage(t *testing.T) {
	project := t.TempDir()
	t.Chdir(project)
	pagePath := filepath.Join(project, "existing.md")
	if err := os.WriteFile(pagePath, []byte("keep me"), 0644); err != nil {
		t.Fatal(err)
	}

	err := runNew(nil, []string{"existing"})
	if err == nil || !strings.Contains(err.Error(), "file already exists") {
		t.Fatalf("runNew error = %v, want existing-file error", err)
	}
	data, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "keep me" {
		t.Fatalf("existing page was changed: %q", data)
	}
}

func TestGenerateTitleHandlesUnicode(t *testing.T) {
	if got := generateTitle("éclair-notes"); got != "Éclair Notes" {
		t.Fatalf("generateTitle = %q, want %q", got, "Éclair Notes")
	}
}
