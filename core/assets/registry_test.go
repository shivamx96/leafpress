package assets

import "testing"

// pinnedBuiltins locks the stable identifiers of the built-in set. If this
// test fails, a built-in changed — update the pin deliberately. RegistryID
// is content-derived, so it tracks any such change automatically.
var pinnedBuiltins = map[string]struct {
	sha256      string
	outputPath  string
	contentType string
}{
	BuiltinFaviconICO: {
		sha256:      "35f29c9301797bf2ce58c8a79fde16c92ec7a4a16fac131cea9b9d230b809370",
		outputPath:  "favicon.ico",
		contentType: "image/x-icon",
	},
	BuiltinFaviconSVG: {
		sha256:      "796de5b284c308091c61694d4279e5f2f8a76bbdc391e9eb3533087087a85dec",
		outputPath:  "favicon.svg",
		contentType: "image/svg+xml",
	},
	BuiltinFaviconPNG: {
		sha256:      "74ff24e72b334920a961bb8983935a37126e559fdc1d5d7a10b40afb508eed25",
		outputPath:  "favicon-96x96.png",
		contentType: "image/png",
	},
	"static/leafpress/fonts/crimson-pro-italic-latin-ext.woff2":    {sha256: "0ad667be0c601c5a843945ef747d88e12558e264023d1bec92ea1f51257835fb"},
	"static/leafpress/fonts/crimson-pro-italic-latin.woff2":        {sha256: "f6e346888097c3dd29a15620fdd1a727af7df5a71d7dc3c2c84f8682bc2cd716"},
	"static/leafpress/fonts/crimson-pro-normal-latin-ext.woff2":    {sha256: "0ff3122e6772fdfd6835723dfd60f11f244fc6f87720d6811317a4cc580a065c"},
	"static/leafpress/fonts/crimson-pro-normal-latin.woff2":        {sha256: "89374859da8085dbce486eed120a12abcab5a90e725f8b9c41a3e2b32bd010f6"},
	"static/leafpress/fonts/inter-italic-latin-ext.woff2":          {sha256: "e3f94a6c4b2177643513bf5e2317b8458183d064e40031fa04c074136be205e8"},
	"static/leafpress/fonts/inter-italic-latin.woff2":              {sha256: "7ddd9658a57d7c5811b4f6aa43625433d72c05ac83c46a231ef3dfd88813c8a1"},
	"static/leafpress/fonts/inter-normal-latin-ext.woff2":          {sha256: "a28eb6d3ccb534ae0c94ca999371df024aab60b08c3c8a5720ee9e32fa0faaa2"},
	"static/leafpress/fonts/inter-normal-latin.woff2":              {sha256: "c940764593d0fe5d596be327ca7558855e018039fb78509aa21921fd3644c3e4"},
	"static/leafpress/fonts/jetbrains-mono-italic-latin-ext.woff2": {sha256: "3c6c272187a97d8e519af98b6c1255dc02108ef26e3b4bef205e3021128d7dc2"},
	"static/leafpress/fonts/jetbrains-mono-italic-latin.woff2":     {sha256: "e3918da535efc38ebdd4e5d2748de8ecf317e930bb74cd600e8e290ca271fc82"},
	"static/leafpress/fonts/jetbrains-mono-normal-latin-ext.woff2": {sha256: "9c38cb2d0d2d93c1ee6e21fa78db76f13ea7e15e15cc64214c7ca89b6aaa35c4"},
	"static/leafpress/fonts/jetbrains-mono-normal-latin.woff2":     {sha256: "2c32b9b3ee358c119e210f6f5195f9bd34894d78a785ff2e95d60e718e400af4"},
	"static/leafpress/fonts/OFL-crimson-pro.txt":                   {sha256: "1820869bd5baa1c2d88fa87c89eea532cf9442d841008acab720654b7f82823d"},
	"static/leafpress/fonts/OFL-inter.txt":                         {sha256: "5b9321a4298cfeb6b34354164a1c3afc3db114569984c502b9b35d988fd58c57"},
	"static/leafpress/fonts/OFL-jetbrains-mono.txt":                {sha256: "b2fe5e8987594e9ffd1d2ca52a2f5d73eb8335243893c5d6254b5ad69269591d"},
	BuiltinMermaidJS: {
		sha256:      "a43bc1afd446f9c4cc66ac5dd45d02e8d65e26fc5344ec0ef787f88d6ddb6f9e",
		contentType: "text/javascript; charset=utf-8",
	},
	BuiltinMermaidLicense: {
		sha256:      "ec9fb67dcb25eccc416ed56e1aab819222c805a2a4bfe4cb19e7556bf2ffde80",
		contentType: "text/plain; charset=utf-8",
	},
}

