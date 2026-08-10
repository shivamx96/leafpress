package content

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// inlineTag is a parsed #tag in ordinary Markdown text.
type inlineTag struct {
	ast.BaseInline
	Name string
}

// Dump implements ast.Node.
func (n *inlineTag) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Name": n.Name}, nil)
}

var kindInlineTag = ast.NewNodeKind("InlineTag")

// Kind implements ast.Node.
func (n *inlineTag) Kind() ast.NodeKind {
	return kindInlineTag
}

type inlineTagParser struct{}

func (p *inlineTagParser) Trigger() []byte {
	return []byte{'#'}
}

func (p *inlineTagParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	// Linking a tag inside an existing Markdown link would create nested
	// anchors, so link labels remain literal text.
	if pc.IsInLinkLabel() || !isInlineTagBoundary(block.PrecendingCharacter()) {
		return nil
	}

	line, _ := block.PeekLine()
	if len(line) < 2 || line[0] != '#' {
		return nil
	}

	end := 1
	for end < len(line) {
		r, size := utf8.DecodeRune(line[end:])
		if r == utf8.RuneError && size == 1 {
			return nil
		}
		if !isInlineTagRune(r) {
			break
		}
		end += size
	}
	if end == 1 {
		return nil
	}

	// Nested Obsidian tags need a route model of their own. Do not silently
	// turn #parent/child into the partial tag #parent.
	if end < len(line) && line[end] == '/' {
		return nil
	}

	name := string(line[1:end])
	block.Advance(end)
	return &inlineTag{Name: name}
}

func isInlineTagBoundary(previous rune) bool {
	return previous == 0 || unicode.IsSpace(previous) || strings.ContainsRune("([{'\"‘’“”", previous)
}

func isInlineTagRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' || r == '-'
}

type inlineTagHTMLRenderer struct {
	basePath string
}

func (r *inlineTagHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindInlineTag, r.renderInlineTag)
}

func (r *inlineTagHTMLRenderer) renderInlineTag(
	w util.BufWriter, _ []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	tag := node.(*inlineTag).Name
	href := r.basePath + "/tags/" + strings.ToLower(tag) + "/"
	_, _ = w.WriteString(`<a class="lp-tag lp-inline-tag" href="`)
	_, _ = w.Write(util.EscapeHTML(util.URLEscape([]byte(href), false)))
	_, _ = w.WriteString(`">#`)
	_, _ = w.Write(util.EscapeHTML([]byte(tag)))
	_, _ = w.WriteString(`</a>`)
	return ast.WalkContinue, nil
}

type inlineTagExtension struct {
	basePath string
}

func newInlineTagExtension(basePath string) goldmark.Extender {
	return &inlineTagExtension{basePath: basePath}
}

func (e *inlineTagExtension) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&inlineTagParser{}, 150),
	))
	md.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&inlineTagHTMLRenderer{basePath: e.basePath}, 500),
	))
}

var inlineTagExtractor = goldmark.New(
	goldmark.WithExtensions(newInlineTagExtension("")),
)

// ExtractInlineTags returns unique inline tags in source order. Goldmark's
// parser keeps code spans, fenced code, raw HTML tags, link destinations, and
// escaped hashes out of the InlineTag node stream.
func ExtractInlineTags(markdown string) []string {
	document := inlineTagExtractor.Parser().Parse(text.NewReader([]byte(markdown)))
	tags := make([]string, 0)
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && node.Kind() == kindInlineTag {
			tags = append(tags, node.(*inlineTag).Name)
		}
		return ast.WalkContinue, nil
	})
	return NormalizeTags(tags)
}

// MergeTags combines explicit metadata and inline tags, keeping explicit tag
// order and spelling authoritative while removing case-variant duplicates.
func MergeTags(explicit, inline []string) []string {
	merged := make([]string, 0, len(explicit)+len(inline))
	merged = append(merged, explicit...)
	merged = append(merged, inline...)
	return NormalizeTags(merged)
}
