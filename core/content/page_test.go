package content

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPageTextTruncationPreservesUTF8(t *testing.T) {
	page := &Page{HTMLContent: "<p>" + strings.Repeat("é", 5001) + "</p>"}
	plain := page.PlainContent()
	if !utf8.ValidString(plain) || utf8.RuneCountInString(plain) != 5000 {
		t.Fatalf("PlainContent produced invalid or wrong-length UTF-8: valid=%v runes=%d", utf8.ValidString(plain), utf8.RuneCountInString(plain))
	}

	page.HTMLContent = "<p>" + strings.Repeat("界", 156) + "</p>"
	description := page.SEODescription()
	if !utf8.ValidString(description) || utf8.RuneCountInString(description) != 158 {
		t.Fatalf("SEODescription produced invalid or wrong-length UTF-8: %q", description)
	}
}

func TestNormalizeTagsDeduplicatesCaseVariants(t *testing.T) {
	got := NormalizeTags([]string{"Go", "systems", "go", "SYSTEMS", "notes"})
	want := []string{"Go", "systems", "notes"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("NormalizeTags = %v, want %v", got, want)
	}
}
