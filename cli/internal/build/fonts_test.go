package build

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/cli/internal/fonts"
	"github.com/shivamx96/leafpress/cli/internal/fonts/fontstest"
	"github.com/shivamx96/leafpress/core/config"
)

// TestMain stops every test in this package from reaching Google Fonts. A
// test that needs downloads passes a fontstest client explicitly.
func TestMain(m *testing.M) {
	defaultFontClient = func() *fonts.Client {
		client := fonts.Google()
		client.HTTP = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network access is disabled in tests")
		})}
		return client
	}
	os.Exit(m.Run())
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBuildDownloadsAndSelfHostsAGoogleFont(t *testing.T) {
	dir := newTestProject(t)
	server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true, Italic: true})

	cfg := config.Default()
	cfg.Theme.FontHeading = "Test Serif"
	stats, err := New(cfg, Options{FontClient: server.Client()}).Build()
	if err != nil {
		t.Fatal(err)
	}
	if stats.WarningCount != 0 {
		t.Errorf("a downloaded family should not warn, got %d warnings", stats.WarningCount)
	}

	css, err := os.ReadFile(filepath.Join(dir, "_site", "style.css"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`font-family: "Test Serif"`,
		`src: url("static/fonts/test-serif/test-serif-normal-latin.woff2") format("woff2");`,
		"unicode-range: " + fontstest.LatinRange + ";",
	} {
		if !strings.Contains(string(css), want) {
			t.Errorf("style.css missing %q", want)
		}
	}
	page, err := os.ReadFile(filepath.Join(dir, "_site", "note", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), `href="/static/fonts/test-serif/test-serif-normal-latin.woff2" as="font"`) {
		t.Error("the heading font should preload its Latin file")
	}
	if strings.Contains(string(page), "fonts.googleapis.com") {
		t.Error("a downloaded family must not load from Google at read time")
	}
	for _, published := range []string{
		"static/fonts/test-serif/test-serif-italic-latin-ext.woff2",
		"static/fonts/test-serif/OFL.txt",
	} {
		if _, err := os.Stat(filepath.Join(dir, "_site", published)); err != nil {
			t.Errorf("%s not published: %v", published, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "_site", fonts.LockFile)); !os.IsNotExist(err) {
		t.Error("the font lock is build input and must not be published")
	}

	// Rebuilding reads the lock: no requests, identical declarations, and
	// the downloaded faces are not added to the theme twice.
	requests := server.Requests.Load()
	builder := New(cfg, Options{FontClient: server.Client()})
	for range 2 {
		if _, err := builder.Build(); err != nil {
			t.Fatal(err)
		}
	}
	if server.Requests.Load() != requests {
		t.Error("rebuilds must not download a family that is already in the lock")
	}
	if n := len(builder.cfg.Theme.Fonts); n != 4 {
		t.Errorf("theme has %d font faces after repeated builds, want 4", n)
	}
}

func TestStrictBuildNeverDownloadsFonts(t *testing.T) {
	dir := newTestProject(t)
	server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true})

	cfg := config.Default()
	cfg.Theme.FontBody = "Test Serif"
	_, err := New(cfg, Options{Strict: true, FontClient: server.Client()}).Build()
	if err == nil || !strings.Contains(err.Error(), "strict build failed") {
		t.Fatalf("a strict build with an undownloaded font must fail, got %v", err)
	}
	if server.Requests.Load() != 0 {
		t.Error("strict builds must not download fonts")
	}
	if _, err := os.Stat(filepath.Join(dir, "static", "fonts")); !os.IsNotExist(err) {
		t.Error("strict builds must not write fonts into the garden")
	}

	// Once downloaded by a normal build, the same strict build passes.
	if _, err := New(cfg, Options{FontClient: server.Client()}).Build(); err != nil {
		t.Fatal(err)
	}
	if _, err := New(cfg, Options{Strict: true, FontClient: server.Client()}).Build(); err != nil {
		t.Fatalf("strict build with a downloaded font: %v", err)
	}
}

func TestOfflineBuildFallsBackWithAReason(t *testing.T) {
	newTestProject(t)
	server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true})

	cfg := config.Default()
	cfg.Theme.FontBody = "Test Serif"
	stats, err := New(cfg, Options{Offline: true, FontClient: server.Client()}).Build()
	if err != nil {
		t.Fatal(err)
	}
	if stats.WarningCount != 1 || server.Requests.Load() != 0 {
		t.Errorf("offline build: %d warnings, %d requests; want 1 and 0", stats.WarningCount, server.Requests.Load())
	}
}

// Changing the font during `leafpress serve` reloads the config and must
// download the new family on that rebuild.
func TestConfigReloadDownloadsNewFont(t *testing.T) {
	dir := newTestProject(t)
	server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true})

	cfg := config.Default()
	if err := config.Write(filepath.Join(dir, "leafpress.json"), cfg); err != nil {
		t.Fatal(err)
	}
	builder := New(cfg, Options{FontClient: server.Client()})
	if _, err := builder.Build(); err != nil {
		t.Fatal(err)
	}

	cfg.Theme.FontBody = "Test Serif"
	if err := config.Write(filepath.Join(dir, "leafpress.json"), cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := builder.RebuildIncremental(filepath.Join(dir, "leafpress.json"), ChangeModify); err != nil {
		t.Fatal(err)
	}
	css, err := os.ReadFile(filepath.Join(dir, "_site", "style.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(css), `font-family: "Test Serif"`) {
		t.Error("the reloaded config's font should be downloaded and self-hosted")
	}
}
