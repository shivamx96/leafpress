package deploy

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCredentialsStoreRoundTripAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "credentials.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	// Existing files must also have their permissions tightened.
	if err := os.WriteFile(path, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}
	store, err := NewCredentialsStoreAt(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set(&Credentials{Provider: "netlify", AccessToken: "secret", Username: "user"}); err != nil {
		t.Fatal(err)
	}

	reloaded, err := NewCredentialsStoreAt(path)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := reloaded.Get("netlify")
	if !ok || got.AccessToken != "secret" || got.Username != "user" {
		t.Fatalf("round trip = %+v, %v", got, ok)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Fatalf("credential mode = %o, want 600", info.Mode().Perm())
		}
	}
}

func TestDeploymentManifestPendingFiles(t *testing.T) {
	m := NewDeploymentManifest()
	m.RecordDeployment(
		&DeployResult{DeployID: "one", URL: "https://example.com", DeployedAt: time.Now()},
		"netlify",
		map[string]string{"/site.html": "built"},
		map[string]string{"note.md": "old", "deleted.md": "gone"},
	)
	pending := m.GetPendingFiles(map[string]string{"note.md": "new", "added.md": "added"})
	for path, want := range map[string]string{
		"note.md":    "new",
		"added.md":   "added",
		"deleted.md": "deleted",
	} {
		if pending[path] != want {
			t.Errorf("pending[%q] = %q, want %q", path, pending[path], want)
		}
	}
	if len(pending) != 3 {
		t.Fatalf("pending = %v", pending)
	}

	dir := t.TempDir()
	if err := m.Save(dir); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadDeploymentManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.LastDeploy == nil || loaded.LastDeploy.DeployID != "one" {
		t.Fatalf("loaded manifest = %+v", loaded)
	}
}

func TestGitCommitUsesRepositoryIndependentIdentity(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	runGit := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return strings.TrimSpace(string(out))
	}
	runGit("init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("site"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "index.html")
	hasChanges, err := hasStagedChanges(context.Background(), dir)
	if err != nil || !hasChanges {
		t.Fatalf("hasStagedChanges = %v, %v", hasChanges, err)
	}
	if output, err := gitCommitCommand(context.Background(), dir, "deploy").CombinedOutput(); err != nil {
		t.Fatalf("commit without global identity: %v\n%s", err, output)
	}
	if got := runGit("log", "-1", "--format=%an <%ae>"); got != "Leafpress Deploy <leafpress@users.noreply.github.com>" {
		t.Fatalf("commit identity = %q", got)
	}
	hasChanges, err = hasStagedChanges(context.Background(), dir)
	if err != nil || hasChanges {
		t.Fatalf("clean hasStagedChanges = %v, %v", hasChanges, err)
	}
}
