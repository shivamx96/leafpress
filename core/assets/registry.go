package assets

import (
	_ "embed"
	"fmt"
)

// RegistryVersion identifies the current built-in asset set. Bump it whenever
// any built-in is added, removed, or changes content, so hosted consumers can
// detect "already materialized" cheaply instead of re-uploading bytes.
const RegistryVersion = 1

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

// Builtin is a Leafpress-owned asset together with its embedded content.
// Content aliases the embedded data and must be treated as read-only.
type Builtin struct {
	Asset   Asset
	Content []byte
}

var builtins = mustBuiltins()

func mustBuiltins() []Builtin {
	entries := []struct {
		logicalPath string
		contentType string
		publicPath  string
		content     []byte
	}{
		{BuiltinFaviconICO, "image/x-icon", "/favicon.ico", faviconICO},
		{BuiltinFaviconSVG, "image/svg+xml", "/favicon.svg", faviconSVG},
		{BuiltinFaviconPNG, "image/png", "/favicon-96x96.png", faviconPNG},
	}

	out := make([]Builtin, 0, len(entries))
	for _, e := range entries {
		out = append(out, Builtin{
			Asset: Asset{
				LogicalPath: e.logicalPath,
				ContentType: e.contentType,
				SHA256:      Sum(e.content),
				Size:        int64(len(e.content)),
				PublicPath:  e.publicPath,
			},
			Content: e.content,
		})
	}
	if err := builtinManifest(out).Validate(); err != nil {
		panic(fmt.Sprintf("assets: built-in registry is invalid: %v", err))
	}
	return out
}

func builtinManifest(list []Builtin) Manifest {
	m := make(Manifest, 0, len(list))
	for _, b := range list {
		m = append(m, b.Asset)
	}
	m.Sort()
	return m
}

// Builtins returns every built-in asset in deterministic order.
func Builtins() []Builtin {
	out := make([]Builtin, len(builtins))
	copy(out, builtins)
	return out
}

// BuiltinByLogicalPath looks up a built-in by its stable logical path.
func BuiltinByLogicalPath(logicalPath string) (Builtin, bool) {
	for _, b := range builtins {
		if b.Asset.LogicalPath == logicalPath {
			return b, true
		}
	}
	return Builtin{}, false
}

// BuiltinManifest returns the metadata-only manifest of the built-in set,
// sorted by logical path.
func BuiltinManifest() Manifest {
	return builtinManifest(builtins)
}
