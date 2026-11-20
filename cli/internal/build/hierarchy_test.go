package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/render"
)

var listingLinkRE = regexp.MustCompile(`class="lp-index-link" href="([^"]+)"`)

func listingLinks(markup string) []string {
	var links []string
	for _, m := range listingLinkRE.FindAllStringSubmatch(markup, -1) {
		links = append(links, m[1])
	}
	return links
}

func TestHierarchicalCLIAndRendererParity(t *testing.T) {
	for _, baseURL := range []string{"", "https://custom.example", "https://example.com/garden/alice"} {
		for _, explicit := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/explicit=%t", baseURL, explicit), func(t *testing.T) {
				dir := t.TempDir()
				t.Chdir(dir)
				cfg := config.Default()
				cfg.Site.Title = "Hierarchy"
				cfg.Site.BaseURL = baseURL
				cfg.Features.Graph = true
				cfg.Features.RSS = true
				pages := []render.InputPage{
					{Slug: "", Title: "Home", IsIndex: true, Sort: "title"},
					{Slug: "now", Title: "Now"},
					{Slug: "clients/acme/projects/plan", Title: "Acme Plan", Markdown: "[[clients/birch/projects/plan]]", Tags: []string{"plans"}},
					{Slug: "clients/birch/projects/plan", Title: "Birch Plan"},
				}
				want := map[string][]string{
					"":                       {"/clients/", "/now/"},
					"clients":                {"/clients/acme/", "/clients/birch/"},
					"clients/acme":           {"/clients/acme/projects/"},
					"clients/acme/projects":  {"/clients/acme/projects/plan/"},
					"clients/birch":          {"/clients/birch/projects/"},
					"clients/birch/projects": {"/clients/birch/projects/plan/"},
				}
				if explicit {
					hidden := false
					pages = append(pages,
						render.InputPage{Slug: "clients", Title: "Our Clients", IsIndex: true, Markdown: "Ancestor introduction", Sort: "title"},
						render.InputPage{Slug: "clients/acme", Title: "Acme", IsIndex: true, Markdown: "Acme introduction", Sort: "title"},
						render.InputPage{Slug: "clients/acme/brief", Title: "A Brief"},
						render.InputPage{Slug: "clients/acme/empty", Title: "Empty", IsIndex: true, Markdown: "Empty introduction"},
						render.InputPage{Slug: "clients/birch/projects", Title: "Private listing", IsIndex: true, Markdown: "Hidden introduction", ShowList: &hidden},
					)
					want[""] = []string{"/now/", "/clients/"}
					want["clients/acme"] = []string{"/clients/acme/brief/", "/clients/acme/empty/", "/clients/acme/projects/"}
					want["clients/acme/empty"] = nil
					want["clients/birch/projects"] = nil
				}
				for i := range pages {
					p := &pages[i]
					p.CreatedAt = "2026-01-01T00:00:00Z"
					p.UpdatedAt = "2026-01-01T00:00:00Z"
					rel := p.Slug + ".md"
					if p.IsIndex {
						rel = filepath.Join(p.Slug, "_index.md")
					}
					if err := os.MkdirAll(filepath.Dir(rel), 0755); err != nil {
						t.Fatal(err)
					}
					front := fmt.Sprintf("---\ntitle: %q\ndate: 2026-01-01\nmodified: 2026-01-01\n", p.Title)
					if p.Sort != "" {
						front += "sort: " + p.Sort + "\n"
					}
					if p.ShowList != nil {
						front += fmt.Sprintf("showList: %t\n", *p.ShowList)
					}
					if len(p.Tags) > 0 {
						front += "tags: [plans]\n"
					}
					if err := os.WriteFile(rel, []byte(front+"---\n\n"+p.Markdown), 0644); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := New(cfg, Options{}).Build(); err != nil {
					t.Fatal(err)
				}
				rawConfig, _ := json.Marshal(cfg)
				raw, _ := json.Marshal(render.Input{Config: rawConfig, Render: render.RenderOpts{Slug: "test"}, Content: render.Content{Pages: pages}})
				out, err := render.Run(raw)
				if err != nil {
					t.Fatal(err)
				}
				rendered := map[string]string{"": out.Index}
				for _, p := range out.Pages {
					rendered[p.Slug] = p.HTML
				}
				for _, p := range out.Sections {
					rendered[p.Slug] = p.HTML
				}
				prefix := strings.TrimPrefix(baseURL, "https://example.com")
				if baseURL != "https://example.com/garden/alice" {
					prefix = ""
				}
				for section, links := range want {
					expected := append([]string(nil), links...)
					for i := range expected {
						expected[i] = prefix + expected[i]
					}
					cliHTML := string(readHierarchyFile(t, filepath.Join("_site", section, "index.html")))
					for surface, markup := range map[string]string{"cli": cliHTML, "renderer": rendered[section]} {
						if got := listingLinks(markup); !reflect.DeepEqual(got, expected) {
							t.Errorf("%s section %q links = %v, want %v", surface, section, got, expected)
						}
						if section != "" && !regexp.MustCompile(`class="lp-nav-link lp-nav-link--active[^\"]*" href="`+regexp.QuoteMeta(prefix)+`/clients/"`).MatchString(markup) {
							t.Errorf("%s section %q missing active ancestor nav", surface, section)
						}
					}
				}
				if explicit {
					for slug, intro := range map[string]string{"clients": "Ancestor introduction", "clients/acme": "Acme introduction", "clients/acme/empty": "Empty introduction", "clients/birch/projects": "Hidden introduction"} {
						if !strings.Contains(rendered[slug], intro) || !strings.Contains(string(readHierarchyFile(t, filepath.Join("_site", slug, "index.html"))), intro) {
							t.Errorf("missing intro for %s", slug)
						}
					}
				}
				for _, a := range out.Artifacts {
					if a.Path == "graph.json" || a.Path == "search-index.json" || a.Path == "sitemap.xml" {
						if got := string(readHierarchyFile(t, filepath.Join("_site", a.Path))); normalizeHierarchyArtifact(t, a.Path, got) != normalizeHierarchyArtifact(t, a.Path, a.Content) {
							t.Errorf("artifact parity: %s", a.Path)
						}
					}
					if a.Path == "search-index.json" || a.Path == "graph.json" {
						if strings.Contains(a.Content, `"url": "`+prefix+`/clients/acme/projects/"`) {
							t.Errorf("synthetic section in %s", a.Path)
						}
					}
				}
				for _, slug := range []string{"clients/acme/projects/plan", "clients/birch/projects/plan"} {
					if rendered[slug] == "" {
						t.Errorf("lost note permalink: %s", slug)
					}
					readHierarchyFile(t, filepath.Join("_site", slug, "index.html"))
				}
				backlink := `class="lp-backlink" href="` + prefix + `/clients/acme/projects/plan/">Acme Plan</a>`
				if !strings.Contains(rendered["clients/birch/projects/plan"], backlink) {
					t.Error("missing cross-branch backlink")
				}
				assertFileContains(t, filepath.Join("_site", "clients/birch/projects/plan/index.html"), backlink)
				if baseURL != "" {
					artifacts := map[string]string{}
					for _, a := range out.Artifacts {
						artifacts[a.Path] = a.Content
					}
					for section := range want {
						if section == "" {
							continue
						}
						feedPath := section + "/feed.xml"
						readHierarchyFile(t, filepath.Join("_site", feedPath))
						if artifacts[feedPath] == "" {
							t.Errorf("missing feed %s", feedPath)
						}
						if !strings.Contains(artifacts["sitemap.xml"], "<loc>"+baseURL+"/"+section+"/</loc>") {
							t.Errorf("missing sitemap section %s", section)
						}
					}
					if strings.Count(artifacts["clients/feed.xml"], "<item>") != 2+boolInt(explicit) {
						t.Error("ancestor feed does not contain precisely descendant notes")
					}
				}
			})
		}
	}
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func readHierarchyFile(t *testing.T, p string) []byte {
	t.Helper()
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Authored input order is transport-specific; compare complete artifact records
// independently of that order, including sitemap dates and graph edges.
func normalizeHierarchyArtifact(t *testing.T, name, data string) string {
	t.Helper()
	if name == "sitemap.xml" {
		records := regexp.MustCompile(`(?s)<url>.*?</url>`).FindAllString(data, -1)
		sort.Strings(records)
		return strings.Join(records, "\n")
	}
	var value any
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		t.Fatal(err)
	}
	var normalize func(any) any
	normalize = func(v any) any {
		switch v := v.(type) {
		case []any:
			for i := range v {
				v[i] = normalize(v[i])
			}
			sort.Slice(v, func(i, j int) bool {
				a, _ := json.Marshal(v[i])
				b, _ := json.Marshal(v[j])
				return string(a) < string(b)
			})
		case map[string]any:
			for k, item := range v {
				v[k] = normalize(item)
			}
		}
		return v
	}
	result, _ := json.Marshal(normalize(value))
	return string(result)
}

