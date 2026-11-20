package render

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestHierarchyFromOnlyDeepIndex(t *testing.T) {
	out := runJSON(t, `{"content":{"pages":[{"slug":"clients/acme/empty","title":"Empty","isIndex":true,"markdown":"Empty intro"}]}}`)
	want := []string{"clients", "clients/acme"}
	var got []string
	for _, s := range out.Sections {
		got = append(got, s.Slug)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sections=%v, want %v", got, want)
	}
	if !strings.Contains(out.Index, `class="lp-index-link" href="/clients/"`) {
		t.Fatal("missing top-level ancestor")
	}
	if !strings.Contains(out.Sections[1].HTML, `class="lp-index-link" href="/clients/acme/empty/"`) {
		t.Fatal("empty index missing from parent")
	}
	if !strings.Contains(pageHTML(t, out, "clients/acme/empty"), "Empty intro") {
		t.Fatal("missing explicit intro")
	}
}

func TestHierarchySort(t *testing.T) {
	for _, tt := range []struct {
		sort string
		want []string
	}{
		{"date", []string{"/clients/deep/", "/clients/early/", "/clients/old/"}},
		{"title", []string{"/clients/early/", "/clients/deep/", "/clients/old/"}},
		{"growth", []string{"/clients/early/", "/clients/old/", "/clients/deep/"}},
	} {
		t.Run(tt.sort, func(t *testing.T) {
			in := Input{Content: Content{Pages: []InputPage{
				{Slug: "clients", IsIndex: true, Sort: tt.sort},
				{Slug: "clients/old", Title: "Z", CreatedAt: "2026-01-01T00:00:00Z", Growth: "evergreen"},
				{Slug: "clients/early", Title: "A", CreatedAt: "2026-02-01T00:00:00Z", Growth: "seedling"},
				{Slug: "clients/deep/nested/plan", Title: "Plan", UpdatedAt: "2026-03-01T00:00:00Z"},
			}}}
			raw, _ := json.Marshal(in)
			out, err := Run(raw)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, m := range regexp.MustCompile(`class="lp-index-link" href="([^"]+)"`).FindAllStringSubmatch(pageHTML(t, out, "clients"), -1) {
				got = append(got, m[1])
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("links=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestHierarchyRejectsAncestorDocument(t *testing.T) {
	_, err := Run([]byte(`{"content":{"pages":[{"slug":"clients"},{"slug":"clients/acme/projects/plan"}]}}`))
	if err == nil || !strings.Contains(err.Error(), `output route "/clients/"`) {
		t.Fatalf("error=%v", err)
	}
}
