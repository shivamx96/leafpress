package content

import (
	"reflect"
	"strings"
	"testing"
)

func TestExtractInlineTags(t *testing.T) {
	markdown := "Working on #LeafPress, #go-lang, #side_project, and #日本語. Repeat #leafpress."
	want := []string{"LeafPress", "go-lang", "side_project", "日本語"}
	if got := ExtractInlineTags(markdown); !reflect.DeepEqual(got, want) {
		t.Fatalf("ExtractInlineTags() = %v, want %v", got, want)
	}
}

func TestExtractInlineTagsIgnoresProtectedAndAmbiguousHashes(t *testing.T) {
	markdown := strings.Join([]string{
		"# Heading",
		"Visible (#kept).",
		"`#inline-code`",
		"```markdown",
		"#fenced-code",
		"```",
		`\#escaped`,
		"word#joined",
		"https://example.com/#fragment",
		"[linked #label](https://example.com/)",
		"[[note|#wiki-label]]",
		`<span data-tag="#attribute">text</span>`,
		"#nested/tag",
	}, "\n\n")
	want := []string{"kept"}
	if got := ExtractInlineTags(markdown); !reflect.DeepEqual(got, want) {
		t.Fatalf("ExtractInlineTags() = %v, want %v", got, want)
	}
}

func TestMergeTagsPreservesFrontmatterSpelling(t *testing.T) {
	frontmatter := []string{"Systems", "notes"}
	inline := []string{"systems", "LeafPress", "NOTES", "go-lang"}
	want := []string{"Systems", "notes", "LeafPress", "go-lang"}
	if got := MergeTags(frontmatter, inline); !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeTags() = %v, want %v", got, want)
	}
}

func TestRenderInlineTagsAsLinks(t *testing.T) {
	r := NewRenderer(nil, false, "/garden")
	html, warnings := r.Render("Working on #LeafPress and `#literal`.")
	if len(warnings) != 0 {
		t.Fatalf("Render() warnings = %v", warnings)
	}
	want := `<a class="lp-tag lp-inline-tag" href="/garden/tags/leafpress/">#LeafPress</a>`
	if !strings.Contains(html, want) {
		t.Fatalf("rendered HTML missing inline tag link %q:\n%s", want, html)
	}
	if strings.Contains(html, `href="/garden/tags/literal/"`) {
		t.Fatalf("inline code tag became a link:\n%s", html)
	}
}

func TestRenderInlineTagsInEscapedHTMLMode(t *testing.T) {
	r := NewRenderer(nil, false, "")
	r.SetEscapeRawHTML(true)
	html, _ := r.Render(`<span data-tag="#attribute">raw</span> #safe`)
	if strings.Contains(html, `/tags/attribute/`) {
		t.Fatalf("tag inside raw HTML attribute became a link:\n%s", html)
	}
	if !strings.Contains(html, `href="/tags/safe/"`) {
		t.Fatalf("ordinary inline tag did not become a link:\n%s", html)
	}
}

func TestRenderDoesNotLinkTagsInsideWikiLinkLabels(t *testing.T) {
	target := &Page{Slug: "note", Permalink: "/note/"}
	r := NewRenderer(NewLinkResolver([]*Page{target}), true, "")
	html, _ := r.Render("See [[note|topic #wiki-label]].")
	if strings.Contains(html, "lp-inline-tag") {
		t.Fatalf("wiki-link label contains a nested inline tag link:\n%s", html)
	}
	if !strings.Contains(html, `<a class="lp-wikilink" href="/note/">topic #wiki-label</a>`) {
		t.Fatalf("wiki-link label did not render literally:\n%s", html)
	}
}
