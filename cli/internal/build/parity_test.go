package build

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/assets"
	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/render"
)

// The parity suite proves the CLI build and the embedded renderer are two
// views of one site: same canonical leafpress.json, same markdown, same
// output paths, same generated artifacts, and an export that contains every
// asset the renderer's manifest promises.
//
// Intentional interface divergences (everything else must match):
//
//   - Hosted home: the renderer always emits Output.Index; a native site
//     without a root index page omits / entirely. Guarded by
//     TestParityHTMLPaths (the exclusion is asserted, not assumed).
//   - Hosted hardening: the renderer escapes raw author HTML
//     (SetEscapeRawHTML), degrades broken wikilinks to plain text
//     (SetPlainBrokenLinks), and escapes site fields (safeSiteData); the
//     CLI trusts local content. Full-page byte equality is therefore out of
//     scope — structural HTML parity (hrefs, resolved links) is asserted
//     instead in TestParityPageHTMLStructure.
//   - Missing custom fonts: the CLI hard-fails the build; the renderer
//     warns (a pure transform cannot check storage). Hosts must treat the
//     warning as a publish blocker.

const parityConfig = `{
  "site": {
    "title": "Parity Garden",
    "description": "Cross-interface parity fixture",
    "author": "Shivam",
    "baseURL": "https://example.com/garden"
  },
  "navigation": {
    "mode": "explicit",
    "items": [
      {"label": "Home", "path": "/"},
      {"label": "Essays", "path": "/essays/"}
    ]
  },
  "theme": {
    "accent": "#336699",
    "background": {"light": "#fdf6e3", "dark": "#002b36"},
    "fontBody": "My Serif",
    "fonts": [
      {"family": "My Serif", "file": "static/fonts/my-serif.woff2", "weight": "400 700"}
    ]
  },
  "features": {
    "graph": true,
    "search": true,
    "toc": true,
    "backlinks": true,
    "wikilinks": true,
    "rss": true
  }
}`

// noteABody links to the same page in all three wikilink forms — display
// title, filename, and full slug — so parity covers the whole resolution
// matrix.
const noteABody = "Linking to [[Note B]], [[note-b]], and [[essays/note-b]] while thinking about systems.\n\n## Ideas\n\nSome plain markdown content.\n"
const noteBBody = "# Heading\n\nEssay content with a [link](https://example.org).\n"

const parityStyleCSS = ".my-rule { color: rebeccapurple; }\n"

// User-owned fixture bytes: on disk for the CLI, declared through the caller
// asset manifest for the renderer.
const (
	parityCustomFontBytes  = "wOF2fakefontbytes"
	parityUserFaviconBytes = "user favicon override bytes"
)

// buildParityProject runs the CLI build over the fixture and returns its
// output directory.
func buildParityProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	write("leafpress.json", parityConfig)
	write("style.css", parityStyleCSS)
	write("static/fonts/my-serif.woff2", parityCustomFontBytes)
	write("favicon.ico", parityUserFaviconBytes)
	write("note-a.md", "---\ntitle: Note A\ndate: 2026-01-05\ntags: [systems]\n---\n\n"+noteABody)
	write("essays/note-b.md", "---\ntitle: Note B\ndate: 2026-02-10\n---\n\n"+noteBBody)

	t.Chdir(dir)
	cfg, err := config.Load("leafpress.json")
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if _, err := New(cfg, Options{}).Build(); err != nil {
		t.Fatalf("CLI build: %v", err)
	}
	return filepath.Join(dir, "_site")
}

