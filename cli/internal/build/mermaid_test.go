package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/assets"
	"github.com/shivamx96/leafpress/core/config"
)

// A sequence diagram whose labels carry KaTeX and an init directive that
// tries to switch HTML labels back on. This is the shape a hostile diagram
// would take: the directive is the bypass, the math is the payload carrier.
const mermaidSequenceKatexFixture = "---\ntitle: Diagram\n---\n\n" +
	"```mermaid\n" +
	"%%{init: {'securityLevel': 'loose', 'sequence': {'useHtmlLabels': true}, 'flowchart': {'htmlLabels': true}}}%%\n" +
	"sequenceDiagram\n" +
	"    participant A as Alice $$\\alpha_1$$\n" +
	"    participant B as Bob\n" +
	"    A->>B: Latency is $$O(n^2)$$\n" +
	"    B-->>A: Ack $$\\sum_{i=0}^{n} x_i$$\n" +
	"```\n"

func buildMermaidFixture(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "diagram.md"), []byte(mermaidSequenceKatexFixture), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if _, err := New(config.Default(), Options{}).Build(); err != nil {
		t.Fatal(err)
	}
	siteDir := filepath.Join(dir, "_site")
	html, err := os.ReadFile(filepath.Join(siteDir, "diagram", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	return siteDir, string(html)
}

// A page containing a Mermaid diagram must materialize the vendored script
// and reach it locally — no CDN request, per docs/07_ASSET_ARCHITECTURE.md.
func TestBuildMermaidSequenceWithKatexUsesVendoredAsset(t *testing.T) {
	siteDir, html := buildMermaidFixture(t)

	vendored := filepath.Join(siteDir, "static", "leafpress", "mermaid", "mermaid.min.js")
	info, err := os.Stat(vendored)
	if err != nil {
		t.Fatalf("mermaid script was not materialized: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("materialized mermaid script is empty")
	}
	if _, err := os.Stat(filepath.Join(siteDir, "static", "leafpress", "mermaid", "LICENSE.txt")); err != nil {
		t.Errorf("mermaid LICENSE was not materialized: %v", err)
	}
	if !strings.Contains(html, `class="mermaid"`) {
		t.Errorf("diagram block did not render as a mermaid node:\n%s", html)
	}
	if assets.MermaidVersion == "" {
		t.Error("MermaidVersion must stay pinned")
	}
}

// The hostile init directive travels with the diagram source into the page.
// It is inert because the shared client bundle locks those keys via `secure`,
// so assert the lock is present on the page that carries the directive.
func TestBuildMermaidDirectiveCannotUnlockHTMLLabels(t *testing.T) {
	siteDir, html := buildMermaidFixture(t)

	clientPath, client := readClientScript(t, siteDir)
	if !strings.Contains(html, clientPath) {
		t.Fatalf("diagram page does not load the shared client script %s", clientPath)
	}
	for _, want := range []string{"'htmlLabels'", "'flowchart'", "'sequence'", "'securityLevel'"} {
		if !strings.Contains(client, want) {
			t.Errorf("client bundle does not lock %s against diagram directives", want)
		}
	}
	if !strings.Contains(client, "securityLevel: 'strict'") {
		t.Error("client bundle does not pin securityLevel to strict")
	}
}
