package render

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/config"
)

func footerHTML(t *testing.T, document string) string {
	t.Helper()
	start := strings.Index(document, "<footer")
	end := strings.Index(document, "</footer>")
	if start < 0 || end < start {
		t.Fatal("missing footer")
	}
	return document[start : end+len("</footer>")]
}

func TestFooterMatrix(t *testing.T) {
	for _, source := range []struct{ name, footer, author, title, want string }{
		{"footer", "My footer", "", "Garden", "My footer"},
		{"footer beats author", "My footer", "Author", "Garden", "My footer"},
		{"author", "", "Author", "Garden", "&copy; Author. All rights reserved."},
		{"neither", "", "", "Garden", ""},
		{"empty title", "", "", "", ""},
		{"escaping", "<script>&\"'", "", "Garden", "&lt;script&gt;&amp;&#34;&#39;"},
	} {
		for _, credit := range []string{"default", "attribution", "unlinked attribution", "hidden", "explicit true"} {
			for _, fallback := range []string{"", "title"} {
				t.Run(source.name+"/"+credit+"/"+fallback, func(t *testing.T) {
					opts := RenderOpts{Slug: "garden", CopyrightFallback: fallback}
					wantCredit := `<span class="lp-footer-text">Grown with <a href="https://leafpress.in" target="_blank" rel="noopener noreferrer">leafpress</a></span>`
					switch credit {
					case "attribution":
						opts.FooterAttribution = &FooterAttribution{Name: "Host", URL: "https://host.example"}
						wantCredit = `<span class="lp-footer-text">Grown with <a href="https://host.example" target="_blank" rel="noopener noreferrer">Host</a></span>` + "\n    "
					case "unlinked attribution":
						opts.FooterAttribution = &FooterAttribution{Name: "Host"}
						wantCredit = `<span class="lp-footer-text">Grown with Host</span>` + "\n    "
					case "hidden":
						b := false
						opts.GeneratorCredit = &b
						wantCredit = ""
					case "explicit true":
						b := true
						opts.GeneratorCredit = &b
					}
					cfg, err := json.Marshal(map[string]any{"site": map[string]string{"footer": source.footer, "author": source.author, "title": source.title}})
					if err != nil {
						t.Fatal(err)
					}
					out, err := Render(&Input{Config: cfg, Render: opts})
					if err != nil {
						t.Fatal(err)
					}
					copyright := source.want
					if copyright == "" && fallback == "title" && source.title != "" {
						copyright = "&copy; " + source.title + ". All rights reserved."
					}
					if copyright != "" {
						copyright = `<span class="lp-footer-text">` + copyright + "</span>"
					}
					want := "<footer class=\"lp-footer\">\n    " + copyright + "\n    " + wantCredit + "\n  </footer>"
					if got := footerHTML(t, out.Index); got != want {
						t.Fatalf("footer = %q; want %q", got, want)
					}
				})
			}
		}
	}
}

// Exercise the identical shared validation through config parsing (CLI) and
// the renderer boundary, including normalization and multi-byte rune limits.
func TestFooterValidationParity(t *testing.T) {
	for _, tc := range []struct {
		name, value, want string
		invalid           bool
	}{
		{"200 runes", strings.Repeat("界", 200), strings.Repeat("界", 200), false},
		{"201 runes", strings.Repeat("界", 201), "", true},
		{"newline", "a\nb", "", true},
		{"control", "a\x00b", "", true},
		{"tab", "a\tb", "", true},
		{"unicode newline", "a\u2028b", "", true},
		{"blank", " \t\n ", "", false},
		{"trim", "  footer  ", "footer", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, _ := json.Marshal(map[string]any{"site": map[string]string{"footer": tc.value}})
			cfg, err := config.Parse(raw)
			if tc.invalid {
				if err == nil || !strings.Contains(err.Error(), "site.footer") {
					t.Fatalf("config error = %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			} else if cfg.Site.Footer != tc.want {
				t.Fatalf("footer = %q", cfg.Site.Footer)
			}
			_, err = Render(&Input{Config: raw})
			if tc.invalid {
				var inputErr *InputError
				if !errors.As(err, &inputErr) || !strings.Contains(err.Error(), "site.footer") {
					t.Fatalf("renderer error = %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFooterEnvelopeValidation(t *testing.T) {
	for _, raw := range []string{
		`{"render":{"generatorCredit":false,"footerAttribution":{"name":"Host"}}}`,
		`{"render":{"copyrightFallback":"author"}}`,
		`{"render":{"footer":"wrong"}}`,
		`{"site":{"footer":"wrong"}}`,
		`{"footer":"wrong"}`,
		`{"copyrightFallback":"title"}`,
		`{"config":{"site":{"copyrightFallback":"title"}}}`,
		`{"config":{"copyrightFallback":"title"}}`,
		`{"config":{"site":{"generatorCredit":false}}}`,
	} {
		t.Run(raw, func(t *testing.T) {
			_, err := Run([]byte(raw))
			var inputErr *InputError
			if !errors.As(err, &inputErr) {
				t.Fatalf("expected input error, got %v", err)
			}
		})
	}
}
