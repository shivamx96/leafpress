package assets

import "testing"

// pinnedBuiltins locks the stable identifiers of the built-in set. If this
// test fails, a built-in changed: update the pin AND bump RegistryVersion so
// hosted consumers re-materialize.
var pinnedBuiltins = map[string]struct {
	sha256     string
	publicPath string
}{
	BuiltinFaviconICO: {
		sha256:     "35f29c9301797bf2ce58c8a79fde16c92ec7a4a16fac131cea9b9d230b809370",
		publicPath: "/favicon.ico",
	},
	BuiltinFaviconSVG: {
		sha256:     "796de5b284c308091c61694d4279e5f2f8a76bbdc391e9eb3533087087a85dec",
		publicPath: "/favicon.svg",
	},
	BuiltinFaviconPNG: {
		sha256:     "74ff24e72b334920a961bb8983935a37126e559fdc1d5d7a10b40afb508eed25",
		publicPath: "/favicon-96x96.png",
	},
}

func TestBuiltinRegistryPinned(t *testing.T) {
	all := Builtins()
	if len(all) != len(pinnedBuiltins) {
		t.Fatalf("registry has %d builtins, pin covers %d — update pins and bump RegistryVersion", len(all), len(pinnedBuiltins))
	}
	for _, b := range all {
		pin, ok := pinnedBuiltins[b.Asset.LogicalPath]
		if !ok {
			t.Errorf("unpinned builtin %q — add a pin and bump RegistryVersion", b.Asset.LogicalPath)
			continue
		}
		if b.Asset.SHA256 != pin.sha256 {
			t.Errorf("%s: sha256 = %s, pinned %s — content changed, bump RegistryVersion", b.Asset.LogicalPath, b.Asset.SHA256, pin.sha256)
		}
		if b.Asset.PublicPath != pin.publicPath {
			t.Errorf("%s: publicPath = %q, pinned %q", b.Asset.LogicalPath, b.Asset.PublicPath, pin.publicPath)
		}
	}
}

func TestBuiltinRegistryConsistency(t *testing.T) {
	if err := BuiltinManifest().Validate(); err != nil {
		t.Fatalf("builtin manifest invalid: %v", err)
	}
	for _, b := range Builtins() {
		if !IsBuiltinPath(b.Asset.LogicalPath) {
			t.Errorf("%s: not under %s", b.Asset.LogicalPath, BuiltinPrefix)
		}
		if got := Sum(b.Content); got != b.Asset.SHA256 {
			t.Errorf("%s: content hash %s does not match asset hash %s", b.Asset.LogicalPath, got, b.Asset.SHA256)
		}
		if int64(len(b.Content)) != b.Asset.Size {
			t.Errorf("%s: content length %d does not match size %d", b.Asset.LogicalPath, len(b.Content), b.Asset.Size)
		}
		if len(b.Content) == 0 {
			t.Errorf("%s: empty content", b.Asset.LogicalPath)
		}
	}
}

func TestBuiltinByLogicalPath(t *testing.T) {
	b, ok := BuiltinByLogicalPath(BuiltinFaviconSVG)
	if !ok {
		t.Fatal("favicon.svg builtin not found")
	}
	if b.Asset.ContentType != "image/svg+xml" {
		t.Errorf("favicon.svg content type = %q", b.Asset.ContentType)
	}
	if _, ok := BuiltinByLogicalPath("static/leafpress/nope"); ok {
		t.Error("lookup of unknown path succeeded")
	}
}

func TestRegistryVersionPositive(t *testing.T) {
	if RegistryVersion < 1 {
		t.Fatalf("RegistryVersion = %d", RegistryVersion)
	}
}
