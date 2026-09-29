package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shivamx96/leafpress/core/config"
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

func TestInitWritesTidyConfigAndDatedIndex(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := runInit(nil, nil); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "leafpress.json"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := string(data)
	for _, unwanted := range []string{`"deploy"`, "null"} {
		if strings.Contains(cfg, unwanted) {
			t.Errorf("generated leafpress.json contains %s:\n%s", unwanted, cfg)
		}
	}
	if !strings.HasSuffix(cfg, "}\n") {
		t.Error("generated leafpress.json should end with a newline")
	}
	if _, err := config.Load(filepath.Join(dir, "leafpress.json")); err != nil {
		t.Errorf("generated leafpress.json does not load: %v", err)
	}

	index, err := os.ReadFile(filepath.Join(dir, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "date: " + time.Now().Format("2006-01-02") + "\n"; !strings.Contains(string(index), want) {
		t.Errorf("index.md should be dated today (%q):\n%s", want, index)
	}
}

func TestInitDoesNotDuplicateGitignoreEntry(t *testing.T) {
	for name, existing := range map[string]string{
		"trailing slash": "node_modules/\n_site/\n",
		"rooted":         "/_site\n",
		"crlf":           "_site/\r\n",
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			path := filepath.Join(dir, ".gitignore")
			if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
				t.Fatal(err)
			}
			if err := runInit(nil, nil); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != existing {
				t.Errorf(".gitignore changed although it already ignores _site:\n%q", data)
			}
		})
	}

	dir := t.TempDir()
	t.Chdir(dir)
	path := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(path, []byte("node_modules/\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runInit(nil, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "_site/") {
		t.Errorf("init should append _site/ to an existing .gitignore:\n%s", data)
	}
}
