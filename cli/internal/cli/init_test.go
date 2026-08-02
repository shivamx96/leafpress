package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitGitignoreContainsOnlyGeneratedOutput(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := runInit(nil, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	gitignore := string(data)
	if !strings.Contains(gitignore, "_site/") {
		t.Fatal("init .gitignore is missing _site/")
	}
	if strings.Contains(gitignore, ".leafpress/") {
		t.Fatal("init .gitignore contains unused .leafpress/ entry")
	}
}
