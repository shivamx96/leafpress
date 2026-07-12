package content

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// rawHTMLEscaper is a goldmark NodeRenderer that overrides the default
// rendering of raw HTML (both inline ast.RawHTML and block-level
// ast.HTMLBlock). Instead of passing the author's HTML through (unsafe
// mode) or dropping it ("<!-- raw HTML omitted -->", safe mode), it writes
// the original source segment HTML-escaped, so the reader sees the literal
// characters the author typed (e.g. <script> renders as &lt;script&gt;).
type rawHTMLEscaper struct{}

// RegisterFuncs registers this renderer for the raw HTML node kinds. It is
// registered with a higher priority (lower number) than goldmark's default
// html.Renderer (1000), so it wins for exactly these two kinds.
func (r *rawHTMLEscaper) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
}

// renderRawHTML mirrors goldmark's html.Renderer.renderRawHTML unsafe path,
// but writes each source segment through util.EscapeHTML.
func (r *rawHTMLEscaper) renderRawHTML(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	n := node.(*ast.RawHTML)
	l := n.Segments.Len()
	for i := 0; i < l; i++ {
		segment := n.Segments.At(i)
		_, _ = w.Write(util.EscapeHTML(segment.Value(source)))
	}
	return ast.WalkSkipChildren, nil
}

// renderHTMLBlock mirrors goldmark's html.Renderer.renderHTMLBlock unsafe
// path, but writes each line (and the closure line, for block types that
// have one) through util.EscapeHTML.
func (r *rawHTMLEscaper) renderHTMLBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.HTMLBlock)
	if entering {
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			_, _ = w.Write(util.EscapeHTML(line.Value(source)))
		}
	} else if n.HasClosure() {
		_, _ = w.Write(util.EscapeHTML(n.ClosureLine.Value(source)))
	}
	return ast.WalkContinue, nil
}

// trustedChunks protects renderer-generated HTML (wikilink anchors, callout
// wrappers, media embeds) from the raw-HTML escaper. Each trusted chunk is
// swapped for an unguessable plain-text token before markdown conversion and
// restored afterwards. Tokens are alphanumeric so goldmark passes them
// through untouched, and they carry a per-render random nonce so authors
// cannot forge them from markdown content.
type trustedChunks struct {
	prefix string
	chunks []string
}

// newTrustedChunks creates a collector with a fresh random token prefix.
func newTrustedChunks() *trustedChunks {
	var b [9]byte
	_, _ = rand.Read(b[:]) // never fails (crypto/rand)
	return &trustedChunks{prefix: "lp" + hex.EncodeToString(b[:]) + "t"}
}

// wrap stores a trusted HTML chunk and returns its placeholder token.
func (t *trustedChunks) wrap(html string) string {
	t.chunks = append(t.chunks, html)
	return t.token(len(t.chunks) - 1)
}

// token builds the placeholder for chunk i. The trailing "z" terminator
// keeps token i from being a prefix of token i0..in.
func (t *trustedChunks) token(i int) string {
	return t.prefix + strconv.Itoa(i) + "z"
}

// restore replaces every placeholder token in rendered HTML with its trusted
// chunk. Block-level tokens come back from goldmark wrapped in a paragraph;
// that wrapper is removed so block chunks land at the top level, matching
// how the same HTML would have been emitted by the unsafe path.
func (t *trustedChunks) restore(html string) string {
	if len(t.chunks) == 0 {
		return html
	}
	pairs := make([]string, 0, len(t.chunks)*4)
	for i, chunk := range t.chunks {
		token := t.token(i)
		pairs = append(pairs, "<p>"+token+"</p>", chunk, token, chunk)
	}
	return strings.NewReplacer(pairs...).Replace(html)
}