// renderParityGarden runs the embedded renderer over the equivalent input:
// same config and markdown, with the fixture's user files declared through
// the caller asset manifest (the hosted counterpart of files on disk) and
// asset bytes requested so exports can be compared byte-for-byte.
func renderParityGarden(t *testing.T) *render.Output {
	t.Helper()
	input := map[string]any{
		"contractVersion": 2,
		"config":          json.RawMessage(parityConfig),
		"render":          map[string]any{"slug": "parity"},
		"options":         map[string]any{"emitAssets": true},
		"content": map[string]any{
			"styleCSS": parityStyleCSS,
			"assets": []map[string]any{
				{
					"logicalPath": "static/fonts/my-serif.woff2",
					"contentType": "font/woff2",
					"sha256":      assets.Sum([]byte(parityCustomFontBytes)),
					"size":        len(parityCustomFontBytes),
				},
				{
					"logicalPath": "static/user-favicon.ico",
					"contentType": "image/x-icon",
					"sha256":      assets.Sum([]byte(parityUserFaviconBytes)),
					"size":        len(parityUserFaviconBytes),
					"outputPath":  "favicon.ico",
				},
			},
			"pages": []map[string]any{
				{
					"slug":      "note-a",
					"title":     "Note A",
					"markdown":  noteABody,
					"tags":      []string{"systems"},
					"createdAt": "2026-01-05T00:00:00Z",
				},
				{
					"slug":      "essays/note-b",
					"title":     "Note B",
					"markdown":  noteBBody,
					"createdAt": "2026-02-10T00:00:00Z",
				},
			},
		},
	}
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	out, err := render.Run(raw)
	if err != nil {
		t.Fatalf("render.Run: %v", err)
	}
	return out
}

