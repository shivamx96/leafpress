package render

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestDefaultHeadUnchanged(t *testing.T) {
	// Captured from main (f137878) before the favicon/social metadata changes.
	want, err := os.ReadFile("testdata/default-head.html")
	if err != nil {
		t.Fatal(err)
	}
	html := runJSON(t, `{}`).Index
	got := html[strings.Index(html, "<head>") : strings.Index(html, "</head>")+7]
	if got != string(want) {
		t.Fatalf("default head changed:\n%s", got)
	}
}

func TestFaviconAndSocialFixtures(t *testing.T) {
	for _, name := range []string{"default", "custom"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile("testdata/" + name + "-social.json")
			if err != nil {
				t.Fatal(err)
			}
			out := runJSON(t, string(data))
			custom := name == "custom"
			docs := []string{out.Index, pageHTML(t, out, "notes/plain"), out.Tags.Index, out.Tags.Pages[0].HTML, pageHTML(t, out, "notes"), out.Sections[0].HTML, artifact(t, out, "404.html").Content}
			for _, html := range docs {
				if custom {
					for _, rel := range []string{"icon", "apple-touch-icon"} {
						link := `<link rel="` + rel + `" href="/garden/static/uploads/icon.png">`
						if strings.Count(html, link) != 1 {
							t.Errorf("missing unique custom link %s", link)
						}
					}
					if strings.Contains(html, "/favicon.") || strings.Contains(html, "/favicon-96x96.png") {
						t.Error("custom site links built-in favicon")
					}
				} else {
					for _, path := range []string{"favicon.svg", "favicon.ico", "favicon-96x96.png"} {
						if !strings.Contains(html, `href="/garden/`+path+`"`) {
							t.Errorf("missing default %s", path)
						}
					}
				}
			}
			for _, path := range []string{"favicon.svg", "favicon.ico", "favicon-96x96.png"} {
				manifest, emitted := false, false
				for _, a := range out.AssetManifest {
					if a.EffectiveOutputPath() == path {
						manifest = true
					}
				}
				for _, a := range out.Artifacts {
					if a.Path == path {
						emitted = true
					}
				}
				if manifest == custom || emitted == custom {
					t.Errorf("%s: manifest=%v emitted=%v custom=%v", path, manifest, emitted, custom)
				}
			}
			for _, html := range docs[:len(docs)-1] {
				image := ""
				if custom {
					image = "https://example.com/garden/static/uploads/social.jpg"
				}
				assertSocial(t, html, image)
			}
			assertSocial(t, pageHTML(t, out, "notes/photo"), "https://example.com/garden/static/uploads/page.webp")
			if custom {
				for _, w := range out.Warnings {
					if strings.Contains(w, "site.favicon") || strings.Contains(w, "site.image") {
						t.Error(w)
					}
				}
				for _, a := range out.Artifacts {
					if strings.HasPrefix(a.Path, "static/uploads/") {
						t.Errorf("caller bytes emitted: %s", a.Path)
					}
				}
			}
		})
	}
}

func assertSocial(t *testing.T, html, image string) {
	t.Helper()
	card := "summary"
	if image != "" {
		card = "summary_large_image"
	}
	if !strings.Contains(html, `<meta name="twitter:card" content="`+card+`">`) {
		t.Errorf("missing Twitter card %s", card)
	}
	for _, attr := range []string{`property="og:image"`, `name="twitter:image"`} {
		if image == "" {
			if strings.Contains(html, attr) {
				t.Errorf("unexpected image %s", attr)
			}
		} else if strings.Count(html, `<meta `+attr+` content="`+image+`">`) != 1 {
			t.Errorf("missing unique %s: %s", attr, image)
		}
	}
}

func TestSiteImageManifestWarnings(t *testing.T) {
	for _, field := range []string{"favicon", "image"} {
		for _, path := range []string{"/static/uploads/missing.png", "/outside.png", "/static/leafpress/favicon.svg"} {
			input, _ := json.Marshal(map[string]any{"config": map[string]any{"site": map[string]string{field: path}}})
			out := runJSON(t, string(input))
			warned := false
			for _, w := range out.Warnings {
				if strings.Contains(w, "site."+field+" references") {
					warned = true
				}
			}
			// The root favicon's logical path is not its served path.
			want := strings.HasPrefix(path, "/static/")
			if warned != want {
				t.Errorf("%s %s warning=%v want=%v", field, path, warned, want)
			}
		}
	}
}

func TestSiteImagesUseMergedManifestServedPaths(t *testing.T) {
	out := runJSON(t, `{"config":{"site":{"image":"/static/leafpress/fonts/inter-normal-latin.woff2"}}}`)
	// This entry comes from the selected built-in set, not caller declarations.
	for _, warning := range out.Warnings {
		if strings.Contains(warning, "site.image references") {
			t.Error(warning)
		}
	}
}

func TestCustomFaviconEscapedInHead(t *testing.T) {
	out := runJSON(t, `{"config":{"site":{"favicon":"/static/a&\"b.png"}}}`)
	if !strings.Contains(out.Index, `<link rel="icon" href="/static/a&amp;&#34;b.png">`) {
		t.Fatal("custom favicon was not escaped in head")
	}
}
