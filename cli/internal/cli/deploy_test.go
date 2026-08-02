package cli

import "testing"

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