// siteHTMLPaths walks _site and returns the relative paths of all HTML files.
func siteHTMLPaths(t *testing.T, siteDir string) map[string]bool {
	t.Helper()
	paths := map[string]bool{}
	err := filepath.Walk(siteDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(p, ".html") {
			rel, err := filepath.Rel(siteDir, p)
			if err != nil {
				return err
			}
			paths[filepath.ToSlash(rel)] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return paths
}

// rendererHTMLPaths projects renderer output onto the CLI's file layout.
// The hosted home (Output.Index) is excluded: hosted gardens always have a
// home, while native sites omit / when no root index page exists.
func rendererHTMLPaths(out *render.Output) map[string]bool {
	paths := map[string]bool{}
	for _, p := range out.Pages {
		paths[p.Slug+"/index.html"] = true
	}
	for _, s := range out.Sections {
		paths[s.Slug+"/index.html"] = true
	}
	if out.Tags.Index != "" {
		paths["tags/index.html"] = true
	}
	for _, tp := range out.Tags.Pages {
		paths["tags/"+tp.Tag+"/index.html"] = true
	}
	for _, a := range out.Artifacts {
		if strings.HasSuffix(a.Path, ".html") {
			paths[a.Path] = true
		}
	}
	return paths
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func TestParityHTMLPaths(t *testing.T) {
	siteDir := buildParityProject(t)
	out := renderParityGarden(t)

	cliPaths := siteHTMLPaths(t, siteDir)
	rendererPaths := rendererHTMLPaths(out)

	for p := range rendererPaths {
		if !cliPaths[p] {
			t.Errorf("renderer emits %s but the CLI export lacks it", p)
		}
	}
	for p := range cliPaths {
		if !rendererPaths[p] {
			t.Errorf("CLI exports %s but the renderer does not emit it", p)
		}
	}
	if t.Failed() {
		t.Logf("cli: %v", sortedKeys(cliPaths))
		t.Logf("renderer: %v", sortedKeys(rendererPaths))
	}
}

func TestParityCSSAndTextArtifacts(t *testing.T) {
	siteDir := buildParityProject(t)
	out := renderParityGarden(t)

	readSite := func(rel string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(siteDir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("CLI export missing %s: %v", rel, err)
		}
		return string(data)
	}
	artifactContent := func(path string) string {
		t.Helper()
		for _, a := range out.Artifacts {
			if a.Path == path {
				if a.Encoding != "utf8" {
					t.Fatalf("artifact %s encoding %q", path, a.Encoding)
				}
				return a.Content
			}
		}
		t.Fatalf("renderer output missing artifact %s", path)
		return ""
	}

	if readSite("style.css") != out.CSS {
		t.Error("style.css differs between CLI and renderer")
	}
	var clientArtifacts []render.OutputArtifact
	for _, a := range out.Artifacts {
		if strings.HasPrefix(a.Path, "static/leafpress/app.") && strings.HasSuffix(a.Path, ".js") {
			clientArtifacts = append(clientArtifacts, a)
		}
	}
	if len(clientArtifacts) != 1 {
		t.Fatalf("renderer client script artifacts = %d, want exactly one", len(clientArtifacts))
	}
	client := clientArtifacts[0]
	if client.ContentType != "text/javascript; charset=utf-8" || client.Encoding != "utf8" {
		t.Fatalf("renderer client script metadata = (%q, %q), want JavaScript utf8", client.ContentType, client.Encoding)
	}
	if want := "static/leafpress/app." + assets.Sum([]byte(client.Content))[:32] + ".js"; client.Path != want {
		t.Fatalf("renderer client script path = %q, want %q", client.Path, want)
	}
	if readSite(client.Path) != client.Content {
		t.Error("shared client script differs between CLI and renderer")
	}
	for _, path := range []string{"robots.txt", "404.html"} {
		if readSite(path) != artifactContent(path) {
			t.Errorf("%s differs between CLI and renderer", path)
		}
	}

	// Order-independent comparison for artifacts assembled from page sets.
	locRegex := regexp.MustCompile(`<loc>(.*?)</loc>`)
	extractSet := func(re *regexp.Regexp, s string) []string {
		var vals []string
		for _, m := range re.FindAllStringSubmatch(s, -1) {
			vals = append(vals, m[1])
		}
		sort.Strings(vals)
		return vals
	}
	cliLocs := extractSet(locRegex, readSite("sitemap.xml"))
	renderLocs := extractSet(locRegex, artifactContent("sitemap.xml"))
	if strings.Join(cliLocs, "|") != strings.Join(renderLocs, "|") {
		t.Errorf("sitemap locs differ:\ncli: %v\nrenderer: %v", cliLocs, renderLocs)
	}

	linkRegex := regexp.MustCompile(`<link>(.*?)</link>`)
	cliLinks := extractSet(linkRegex, readSite("feed.xml"))
	renderLinks := extractSet(linkRegex, artifactContent("feed.xml"))
	if strings.Join(cliLinks, "|") != strings.Join(renderLinks, "|") {
		t.Errorf("feed links differ:\ncli: %v\nrenderer: %v", cliLinks, renderLinks)
	}

	// graph.json and search-index.json: compare URL sets.
	type graphShape struct {
		Nodes []struct {
			URL   string `json:"url"`
			Title string `json:"title"`
		} `json:"nodes"`
		Edges []struct {
			Source string `json:"source"`
			Target string `json:"target"`
		} `json:"edges"`
	}
	var cliGraph, renderGraph graphShape
	if err := json.Unmarshal([]byte(readSite("graph.json")), &cliGraph); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(artifactContent("graph.json")), &renderGraph); err != nil {
		t.Fatal(err)
	}
	graphKey := func(g graphShape) string {
		var parts []string
		for _, n := range g.Nodes {
			parts = append(parts, "n:"+n.URL+"#"+n.Title)
		}
		for _, e := range g.Edges {
			parts = append(parts, "e:"+e.Source+">"+e.Target)
		}
		sort.Strings(parts)
		return strings.Join(parts, "|")
	}
	if graphKey(cliGraph) != graphKey(renderGraph) {
		t.Errorf("graph.json differs:\ncli: %s\nrenderer: %s", graphKey(cliGraph), graphKey(renderGraph))
	}

	var cliSearch, renderSearch []struct {
		URL     string `json:"url"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(readSite("search-index.json")), &cliSearch); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(artifactContent("search-index.json")), &renderSearch); err != nil {
		t.Fatal(err)
	}
	searchKey := func(entries []struct {
		URL     string `json:"url"`
		Content string `json:"content"`
	}) string {
		var parts []string
		for _, e := range entries {
			parts = append(parts, e.URL+"#"+e.Content)
		}
		sort.Strings(parts)
		return strings.Join(parts, "|")
	}
	if searchKey(cliSearch) != searchKey(renderSearch) {
		t.Error("search-index.json content differs between CLI and renderer")
	}
}

func TestParityExportContainsManifestAssets(t *testing.T) {
	siteDir := buildParityProject(t)
	out := renderParityGarden(t)

	if out.AssetRegistryID != assets.RegistryID() {
		t.Errorf("assetRegistryId = %q, want %q", out.AssetRegistryID, assets.RegistryID())
	}
	if len(out.AssetManifest) == 0 {
		t.Fatal("renderer emitted empty asset manifest")
	}
	for _, entry := range out.AssetManifest {
		rel := entry.EffectiveOutputPath()
		data, err := os.ReadFile(filepath.Join(siteDir, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("export missing manifest asset %s at %s: %v", entry.LogicalPath, rel, err)
			continue
		}
		if got := assets.Sum(data); got != entry.SHA256 {
			t.Errorf("%s: exported hash %s != manifest %s", entry.LogicalPath, got, entry.SHA256)
		}
		if int64(len(data)) != entry.Size {
			t.Errorf("%s: exported size %d != manifest %d", entry.LogicalPath, len(data), entry.Size)
		}
	}

	// The manifest must cover exactly the bundled families the theme uses:
	// body is a custom font, so only Bricolage Grotesque + JetBrains Mono faces and
	// their two OFL license texts.
	fontEntries := 0
	for _, entry := range out.AssetManifest {
		if strings.HasPrefix(entry.LogicalPath, assets.BuiltinPrefix+"fonts/") {
			fontEntries++
			if strings.Contains(entry.LogicalPath, "inter") {
				t.Errorf("unused family asset %s in manifest", entry.LogicalPath)
			}
		}
	}
	if fontEntries != 8 {
		t.Errorf("manifest has %d font entries, want 8 (2 Bricolage subsets + 4 JetBrains Mono faces + 2 OFL texts)", fontEntries)
	}

	// The caller-declared user assets are in the same manifest, so a host
	// can pin them into the publication snapshot with everything else.
	byOutput := map[string]assets.Asset{}
	for _, a := range out.AssetManifest {
		byOutput[a.EffectiveOutputPath()] = a
	}
	if a, ok := byOutput["static/fonts/my-serif.woff2"]; !ok || a.SHA256 != assets.Sum([]byte(parityCustomFontBytes)) {
		t.Error("caller-declared custom font missing from combined manifest")
	}
	// The user favicon override replaces the built-in entry, matching the
	// CLI preferring the on-disk favicon — the generic loop above already
	// proved the export at favicon.ico carries this same hash.
	if a, ok := byOutput["favicon.ico"]; !ok || a.SHA256 != assets.Sum([]byte(parityUserFaviconBytes)) {
		t.Error("user favicon override missing from combined manifest")
	}
}

func TestParityEmittedAssetBytesMatchExport(t *testing.T) {
	siteDir := buildParityProject(t)
	out := renderParityGarden(t)

	emitted := 0
	for _, artifact := range out.Artifacts {
		if artifact.Encoding != "base64" {
			continue
		}
		emitted++
		raw, err := base64.StdEncoding.DecodeString(artifact.Content)
		if err != nil {
			t.Errorf("%s: invalid base64: %v", artifact.Path, err)
			continue
		}
		// Emitted artifacts are keyed by effective output path, so they
		// must land byte-for-byte where the CLI export put the same file.
		exported, err := os.ReadFile(filepath.Join(siteDir, filepath.FromSlash(artifact.Path)))
		if err != nil {
			t.Errorf("export has no file at emitted artifact path %s: %v", artifact.Path, err)
			continue
		}
		if !bytes.Equal(raw, exported) {
			t.Errorf("%s: emitted bytes differ from CLI export", artifact.Path)
		}
	}
	if emitted == 0 {
		t.Fatal("emitAssets produced no base64 artifacts")
	}
	// Exactly one artifact per non-overridden built-in manifest entry: a
	// partial-emission regression must fail loudly.
	builtinEntries := 0
	for _, entry := range out.AssetManifest {
		if b, ok := assets.BuiltinByLogicalPath(entry.LogicalPath); ok && b.Asset.SHA256 == entry.SHA256 {
			builtinEntries++
		}
	}
	if emitted != builtinEntries {
		t.Errorf("emitted %d artifacts for %d built-in manifest entries", emitted, builtinEntries)
	}
	// The overridden favicon must not be emitted from the registry: the
	// export at favicon.ico is the user's file.
	for _, artifact := range out.Artifacts {
		if artifact.Encoding == "base64" && artifact.Path == "favicon.ico" {
			t.Error("overridden favicon emitted from the built-in registry")
		}
	}
}

func TestParityStylesheetFonts(t *testing.T) {
	siteDir := buildParityProject(t)
	out := renderParityGarden(t)

	cliCSS, err := os.ReadFile(filepath.Join(siteDir, "style.css"))
	if err != nil {
		t.Fatal(err)
	}

	// @font-face rules live in the shared stylesheet with stylesheet-
	// relative URLs on both interfaces; byte equality is asserted by
	// TestParityCSSAndTextArtifacts.
	for _, ref := range []string{
		`url("static/fonts/my-serif.woff2")`,
		`url("static/leafpress/fonts/bricolage-grotesque-normal-latin.woff2")`,
	} {
		if !strings.Contains(string(cliCSS), ref) {
			t.Errorf("CLI stylesheet missing font reference %s", ref)
		}
		if !strings.Contains(out.CSS, ref) {
			t.Errorf("renderer stylesheet missing font reference %s", ref)
		}
	}
	if strings.Contains(string(cliCSS), "fonts.googleapis.com") || strings.Contains(out.CSS, "fonts.googleapis.com") {
		t.Error("fully self-hosted theme must not reference Google Fonts on either interface")
	}
}

// pageBySlug returns a renderer page's HTML.
func pageBySlug(t *testing.T, out *render.Output, slug string) string {
	t.Helper()
	for _, p := range out.Pages {
		if p.Slug == slug {
			return p.HTML
		}
	}
	t.Fatalf("renderer output missing page %q", slug)
	return ""
}

func TestParityPageHTMLStructure(t *testing.T) {
	siteDir := buildParityProject(t)
	out := renderParityGarden(t)

	// The home divergence is asserted, not assumed: hosted always has a
	// home; this fixture's native site must not.
	if out.Index == "" {
		t.Error("renderer must always emit a hosted home")
	}
	if _, err := os.Stat(filepath.Join(siteDir, "index.html")); !os.IsNotExist(err) {
		t.Error("fixture has no root index page, so the CLI must not export /index.html")
	}

	readSite := func(rel string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(siteDir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("CLI export missing %s: %v", rel, err)
		}
		return string(data)
	}

	// Structural markers both interfaces must render identically:
	// basePath-prefixed hrefs and the fully resolved wikilink matrix.
	checks := []struct {
		name     string
		cliHTML  string
		rendered string
		markers  []string
	}{
		{
			name:     "note-a",
			cliHTML:  readSite("note-a/index.html"),
			rendered: pageBySlug(t, out, "note-a"),
			markers: []string{
				`href="/garden/essays/note-b/"`, // resolved wikilinks
				`href="/garden/style.css"`,
				`href="/garden/tags/systems/"`,
			},
		},
		{
			name:     "essays/note-b",
			cliHTML:  readSite("essays/note-b/index.html"),
			rendered: pageBySlug(t, out, "essays/note-b"),
			markers:  []string{`href="/garden/style.css"`, `href="/garden/favicon.svg"`},
		},
		{
			name:     "tags/systems",
			cliHTML:  readSite("tags/systems/index.html"),
			rendered: out.Tags.Pages[0].HTML,
			markers:  []string{`href="/garden/note-a/"`},
		},
	}
	for _, c := range checks {
		for _, marker := range c.markers {
			if !strings.Contains(c.cliHTML, marker) {
				t.Errorf("%s (CLI): missing %s", c.name, marker)
			}
			if !strings.Contains(c.rendered, marker) {
				t.Errorf("%s (renderer): missing %s", c.name, marker)
			}
		}
	}

	// All three wikilink forms resolve to the same target on both sides.
	wantLinks := 3
	cliLinks := strings.Count(checks[0].cliHTML, `href="/garden/essays/note-b/"`)
	renderLinks := strings.Count(checks[0].rendered, `href="/garden/essays/note-b/"`)
	if cliLinks < wantLinks || renderLinks < wantLinks {
		t.Errorf("wikilink matrix: CLI resolved %d, renderer %d, want >= %d each", cliLinks, renderLinks, wantLinks)
	}
}

func TestParityNavigationOutput(t *testing.T) {
	siteDir := buildParityProject(t)
	out := renderParityGarden(t)
	cliPage, err := os.ReadFile(filepath.Join(siteDir, "note-a", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	renderedPage := pageBySlug(t, out, "note-a")
	for _, marker := range []string{
		`href="/garden/"`,
		`class="lp-nav-link" href="/garden/essays/"`,
		`>Home</a>`,
		`>Essays</a>`,
	} {
		if !strings.Contains(string(cliPage), marker) {
			t.Errorf("CLI navigation missing %q", marker)
		}
		if !strings.Contains(renderedPage, marker) {
			t.Errorf("renderer navigation missing %q", marker)
		}
	}
}

// fontParityCase builds a minimal project + equivalent renderer input for a
// given theme and returns the CLI page head, CLI stylesheet, renderer page
// head, and renderer stylesheet.
func fontParityCase(t *testing.T, themeJSON string) (string, string, string, string) {
	t.Helper()
	dir := t.TempDir()
	cfgJSON := `{"site": {"title": "Fonts"}, "theme": ` + themeJSON + `}`
	if err := os.WriteFile(filepath.Join(dir, "leafpress.json"), []byte(cfgJSON), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte("# Note\n\nhi\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	cfg, err := config.Load("leafpress.json")
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if _, err := New(cfg, Options{}).Build(); err != nil {
		t.Fatalf("CLI build: %v", err)
	}
	cliPage, err := os.ReadFile(filepath.Join(dir, "_site", "note", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	cliCSS, err := os.ReadFile(filepath.Join(dir, "_site", "style.css"))
	if err != nil {
		t.Fatal(err)
	}

	out, err := render.Run([]byte(`{
	  "contractVersion": 2,
	  "render": {"slug": "fonts"},
	  "config": ` + cfgJSON + `,
	  "content": {"pages": [{"slug": "note", "title": "Note", "markdown": "hi"}]}
	}`))
	if err != nil {
		t.Fatalf("render.Run: %v", err)
	}
	return string(cliPage), string(cliCSS), pageBySlug(t, out, "note"), out.CSS
}

func TestParityUnbundledFontFallsBackOnBothInterfaces(t *testing.T) {
	cliPage, cliCSS, renderPage, renderCSS := fontParityCase(t, `{"fontBody": "Lobster"}`)
	for name, doc := range map[string]string{
		"cli page": cliPage, "cli css": cliCSS, "renderer page": renderPage, "renderer css": renderCSS,
	} {
		if strings.Contains(doc, "fonts.googleapis.com") {
			t.Errorf("%s references Google Fonts without the remoteFonts opt-in", name)
		}
	}
	if strings.Contains(cliCSS, `font-family: "Lobster"`) || strings.Contains(renderCSS, `font-family: "Lobster"`) {
		t.Error("unbundled family must not get @font-face rules")
	}
}

func TestParityRemoteFontsOptInMatches(t *testing.T) {
	cliPage, cliCSS, renderPage, renderCSS := fontParityCase(t, `{"fontBody": "Lobster", "remoteFonts": true}`)
	if cliCSS != renderCSS {
		t.Error("stylesheets differ under remoteFonts")
	}
	for name, page := range map[string]string{"cli": cliPage, "renderer": renderPage} {
		if !strings.Contains(page, "fonts.googleapis.com/css2?family=Lobster") {
			t.Errorf("%s page missing opted-in remote font link", name)
		}
		if strings.Contains(page, "family=Crimson+Pro") {
			t.Errorf("%s page leaked a bundled family into the remote URL", name)
		}
	}
}
