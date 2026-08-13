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
	if got := Names(); !reflect.DeepEqual(got, []string{Aurora, Classic}) {
		t.Fatalf("theme names = %v, want [%s %s]", got, Aurora, Classic)
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
}

func TestLookupRejectsUnknownPreset(t *testing.T) {
	if _, ok := Lookup(""); ok {
		t.Error("empty preset unexpectedly resolved")
	}
	if _, ok := Lookup("unknown"); ok {
		t.Error("unknown preset unexpectedly resolved")
	}
}
