package assets

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

// Logical paths of the built-in assets Leafpress ships. They are stable
// identifiers: consumers may persist them.
const (
	BuiltinFaviconICO = BuiltinPrefix + "favicon.ico"
	BuiltinFaviconSVG = BuiltinPrefix + "favicon.svg"
	BuiltinFaviconPNG = BuiltinPrefix + "favicon-96x96.png"
)

//go:embed builtin/favicon.ico
var faviconICO []byte

//go:embed builtin/favicon.svg
var faviconSVG []byte

//go:embed builtin/favicon-96x96.png
var faviconPNG []byte

//go:embed builtin/fonts/*.woff2 builtin/fonts/OFL-*.txt
var builtinFontsFS embed.FS

// Builtin is a Leafpress-owned asset together with its embedded content.
type Builtin struct {
	Asset   Asset
	content []byte
}

// Content returns a copy of the embedded bytes. Callers get their own slice:
// nothing they do to it can drift the registry's recorded hash or size.
func (b Builtin) Content() []byte {
	out := make([]byte, len(b.content))
	copy(out, b.content)
	return out
}

var (
	builtins         = mustBuiltins()
	builtinManifest  = mustBuiltinManifest(builtins)
	builtinRegistry  = mustRegistryID(builtinManifest)
	builtinByLogical = indexBuiltins(builtins)
)

func mustBuiltins() []Builtin {
	entries := []struct {
		logicalPath string
		contentType string
		outputPath  string
		content     []byte
	}{
		{BuiltinFaviconICO, "image/x-icon", "favicon.ico", faviconICO},
		{BuiltinFaviconSVG, "image/svg+xml", "favicon.svg", faviconSVG},
		{BuiltinFaviconPNG, "image/png", "favicon-96x96.png", faviconPNG},
	}

	out := make([]Builtin, 0, len(entries)+len(builtinFontFaces))
	for _, e := range entries {
		out = append(out, Builtin{
			Asset: Asset{
				LogicalPath: e.logicalPath,
				ContentType: e.contentType,
				SHA256:      Sum(e.content),
				Size:        int64(len(e.content)),
				OutputPath:  e.outputPath,
			},
			content: e.content,
		})
	}
	readFont := func(logicalPath string) []byte {
		// Embedded paths mirror logical paths minus the static/leafpress/
		// prefix, rooted at builtin/.
		embedPath := "builtin/" + strings.TrimPrefix(logicalPath, BuiltinPrefix)
		data, err := builtinFontsFS.ReadFile(embedPath)
		if err != nil {
			panic(fmt.Sprintf("assets: built-in font asset %s not embedded: %v", logicalPath, err))
		}
		return data
	}
	for _, face := range builtinFontFaces {
		data := readFont(face.LogicalPath)
		out = append(out, Builtin{
			Asset: Asset{
				LogicalPath: face.LogicalPath,
				ContentType: "font/woff2",
				SHA256:      Sum(data),
				Size:        int64(len(data)),
			},
			content: data,
		})
	}
	for _, family := range BuiltinFontFamilies() {
		licensePath, ok := BuiltinFontLicense(family)
		if !ok {
			panic(fmt.Sprintf("assets: curated family %q has no OFL license asset", family))
		}
		data := readFont(licensePath)
		out = append(out, Builtin{
			Asset: Asset{
				LogicalPath: licensePath,
				ContentType: "text/plain; charset=utf-8",
				SHA256:      Sum(data),
				Size:        int64(len(data)),
			},
			content: data,
		})
	}
	return out
}

func mustBuiltinManifest(list []Builtin) Manifest {
	assets := make([]Asset, 0, len(list))
	for _, b := range list {
		assets = append(assets, b.Asset)
	}
	m, err := NewManifest(assets...)
	if err != nil {
		panic(fmt.Sprintf("assets: built-in registry is invalid: %v", err))
	}
	return m
}

// mustRegistryID derives the registry identity from the canonical manifest:
// the SHA-256 of its JSON encoding. It changes exactly when any built-in is
// added, removed, or altered — there is no manually-bumped version to forget.
func mustRegistryID(m Manifest) string {
	encoded, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("assets: cannot encode built-in manifest: %v", err))
	}
	return Sum(encoded)
}

func indexBuiltins(list []Builtin) map[string]Builtin {
	index := make(map[string]Builtin, len(list))
	for _, b := range list {
		index[b.Asset.LogicalPath] = b
	}
	return index
}

// RegistryID identifies the current built-in asset set by content: it
// changes exactly when any built-in changes. It is a change signal for
// observability and prefetching only — synchronization is hash-driven per
// manifest entry, because any one render's manifest is a theme-dependent
// subset of the registry.
func RegistryID() string {
	return builtinRegistry
}

// Builtins returns every built-in asset in deterministic order.
func Builtins() []Builtin {
	out := make([]Builtin, len(builtins))
	copy(out, builtins)
	return out
}

// BuiltinByLogicalPath looks up a built-in by its stable logical path.
func BuiltinByLogicalPath(logicalPath string) (Builtin, bool) {
	b, ok := builtinByLogical[logicalPath]
	return b, ok
}

// OverridableOutputPaths returns the effective output paths users and
// callers may override: exactly the built-ins with an explicit output
// location (the root favicons). This is the shared policy list — the
// renderer's caller-asset validation and any CLI manifest assembly must use
// it rather than private copies.
func OverridableOutputPaths() map[string]bool {
	out := make(map[string]bool)
	for _, b := range builtins {
		if b.Asset.OutputPath != "" {
			out[b.Asset.OutputPath] = true
		}
	}
	return out
}

// ValidateUserAsset enforces the policy for user/caller-declared assets on
// top of the shape contract (docs/07_ASSET_ARCHITECTURE.md §5): logical
// paths must stay out of the reserved built-in namespace, and an explicit
// output path is only legal as a supported built-in override — ordinary
// assets serve at their logical path. Shared here so the CLI and the
// renderer cannot drift on what a legal user asset is.
func ValidateUserAsset(a Asset) error {
	if err := a.Validate(); err != nil {
		return err
	}
	if IsBuiltinPath(a.LogicalPath) {
		return fmt.Errorf("logical path %q is in the reserved built-in namespace", a.LogicalPath)
	}
	if a.OutputPath != "" && !OverridableOutputPaths()[a.OutputPath] {
		return fmt.Errorf("outputPath %q is not a built-in override; ordinary assets serve at their logical path (leave outputPath empty)", a.OutputPath)
	}
	return nil
}

// RootBuiltins returns the built-ins served at an explicit root location
// (the favicons), in registry order. CLI materialization and renderer
// selection share this list.
func RootBuiltins() []Builtin {
	var out []Builtin
	for _, b := range builtins {
		if b.Asset.OutputPath != "" {
			out = append(out, b)
		}
	}
	return out
}

// BuiltinManifest returns the metadata-only canonical manifest of the
// built-in set.
func BuiltinManifest() Manifest {
	out := make(Manifest, len(builtinManifest))
	copy(out, builtinManifest)
	return out
}
