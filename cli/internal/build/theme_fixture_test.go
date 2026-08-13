package build

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/config"
)

// TestThemeGardenFixtureBuildsThemeSurfaces keeps the checked-in visual
// fixture honest. Theme authors use the garden for manual review; this test
// proves it still renders the component states and generated artifacts it
// promises to cover.
func TestThemeGardenFixtureBuildsThemeSurfaces(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve theme fixture test path")
	}
	fixtureDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "testdata", "theme-garden")
	projectDir := filepath.Join(t.TempDir(), "theme-garden")
	copyThemeGardenFixture(t, fixtureDir, projectDir)

	t.Chdir(projectDir)
	cfg, err := config.Load("leafpress.json")
	if err != nil {
		t.Fatalf("load theme garden config: %v", err)
	}
	stats, err := New(cfg, Options{}).Build()
	if err != nil {
		t.Fatalf("build theme garden: %v", err)
	}
	if stats.PageCount != 7 {
		t.Fatalf("theme garden source pages = %d, want 7", stats.PageCount)
	}
	if stats.WarningCount != 0 {
		t.Fatalf("theme garden warnings = %d, want 0", stats.WarningCount)
	}

	siteDir := filepath.Join(projectDir, "_site")
	for _, rel := range []string{
		"index.html",
		"notes/index.html",
		"notes/components/index.html",
		"notes/callouts/index.html",
		"journal/index.html",
		"projects/index.html",
		"tags/index.html",
		"tags/design/index.html",
		"graph.json",
		"search-index.json",
		"feed.xml",
		"style.css",
		"static/images/theme-swatch.svg",
		"static/leafpress/mermaid/mermaid.min.js",
	} {
		if _, err := os.Stat(filepath.Join(siteDir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("expected theme garden artifact %s: %v", rel, err)
		}
	}

	components := readThemeFixtureFile(t, siteDir, "notes/components/index.html")
	assertThemeFixtureContains(t, "component gallery", components,
		`data-lp-theme="terminal"`,
		`class="lp-toc"`,
		`class="lp-tag"`,
		`<table>`,
		`class="chroma"`,
		`<div class="mermaid">`,
		`theme-swatch.svg`,
		`class="lp-backlinks"`,
		`class="lp-search-overlay"`,
		`class="lp-graph-overlay"`,
	)

	styles := readThemeFixtureFile(t, siteDir, "style.css")
	assertThemeFixtureContains(t, "theme stylesheet", styles,
		"leafpress Base Styles",
		"leafpress Classic Theme",
		"leafpress Terminal Theme",
	)

	callouts := readThemeFixtureFile(t, siteDir, "notes/callouts/index.html")
	for _, kind := range []string{
		"note", "info", "abstract", "tip", "success", "important", "warning",
		"question", "danger", "failure", "bug", "example", "quote", "todo",
	} {
		assertThemeFixtureContains(t, "callout conservatory", callouts, "lp-callout-"+kind)
	}

	connections := readThemeFixtureFile(t, siteDir, "notes/connections/index.html")
	if strings.Contains(connections, `class="lp-toc"`) {
		t.Error("connected note explicitly disables its table of contents")
	}
	assertThemeFixtureContains(t, "explicit section index", readThemeFixtureFile(t, siteDir, "notes/index.html"),
		`class="lp-section-intro lp-content"`, `lp-index-growth--seedling`, `lp-index-growth--evergreen`)
	assertThemeFixtureContains(t, "generated section index", readThemeFixtureFile(t, siteDir, "journal/index.html"),
		`class="lp-section"`, `The First Journal Entry`)

	var graph struct {
		Nodes []json.RawMessage `json:"nodes"`
		Edges []json.RawMessage `json:"edges"`
	}
	if err := json.Unmarshal([]byte(readThemeFixtureFile(t, siteDir, "graph.json")), &graph); err != nil {
		t.Fatalf("decode graph.json: %v", err)
	}
	if len(graph.Nodes) != 7 || len(graph.Edges) < 12 {
		t.Errorf("theme garden graph has %d nodes and %d edges; want 7 nodes and at least 12 edges", len(graph.Nodes), len(graph.Edges))
	}

	var search []json.RawMessage
	if err := json.Unmarshal([]byte(readThemeFixtureFile(t, siteDir, "search-index.json")), &search); err != nil {
		t.Fatalf("decode search-index.json: %v", err)
	}
	if len(search) != 6 {
		t.Errorf("theme garden search entries = %d, want 6", len(search))
	}

	css := readThemeFixtureFile(t, siteDir, "style.css")
	assertThemeFixtureContains(t, "default stylesheet", css,
		"/* leafpress Base Styles */", "/* leafpress Classic Theme */", "/* Self-hosted fonts */")
	if strings.Contains(css, "/* User Styles */") {
		t.Error("theme garden must exercise pristine theme output without user style.css")
	}
}

func copyThemeGardenFixture(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read theme fixture directory %s: %v", src, err)
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatalf("create copied theme fixture directory %s: %v", dst, err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyThemeGardenFixture(t, srcPath, dstPath)
			continue
		}
		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("read theme fixture file %s: %v", srcPath, err)
		}
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			t.Fatalf("copy theme fixture file %s: %v", dstPath, err)
		}
	}
}

func readThemeFixtureFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read theme fixture artifact %s: %v", rel, err)
	}
	return string(data)
}

func assertThemeFixtureContains(t *testing.T, surface, content string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(content, want) {
			t.Errorf("%s is missing %q", surface, want)
		}
	}
}
