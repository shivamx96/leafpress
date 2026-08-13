package themes

import (
	"reflect"
	"strings"
	"testing"
)

func TestRegistryContainsClassicDefault(t *testing.T) {
	if DefaultPreset != Classic {
		t.Fatalf("default preset = %q, want %q", DefaultPreset, Classic)
	}
	if got := Names(); !reflect.DeepEqual(got, []string{Classic}) {
		t.Fatalf("theme names = %v, want [%s]", got, Classic)
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
}

func TestLookupRejectsUnknownPreset(t *testing.T) {
	if _, ok := Lookup(""); ok {
		t.Error("empty preset unexpectedly resolved")
	}
	if _, ok := Lookup("unknown"); ok {
		t.Error("unknown preset unexpectedly resolved")
	}
}
