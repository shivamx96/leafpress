package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shivamx96/leafpress/cli/internal/deploy"
	"github.com/shivamx96/leafpress/core/config"
)

func TestApplyGitHubPagesBaseURL(t *testing.T) {
	t.Run("sets nested site.baseURL when absent", func(t *testing.T) {
		raw := map[string]interface{}{}
		got, set := applyGitHubPagesBaseURL(raw, "user/repo")
		if !set || got != "https://user.github.io/repo" {
			t.Fatalf("got (%q, %v), want (https://user.github.io/repo, true)", got, set)
		}
		site, ok := raw["site"].(map[string]interface{})
		if !ok || site["baseURL"] != "https://user.github.io/repo" {
			t.Fatalf("site.baseURL not written: %#v", raw)
		}
		// It must NOT leak a top-level baseURL (the v2 build ignores that).
		if _, exists := raw["baseURL"]; exists {
			t.Fatal("must not write a top-level baseURL")
		}
	})

	t.Run("preserves existing site.baseURL", func(t *testing.T) {
		raw := map[string]interface{}{"site": map[string]interface{}{"baseURL": "https://custom.example"}}
		if got, set := applyGitHubPagesBaseURL(raw, "user/repo"); set || got != "" {
			t.Fatalf("should not overwrite existing baseURL, got (%q, %v)", got, set)
		}
	})

	t.Run("adds baseURL into an existing site section", func(t *testing.T) {
		raw := map[string]interface{}{"site": map[string]interface{}{"title": "T"}}
		if _, set := applyGitHubPagesBaseURL(raw, "user/repo"); !set {
			t.Fatal("expected baseURL to be set")
		}
		site := raw["site"].(map[string]interface{})
		if site["title"] != "T" || site["baseURL"] != "https://user.github.io/repo" {
			t.Fatalf("existing site keys not preserved: %#v", site)
		}
	})

	t.Run("user pages repo has no subpath", func(t *testing.T) {
		raw := map[string]interface{}{}
		if got, _ := applyGitHubPagesBaseURL(raw, "user/user.github.io"); got != "https://user.github.io" {
			t.Fatalf("got %q, want https://user.github.io", got)
		}
	})
}

func TestSaveDeployConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	original := []byte(`{
  "site": {"title": "Garden"},
  "theme": {"accent": "#123456"}
}`)
	if err := os.WriteFile(filepath.Join(dir, "leafpress.json"), original, 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	deployConfig := &deploy.ProviderConfig{
		Provider: "github-pages",
		Settings: map[string]string{
			deploy.SettingRepo:   "example/garden",
			deploy.SettingBranch: "gh-pages",
		},
	}
	if err := saveDeployConfig(cfg, deployConfig); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load("leafpress.json")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Site.Title != "Garden" || loaded.Site.BaseURL != "https://example.github.io/garden" {
		t.Fatalf("site config = %+v", loaded.Site)
	}
	if loaded.Theme.Accent != "#123456" || loaded.Deploy.Provider != "github-pages" || loaded.Deploy.Settings[deploy.SettingBranch] != "gh-pages" {
		t.Fatalf("round-tripped config = %+v", loaded)
	}

	var raw map[string]any
	data, err := os.ReadFile("leafpress.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, exists := raw["baseURL"]; exists {
		t.Fatal("saveDeployConfig wrote obsolete top-level baseURL")
	}
}
