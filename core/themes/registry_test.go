package themes

import (
	"reflect"
	"strings"
	"testing"
)

func TestRegistryContainsBundledThemesAndClassicDefault(t *testing.T) {
	if DefaultPreset != Classic {
		t.Fatalf("default preset = %q, want %q", DefaultPreset, Classic)
	}
	if got := Names(); !reflect.DeepEqual(got, []string{Aurora, Classic, Paper, Terminal}) {
		t.Fatalf("theme names = %v, want [%s %s %s %s]", got, Aurora, Classic, Paper, Terminal)
	}
	definition, ok := Lookup(Classic)
	if !ok {
		t.Fatal("classic theme is not registered")
	}
	if definition.Name != Classic || strings.TrimSpace(definition.CSS) == "" {
		t.Fatalf("classic definition is incomplete: %+v", definition)
	}
	if definition.Defaults.FontHeading == "" || definition.Defaults.FontBody == "" || definition.Defaults.Accent == "" {
		t.Fatalf("classic defaults are incomplete: %+v", definition.Defaults)
	}
	if strings.TrimSpace(BaseCSS) == "" {
		t.Fatal("embedded base stylesheet is empty")
	}

	aurora, ok := Lookup(Aurora)
	if !ok {
		t.Fatal("aurora theme is not registered")
	}
	if !strings.Contains(aurora.CSS, "leafpress Aurora Theme") {
		t.Fatal("aurora stylesheet is missing its visual layer")
	}
	if aurora.Defaults.FontHeading != "Space Grotesk" ||
		aurora.Defaults.Accent != "#16813d" ||
		aurora.Defaults.NavStyle != "glassy" ||
		aurora.Defaults.BackgroundLight == "" || aurora.Defaults.BackgroundDark == "" {
		t.Fatalf("aurora defaults are incomplete: %+v", aurora.Defaults)
	}

	paper, ok := Lookup(Paper)
	if !ok {
		t.Fatal("paper theme is not registered")
	}
	if !strings.Contains(paper.CSS, "leafpress Paper Theme") {
		t.Fatal("paper stylesheet is missing its visual layer")
	}
	if paper.Defaults.FontHeading != "Newsreader" ||
		paper.Defaults.FontBody != "Source Serif 4" ||
		paper.Defaults.FontMono != "IBM Plex Mono" ||
		paper.Defaults.Accent != "#765432" ||
		paper.Defaults.NavStyle != "sticky" ||
		paper.Defaults.NavActiveStyle != "underlined" ||
		paper.Defaults.BackgroundLight != "#faf8f3" ||
		paper.Defaults.BackgroundDark != "#191714" {
		t.Fatalf("paper defaults are incomplete: %+v", paper.Defaults)
	}

	terminal, ok := Lookup(Terminal)
	if !ok {
		t.Fatal("terminal theme is not registered")
	}
	if !strings.Contains(terminal.CSS, "leafpress Terminal Theme") {
		t.Fatal("terminal stylesheet is missing its visual layer")
	}
	if terminal.Defaults.FontHeading != "IBM Plex Mono" ||
		terminal.Defaults.FontBody != "IBM Plex Mono" ||
		terminal.Defaults.FontMono != "IBM Plex Mono" ||
		terminal.Defaults.Accent != "#087f5b" ||
		terminal.Defaults.NavStyle != "sticky" ||
		terminal.Defaults.NavActiveStyle != "base" ||
		terminal.Defaults.BackgroundLight != "#f2f5ef" ||
		terminal.Defaults.BackgroundDark != "#0b100e" {
		t.Fatalf("terminal defaults are incomplete: %+v", terminal.Defaults)
	}
}

func TestLookupRejectsUnknownPreset(t *testing.T) {
	if _, ok := Lookup(""); ok {
		t.Error("empty preset unexpectedly resolved")
	}
	if _, ok := Lookup("unknown"); ok {
		t.Error("unknown preset unexpectedly resolved")
	}
}

func TestTerminalThemeRespectsSemanticFontRoles(t *testing.T) {
	tests := []struct {
		selector string
		role     string
	}{
		{".lp-body", "body"},
		{".lp-nav-title", "heading"},
		{".lp-nav-link", "body"},
		{".lp-title,\n.lp-section-title", "heading"},
		{".lp-content h1,\n.lp-content h2,\n.lp-content h3,\n.lp-content h4,\n.lp-content h5,\n.lp-content h6,\n.lp-section-intro h1,\n.lp-section-intro h2,\n.lp-section-intro h3,\n.lp-section-intro h4,\n.lp-section-intro h5,\n.lp-section-intro h6", "heading"},
		{".lp-content blockquote", "body"},
		{".lp-index-title", "heading"},
		{".lp-search-result-title", "heading"},
		{".lp-link-preview-title", "heading"},
		{".lp-not-found-title", "heading"},
	}

	for _, tt := range tests {
		t.Run(strings.ReplaceAll(tt.selector, "\n", " "), func(t *testing.T) {
			want := "font-family: var(--lp-font-" + tt.role + ")"
			remaining := terminalCSS
			for {
				start := strings.Index(remaining, tt.selector+" {")
				if start == -1 {
					break
				}
				end := strings.Index(remaining[start:], "}")
				if end == -1 {
					t.Fatalf("terminal stylesheet has an unterminated rule for %q", tt.selector)
				}
				rule := remaining[start : start+end]
				if strings.Contains(rule, want) {
					return
				}
				remaining = remaining[start+end+1:]
			}
			t.Errorf("terminal rules for %q do not use the %s font role", tt.selector, tt.role)
		})
	}
}
