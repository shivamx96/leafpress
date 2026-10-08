package site

import (
	"github.com/shivamx96/leafpress/core/templates"
	"testing"
)

func TestFaviconEscapingRoundTrip(t *testing.T) {
	raw := templates.SiteData{Favicon: `/static/a&"b.png`, Image: `/static/a&"b.jpg`}
	safe := SafeSiteData(raw)
	if safe.Favicon != `/static/a&amp;&#34;b.png` {
		t.Fatalf("favicon escaping = %q", safe.Favicon)
	}
	if got := RawSiteData(safe, ""); got.Favicon != raw.Favicon || got.Image != raw.Image {
		t.Fatalf("round trip = %+v", got)
	}
}