func TestIncrementalHierarchy(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	write := func(p, data string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("_index.md", "---\ntitle: Home\n---\n")
	cfg := config.Default()
	cfg.Site.BaseURL = "https://example.com"
	b := New(cfg, Options{})
	if _, err := b.Build(); err != nil {
		t.Fatal(err)
	}
	leaf := "clients/acme/projects/plan.md"
	write(leaf, "---\ntitle: Plan\ndate: 2026-01-01\n---\n")
	if _, err := b.RebuildIncremental(filepath.Join(dir, leaf), ChangeCreate); err != nil {
		t.Fatal(err)
	}
	for _, section := range []string{"clients", "clients/acme", "clients/acme/projects"} {
		readHierarchyFile(t, filepath.Join("_site", section, "index.html"))
	}
	assertFileContains(t, "_site/index.html", `class="lp-index-link" href="/clients/"`)
	write(leaf, "---\ntitle: Revised Plan\ndate: 2026-01-01\nmodified: 2026-02-01\n---\n")
	if _, err := b.RebuildIncremental(filepath.Join(dir, leaf), ChangeModify); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, "_site/clients/acme/projects/index.html", "Revised Plan")
	assertFileContains(t, "_site/clients/index.html", `datetime="2026-02-01"`)
	assertFileContains(t, "_site/index.html", `datetime="2026-02-01"`)
	// An empty authored index keeps itself and its otherwise empty ancestors.
	empty := "clients/birch/empty/_index.md"
	write(empty, "---\ntitle: Empty\n---\n")
	if _, err := b.RebuildIncremental(filepath.Join(dir, empty), ChangeCreate); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, "_site/clients/index.html", `class="lp-index-link" href="/clients/birch/"`)
	write("now.md", "---\ntitle: Now\n---\n")
	if _, err := b.RebuildIncremental(filepath.Join(dir, "now.md"), ChangeCreate); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, "_site/clients/birch/index.html", `href="/now/">Now</a>`)
	if err := os.Remove("now.md"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.RebuildIncremental(filepath.Join(dir, "now.md"), ChangeDelete); err != nil {
		t.Fatal(err)
	}
	assertFileNotContains(t, "_site/clients/birch/index.html", `href="/now/">Now</a>`)
	if err := os.Remove(leaf); err != nil {
		t.Fatal(err)
	}
	if _, err := b.RebuildIncremental(filepath.Join(dir, leaf), ChangeDelete); err != nil {
		t.Fatal(err)
	}
	for _, section := range []string{"clients/acme", "clients/acme/projects"} {
		for _, file := range []string{"index.html", "feed.xml"} {
			if _, err := os.Stat(filepath.Join("_site", section, file)); !os.IsNotExist(err) {
				t.Errorf("stale artifact %s/%s: %v", section, file, err)
			}
		}
	}
	assertFileNotContains(t, "_site/clients/index.html", `class="lp-index-link" href="/clients/acme/"`)
	readHierarchyFile(t, "_site/clients/birch/empty/index.html")
	if err := os.Remove(empty); err != nil {
		t.Fatal(err)
	}
	if _, err := b.RebuildIncremental(filepath.Join(dir, empty), ChangeDelete); err != nil {
		t.Fatal(err)
	}
	assertFileNotContains(t, "_site/index.html", `class="lp-index-link" href="/clients/"`)
	if _, err := os.Stat("_site/clients/index.html"); !os.IsNotExist(err) {
		t.Errorf("stale ancestor: %v", err)
	}
}

func TestBuildRejectsAncestorDocumentCollision(t *testing.T) {
	for _, rel := range []string{"clients/acme/projects/plan.md", "clients/acme/empty/_index.md"} {
		t.Run(rel, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			if err := os.WriteFile("clients.md", []byte("# Clients"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Dir(rel), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(rel, []byte("# Deep content"), 0644); err != nil {
				t.Fatal(err)
			}
			if _, err := New(config.Default(), Options{}).Build(); err == nil || !strings.Contains(err.Error(), `output route "/clients/"`) {
				t.Fatalf("collision error = %v", err)
			}
		})
	}
}