func TestBuiltinRegistryPinned(t *testing.T) {
	all := Builtins()
	if len(all) != len(pinnedBuiltins) {
		t.Fatalf("registry has %d builtins, pin covers %d — update pins deliberately", len(all), len(pinnedBuiltins))
	}
	for _, b := range all {
		pin, ok := pinnedBuiltins[b.Asset.LogicalPath]
		if !ok {
			t.Errorf("unpinned builtin %q — add a pin deliberately", b.Asset.LogicalPath)
			continue
		}
		if b.Asset.SHA256 != pin.sha256 {
			t.Errorf("%s: sha256 = %s, pinned %s — content changed", b.Asset.LogicalPath, b.Asset.SHA256, pin.sha256)
		}
		if b.Asset.OutputPath != pin.outputPath {
			t.Errorf("%s: outputPath = %q, pinned %q", b.Asset.LogicalPath, b.Asset.OutputPath, pin.outputPath)
		}
		if pin.contentType != "" && b.Asset.ContentType != pin.contentType {
			t.Errorf("%s: contentType = %q, pinned %q", b.Asset.LogicalPath, b.Asset.ContentType, pin.contentType)
		}
	}
}

func TestValidateUserAssetPolicy(t *testing.T) {
	valid := Asset{
		LogicalPath: "static/fonts/my.woff2",
		ContentType: "font/woff2",
		SHA256:      Sum([]byte("x")),
		Size:        1,
	}
	if err := ValidateUserAsset(valid); err != nil {
		t.Fatalf("plain user asset rejected: %v", err)
	}

	override := valid
	override.LogicalPath = "static/my-favicon.ico"
	override.ContentType = "image/x-icon"
	override.OutputPath = "favicon.ico"
	if err := ValidateUserAsset(override); err != nil {
		t.Fatalf("favicon override rejected: %v", err)
	}

	reserved := valid
	reserved.LogicalPath = BuiltinPrefix + "evil.woff2"
	if err := ValidateUserAsset(reserved); err == nil {
		t.Error("reserved-namespace logical path accepted")
	}

	arbitrary := valid
	arbitrary.OutputPath = "style.css"
	if err := ValidateUserAsset(arbitrary); err == nil {
		t.Error("non-override outputPath accepted")
	}
}

func TestRootBuiltinsAreTheOverridableSet(t *testing.T) {
	roots := RootBuiltins()
	overridable := OverridableOutputPaths()
	if len(roots) != len(overridable) || len(roots) == 0 {
		t.Fatalf("RootBuiltins (%d) and OverridableOutputPaths (%d) must describe the same non-empty set", len(roots), len(overridable))
	}
	for _, b := range roots {
		if !overridable[b.Asset.OutputPath] {
			t.Errorf("%s missing from overridable set", b.Asset.OutputPath)
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
		content := b.Content()
		if got := Sum(content); got != b.Asset.SHA256 {
			t.Errorf("%s: content hash %s does not match asset hash %s", b.Asset.LogicalPath, got, b.Asset.SHA256)
		}
		if int64(len(content)) != b.Asset.Size {
			t.Errorf("%s: content length %d does not match size %d", b.Asset.LogicalPath, len(content), b.Asset.Size)
		}
		if len(content) == 0 {
			t.Errorf("%s: empty content", b.Asset.LogicalPath)
		}
	}
}

func TestBuiltinContentIsDefensiveCopy(t *testing.T) {
	b, ok := BuiltinByLogicalPath(BuiltinFaviconICO)
	if !ok {
		t.Fatal("favicon.ico builtin not found")
	}
	first := b.Content()
	for i := range first {
		first[i] = 0xFF
	}
	second := b.Content()
	if got := Sum(second); got != b.Asset.SHA256 {
		t.Fatal("mutating a returned Content() slice corrupted the registry")
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

func TestRegistryIDDerivedFromManifest(t *testing.T) {
	id := RegistryID()
	if len(id) != 64 {
		t.Fatalf("RegistryID %q is not a sha256 hex digest", id)
	}
	if id != RegistryID() {
		t.Error("RegistryID not stable across calls")
	}
	// Derived from the canonical manifest: any change to a built-in changes
	// the manifest and therefore the ID — no version integer to forget.
	if id == Sum([]byte("[]")) {
		t.Error("RegistryID suspiciously matches an empty manifest")
	}
}
