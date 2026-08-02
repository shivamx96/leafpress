// Package render implements the leafpress-render bridge: a pure
// stdin→stdout JSON transform that renders a set of published pages
// (a "garden") into full HTML documents, an index page, and theme CSS.
// It performs no filesystem, network, or database access.
//
// The input is one envelope: a shared `config` object (identical to the CLI's
// leafpress.json), a `render` block of host-only concerns, the `content` to
// render, and `options`. See docs/05_RENDERER_CONTRACT.md.
package render

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/shivamx96/leafpress/core/assets"
	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
	sitegen "github.com/shivamx96/leafpress/core/site"
	"github.com/shivamx96/leafpress/core/templates"
)

// Input is the top-level JSON object read from stdin.
type Input struct {
	// ContractVersion is optional; 0 means latest. An unknown version is
	// rejected rather than guessed at.
	ContractVersion int `json:"contractVersion"`
	// Config is the shared configuration object — the same schema the CLI
	// reads from leafpress.json. Absent/empty renders the default site.
	Config json.RawMessage `json:"config"`
	// Render holds the only host-only concerns: the garden identity slug and
	// optional white-label footer attribution.
	Render RenderOpts `json:"render"`
	// Content carries the in-memory equivalents of the filesystem inputs a
	// CLI build reads: the pages, the stylesheet, and declared user assets.
	Content Content `json:"content"`
	// Options carries render toggles.
	Options Options `json:"options"`
}

// RenderOpts holds renderer/hosting concerns that a filesystem build never
// needs.
type RenderOpts struct {
	// Slug is the hosted garden's identity/routing key. It has no natural
	// default; when omitted it defaults to "garden" with a warning.
	Slug string `json:"slug"`
	// FooterAttribution is renderer-only white-label branding, deliberately
	// structured instead of accepting raw HTML or script content.
	FooterAttribution *FooterAttribution `json:"footerAttribution,omitempty"`
}

// Content is the renderable input a CLI build would read from disk.
type Content struct {
	Pages []InputPage `json:"pages"`
	// StyleCSS is the in-memory counterpart of the CLI project's style.css.
	StyleCSS string `json:"styleCSS"`
	// Assets declares the user assets the caller will serve alongside the
	// rendered site (custom font files under static/fonts/, and in the future
	// other referenced static files). Entries are validated with the shared
	// manifest rules and merged into the output manifest; an entry whose
	// effective output path collides with a built-in replaces it (the
	// favicon-override rule).
	Assets []assets.Asset `json:"assets"`
}

// Options carries render toggles.
type Options struct {
	// EmitAssets requests base64 artifacts for the built-in assets the
	// rendered site requires. The asset manifest is always emitted; bytes are
	// opt-in so routine renders stay small. Synchronization is hash-driven per
	// manifest entry — the registry ID alone is never a valid skip signal,
	// because the manifest is a theme-dependent subset.
	EmitAssets bool `json:"emitAssets"`
}

// FooterAttribution is renderer-only host branding. It is deliberately
// structured instead of accepting raw HTML or script content.
type FooterAttribution struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// InputPage is a single published page. Slugs may carry path segments
// ("essays/my-post"); section membership derives from the slug's directory,
// exactly like the CLI build.
type InputPage struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"` // optional; defaults to Slug
	Markdown    string   `json:"markdown"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"createdAt"` // optional RFC3339
	UpdatedAt   string   `json:"updatedAt"` // optional RFC3339
	Description string   `json:"description"`
	Growth      string   `json:"growth"`
	TOC         *bool    `json:"toc"`
	Image       string   `json:"image"`
	ReadingTime *int     `json:"readingTime"`
	// IsIndex marks a section home (the CLI's _index.md): Slug is the
	// section path itself, Markdown becomes the intro above the child
	// listing. An IsIndex page with slug "" is the garden home.
	IsIndex bool `json:"isIndex"`
	// Sort orders the child listing of an index page: date (default) |
	// title | growth. Mirrors the _index.md `sort` frontmatter key.
	Sort string `json:"sort"`
	// ShowList toggles the child listing of an index page (default true).
	// Mirrors the _index.md `showList` frontmatter key.
	ShowList *bool `json:"showList"`
}

// Output is the top-level JSON object written to stdout.
type Output struct {
	Pages    []OutputPage    `json:"pages"`
	Index    string          `json:"index"`
	Sections []OutputSection `json:"sections"`
	Tags     OutputTags      `json:"tags"`
	CSS      string          `json:"css"`
	// AssetManifest is the combined manifest of every asset the rendered
	// site requires: referenced built-ins plus caller-declared assets, with
	// caller entries replacing built-ins on output-path collision. Metadata
	// only, never bytes. Hosted consumers materialize each entry through
	// their own storage using the content hash; built-in entries also
	// appear as base64 artifacts when the input sets options.emitAssets.
	AssetManifest assets.Manifest `json:"assetManifest"`
	// AssetRegistryID identifies the built-in registry snapshot the manifest
	// came from (content-derived). It is a change signal only — the manifest
	// is a theme-dependent subset, so synchronization stays hash-driven per
	// entry, never keyed on this ID.
	AssetRegistryID string           `json:"assetRegistryId"`
	Artifacts       []OutputArtifact `json:"artifacts"`
	Warnings        []string         `json:"warnings"`
}

// OutputArtifact is a filesystem-free generated site file. Path uses the
// exact CLI filename so consumers can store/serve artifacts generically.
type OutputArtifact struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	// Encoding says how Content encodes the file bytes and is authoritative
	// (never sniff): generated site artifacts are always "utf8"; asset
	// artifacts emitted under options.emitAssets are always "base64" regardless
	// of MIME type (OFL license texts included).
	Encoding string `json:"encoding"`
}

// OutputPage is a rendered page document. Index pages appear here too,
// rendered as their section's home.
type OutputPage struct {
	Slug string `json:"slug"`
	HTML string `json:"html"`
}

// OutputSection is an auto-generated home for a section that has no index
// page (the CLI's auto-index), served at {baseUrl}/<slug>/.
type OutputSection struct {
	Slug string `json:"slug"`
	HTML string `json:"html"`
}

// OutputTags holds the rendered tag index and per-tag pages. When no page
// carries any tag, Index is "" and Pages is empty (mirroring the CLI, which
// skips the tags section entirely in that case).
type OutputTags struct {
	Index string          `json:"index"`
	Pages []OutputTagPage `json:"pages"`
}

// OutputTagPage is a rendered page listing everything under one tag,
// served at {baseUrl}/tags/<tag>/.
type OutputTagPage struct {
	Tag  string `json:"tag"`
	HTML string `json:"html"`
}

// InputError marks failures caused by invalid input (exit code 1),
// as opposed to internal render failures (exit code 2).
type InputError struct {
	msg string
}

func (e *InputError) Error() string { return e.msg }

func inputErrorf(format string, args ...any) error {
	return &InputError{msg: fmt.Sprintf(format, args...)}
}

var (
	// unsafeSlugChars are characters that would break hrefs/attributes when a
	// slug is interpolated into template output.
	unsafeSlugChars = "\"'<>\\ \t\r\n"
	// tagRegex restricts tags to letters/digits/underscore/hyphen (the
	// hosted tag shape); tags are interpolated into hrefs and text unescaped.
	tagRegex = regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)
)

// Run decodes raw JSON input and renders it. Errors of type *InputError
// indicate invalid input; any other error is an internal failure.
func Run(raw []byte) (*Output, error) {
	// Reject unknown/misplaced fields so a v1 payload (or a typo) fails loudly
	// instead of silently rendering an empty default site. This enforces the
	// contract's no-silent-loss invariant across the envelope; the nested
	// config object is validated with the same strictness in config.Parse.
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var in Input
	if err := dec.Decode(&in); err != nil {
		return nil, inputErrorf("invalid input JSON: %v", err)
	}
	return Render(&in)
}

// Render validates the input and produces rendered output.
func Render(in *Input) (*Output, error) {
	if in.ContractVersion != 0 && in.ContractVersion != config.ContractVersionLatest {
		return nil, inputErrorf("unsupported contractVersion %d (this build supports %d)", in.ContractVersion, config.ContractVersionLatest)
	}

	warnings := []string{}

	slug := strings.Trim(in.Render.Slug, "/")
	if slug == "" {
		slug = "garden"
		warnings = append(warnings, `render.slug not provided; defaulting to "garden" (hosts should supply an explicit slug)`)
	}
	if strings.ContainsAny(slug, unsafeSlugChars) || hasDotSegment(slug) {
		return nil, inputErrorf("render.slug is invalid: %q", in.Render.Slug)
	}

	if err := validateFooterAttribution(in.Render.FooterAttribution); err != nil {
		return nil, err
	}

	cfg, site, err := resolveConfig(in)
	if err != nil {
		return nil, err
	}
	title := site.Title
	basePath := site.BasePath

	pages, err := buildPages(in.Content.Pages)
	if err != nil {
		return nil, err
	}

	// Resolve wikilinks over exactly these pages; unresolved links degrade
	// to plain text (anything unresolved is private by design).
	resolver := content.NewLinkResolver(pages)
	// NewLinkResolver already registered each page's stored title — which
	// buildPages HTML-escaped at the input boundary. Authors link by the
	// raw display title ([[Foo & Bar]], not [[Foo &amp; Bar]]), so register
	// the raw input titles too; for titles without special characters this
	// second registration is a no-op. buildPages preserves input order, so
	// zip by index.
	for i, ip := range in.Content.Pages {
		resolver.AddAlias(ip.Title, pages[i])
	}
	if cfg.Features.Backlinks {
		content.BuildBacklinks(pages, resolver)
	}
	renderer := content.NewRenderer(resolver, cfg.Features.Wikilinks, basePath)
	renderer.SetPlainBrokenLinks(true)
	// Raw HTML in author markdown renders as visibly escaped text; this
	// bridge serves multi-tenant user content to third parties.
	renderer.SetEscapeRawHTML(true)

	for _, p := range pages {
		rendered, warns := renderer.Render(p.RawContent)
		p.HTMLContent = rendered
		p.WordCount = content.CountWords(rendered)
		p.ImageCount = content.CountImages(rendered)
		p.ReadingTime = content.CalculateReadingTime(p.WordCount, p.ImageCount)
		if p.ReadingTimeOverride != nil {
			p.ReadingTime = *p.ReadingTimeOverride
		}
		// Pin the auto-generated SEO description to an escaped value.
		// Page.SEODescription derives it from HTMLContent via PlainContent,
		// which runs html.UnescapeString — that would resurrect quotes and
		// angle brackets from body text right before the value is placed in
		// meta attributes by text/template.
		if p.Description == "" {
			p.Description = html.EscapeString(p.SEODescription())
		}
		for _, w := range warns {
			warnings = append(warnings, fmt.Sprintf("page %q: %s", p.Slug, w))
		}
	}

	tmpl, err := templates.New()
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	// Section membership derives from slug paths, exactly like the CLI:
	// a page "essays/my-post" is a direct child of section "essays", and
	// an index page's slug *is* its section path ("" = the garden home).
	children, indexBySection := sitegen.GroupSections(pages)

	// Root notes and sections are the garden's public top level. Use that
	// same projection for both the home listing and site-wide navigation.
	homeEntries := sitegen.RootEntries(children, indexBySection)
	site.Nav = sitegen.BuildNavigation(pages, cfg.Navigation)

	outPages := make([]OutputPage, 0, len(pages))
	for _, p := range pages {
		var buf bytes.Buffer
		if p.IsIndex {
			if p.Slug == "" {
				continue // the garden home renders into Output.Index below
			}
			if err := renderSectionHome(&buf, tmpl, site, p, children[p.Slug]); err != nil {
				return nil, fmt.Errorf("failed to render section %q: %w", p.Slug, err)
			}
		} else {
			var toc []templates.TOCItem
			htmlContent := p.HTMLContent
			showTOC := site.TOC
			if p.TOC != nil {
				showTOC = *p.TOC
			}
			if showTOC {
				htmlContent, toc = templates.ExtractTOC(p.HTMLContent)
			}
			if err := tmpl.RenderPage(&buf, templates.PageData{
				Site:        site,
				Page:        p,
				Content:     htmlContent,
				TOC:         toc,
				CurrentPath: p.Permalink,
			}); err != nil {
				return nil, fmt.Errorf("failed to render page %q: %w", p.Slug, err)
			}
		}
		outPages = append(outPages, OutputPage{Slug: p.Slug, HTML: buf.String()})
	}

	// Sections without an index page get an auto-generated home, mirroring
	// the CLI's auto-indexes: title-cased folder name, date sort, list on.
	sections := []OutputSection{}
	autoDirs := make([]string, 0, len(children))
	for dir := range children {
		if dir != "" && indexBySection[dir] == nil {
			autoDirs = append(autoDirs, dir)
		}
	}
	sort.Strings(autoDirs)
	for _, dir := range autoDirs {
		sorted := make([]*content.Page, len(children[dir]))
		copy(sorted, children[dir])
		sortPages(sorted, "date")
		var buf bytes.Buffer
		if err := tmpl.RenderIndex(&buf, templates.IndexData{
			Site:        site,
			Title:       sitegen.TitleCase(path.Base(dir)),
			Pages:       sorted,
			ShowList:    true,
			CurrentPath: "/" + dir + "/",
		}); err != nil {
			return nil, fmt.Errorf("failed to render section %q: %w", dir, err)
		}
		sections = append(sections, OutputSection{Slug: dir, HTML: buf.String()})
	}

	// The garden home: a root index page replaces the generated index and
	// renders its intro over the listing. Without one, the bridge keeps a
	// synthetic home (hosted gardens always have a home, unlike native
	// sites, which omit / entirely in that case).
	var idx bytes.Buffer
	if root := indexBySection[""]; root != nil {
		if root.Title == "" {
			root.Title = title
		}
		if err := renderSectionHome(&idx, tmpl, site, root, homeEntries); err != nil {
			return nil, fmt.Errorf("failed to render garden home: %w", err)
		}
	} else {
		sorted := make([]*content.Page, len(homeEntries))
		copy(sorted, homeEntries)
		sortPages(sorted, "date")
		if err := tmpl.RenderIndex(&idx, templates.IndexData{
			Site:        site,
			Title:       title,
			Pages:       sorted,
			ShowList:    true,
			CurrentPath: "/",
		}); err != nil {
			return nil, fmt.Errorf("failed to render index: %w", err)
		}
	}

	tags, err := renderTagPages(tmpl, site, pages)
	if err != nil {
		return nil, err
	}

	hasOrigin := sitegen.HasOrigin(cfg.Site.BaseURL)
	if !hasOrigin {
		warnings = append(warnings, "site.baseURL is empty; skipping sitemap.xml, feed.xml, and the robots Sitemap directive (they require an absolute origin)")
	}
	artifacts, err := renderArtifacts(tmpl, site, cfg, pages, resolver, hasOrigin)
	if err != nil {
		return nil, err
	}

	// Self-contained output is the default: families with no self-hosted
	// source fall back to the CSS system stacks, and the author is told.
	if !cfg.Theme.RemoteFonts {
		for _, family := range templates.UnhostedFamilies(cfg.Theme) {
			warnings = append(warnings, templates.UnhostedFontWarning(family))
		}
	}

	manifest, assetWarnings, err := buildAssetManifest(in, cfg, pages)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, assetWarnings...)
	if in.Options.EmitAssets {
		artifacts = append(artifacts, builtinAssetArtifacts(manifest)...)
	}

	return &Output{
		Pages:           outPages,
		Index:           idx.String(),
		Sections:        sections,
		Tags:            tags,
		CSS:             sitegen.Styles(in.Content.StyleCSS, cfg.Theme),
		AssetManifest:   manifest,
		AssetRegistryID: assets.RegistryID(),
		Artifacts:       artifacts,
		Warnings:        warnings,
	}, nil
}

// hasDotSegment reports whether a slash-separated path contains a "." or ".."
// segment. Such segments are rejected: the renderer emits slugs and output
// paths that hosts materialize, so a dot segment could escape a garden's route.
func hasDotSegment(p string) bool {
	for _, seg := range strings.Split(p, "/") {
		if seg == "." || seg == ".." {
			return true
		}
	}
	return false
}

// callerAssets validates the caller-supplied asset manifest: the declaration
// of user files (custom fonts today) the caller will serve alongside the
// rendered site. A pure renderer cannot discover uploads, so this input is
// how they enter the asset contract.
func callerAssets(in *Input) (assets.Manifest, error) {
	if len(in.Content.Assets) == 0 {
		return nil, nil
	}
	// Policy (reserved namespace, outputPath restricted to built-in
	// overrides) is shared with the CLI via assets.ValidateUserAsset, so
	// the two interfaces cannot drift on what a legal user asset is.
	for i, a := range in.Content.Assets {
		if err := assets.ValidateUserAsset(a); err != nil {
			return nil, inputErrorf("assets[%d]: %v", i, err)
		}
	}
	m, err := assets.NewManifest(in.Content.Assets...)
	if err != nil {
		return nil, inputErrorf("assets: %v", err)
	}
	return m, nil
}

// buildAssetManifest produces the combined referenced manifest: required
// built-ins with caller entries merged over them (a caller entry replaces a
// built-in on effective-output-path collision — the favicon-override rule),
// plus warnings for custom font files the configuration references but the
// caller never declared. Mermaid is included only when content uses diagrams.
func buildAssetManifest(in *Input, cfg *config.Config, pages []*content.Page) (assets.Manifest, []string, error) {
	caller, err := callerAssets(in)
	if err != nil {
		return nil, nil, err
	}

	required := assets.RequiredBuiltinsFor(
		content.UsesMermaid(pages),
		cfg.Theme.FontHeading, cfg.Theme.FontBody, cfg.Theme.FontMono,
	)
	builtinAssets := make([]assets.Asset, 0, len(required))
	for _, b := range required {
		builtinAssets = append(builtinAssets, b.Asset)
	}
	base, err := assets.NewManifest(builtinAssets...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build built-in manifest: %w", err)
	}

	merged, err := assets.Merge(base, caller)
	if err != nil {
		return nil, nil, inputErrorf("assets: %v", err)
	}

	var warnings []string
	declared := map[string]bool{}
	for _, a := range caller {
		declared[a.LogicalPath] = true
	}
	for _, face := range cfg.Theme.Fonts {
		if !declared[face.File] {
			warnings = append(warnings, fmt.Sprintf(
				"custom font %q references %s, which is not declared in the caller asset manifest; the site will 404 it unless the host serves it another way",
				face.Family, face.File))
		}
	}
	return merged, warnings, nil
}

// builtinAssetArtifacts returns base64 artifacts for the manifest entries
// whose bytes the renderer actually has: built-ins that survived the merge.
// Caller assets never produce byte artifacts (only the caller has them), and
// an overridden built-in is not emitted. Artifact paths use the effective
// output path — the exact filename a CLI export serves — so generic hosts
// can store artifacts by path without remapping.
func builtinAssetArtifacts(manifest assets.Manifest) []OutputArtifact {
	var out []OutputArtifact
	for _, entry := range manifest {
		b, ok := assets.BuiltinByLogicalPath(entry.LogicalPath)
		if !ok || b.Asset.SHA256 != entry.SHA256 {
			continue
		}
		out = append(out, OutputArtifact{
			Path:        entry.EffectiveOutputPath(),
			Content:     base64.StdEncoding.EncodeToString(b.Content()),
			ContentType: entry.ContentType,
			Encoding:    "base64",
		})
	}
	return out
}

// renderSectionHome renders an index page as its section's home, mirroring
// the CLI's renderSectionIndex: the page body becomes the intro above the
// child listing, sorted by the page's sort key (default date), with the
// listing toggled by showList (default true).
func renderSectionHome(buf *bytes.Buffer, tmpl *templates.Templates, site templates.SiteData, indexPage *content.Page, sectionPages []*content.Page) error {
	sorted := make([]*content.Page, len(sectionPages))
	copy(sorted, sectionPages)
	sortPages(sorted, indexPage.SectionSort)

	showList := true
	if indexPage.ShowList != nil {
		showList = *indexPage.ShowList
	}
	currentPath := "/" + indexPage.Slug
	if currentPath != "/" {
		currentPath += "/"
	}
	return tmpl.RenderIndex(buf, templates.IndexData{
		Site:        site,
		Title:       indexPage.Title,
		Pages:       sorted,
		Intro:       indexPage.HTMLContent,
		ShowList:    showList,
		CurrentPath: currentPath,
	})
}

// renderTagPages renders the tags index and one page per tag, matching the
// CLI's tag output (tags are grouped case-insensitively; pages within a tag
// are sorted by date; tags are listed alphabetically). Page headers link to
// {baseUrl}/tags/<tag>/, so these pages are what keeps those links alive.
func renderTagPages(tmpl *templates.Templates, site templates.SiteData, pages []*content.Page) (OutputTags, error) {
	out := OutputTags{Pages: []OutputTagPage{}}

	// Group pages by lowercased tag (mirrors the CLI's buildTagIndex).
	pagesByTag := make(map[string][]*content.Page)
	for _, p := range pages {
		for _, tag := range p.Tags {
			key := strings.ToLower(tag)
			pagesByTag[key] = append(pagesByTag[key], p)
		}
	}
	if len(pagesByTag) == 0 {
		// No tags anywhere: no tag links are rendered on pages (the tag
		// block is per-page), so emit nothing — like the CLI, which skips
		// the tags/ output directory entirely.
		return out, nil
	}

	tagNames := make([]string, 0, len(pagesByTag))
	for tag := range pagesByTag {
		tagNames = append(tagNames, tag)
	}
	sort.Strings(tagNames)

	tagInfos := make([]templates.TagInfo, 0, len(tagNames))
	for _, tag := range tagNames {
		tagInfos = append(tagInfos, templates.TagInfo{Name: tag, Count: len(pagesByTag[tag])})
	}

	var idx bytes.Buffer
	if err := tmpl.RenderTagIndex(&idx, templates.TagIndexData{
		Site:        site,
		Tags:        tagInfos,
		CurrentPath: "/tags/",
	}); err != nil {
		return out, fmt.Errorf("failed to render tag index: %w", err)
	}
	out.Index = idx.String()

	for _, tag := range tagNames {
		tagged := make([]*content.Page, len(pagesByTag[tag]))
		copy(tagged, pagesByTag[tag])
		sortPages(tagged, "date")

		var buf bytes.Buffer
		if err := tmpl.RenderTagPage(&buf, templates.TagPageData{
			Site:        site,
			Tag:         tag,
			Pages:       tagged,
			CurrentPath: "/tags/" + tag + "/",
		}); err != nil {
			return out, fmt.Errorf("failed to render tag page %q: %w", tag, err)
		}
		out.Pages = append(out.Pages, OutputTagPage{Tag: tag, HTML: buf.String()})
	}

	return out, nil
}

// resolveConfig parses the shared config object into template data. Canonical
// config is parsed by core/config, so the CLI and renderer share defaults and
// validation. Explicit nav items are escaped here at the config trust
// boundary; automatic nav is assembled later from already-escaped page titles.
func resolveConfig(in *Input) (*config.Config, templates.SiteData, error) {
	cfg, err := parseConfig(in.Config)
	if err != nil {
		return nil, templates.SiteData{}, err
	}
	basePath, err := basePathFromURL(cfg.Site.BaseURL)
	if err != nil {
		return nil, templates.SiteData{}, err
	}

	for i := range cfg.Navigation.Items {
		cfg.Navigation.Items[i].Label = html.EscapeString(cfg.Navigation.Items[i].Label)
		cfg.Navigation.Items[i].Path = html.EscapeString(cfg.Navigation.Items[i].Path)
	}

	var footerAttribution *templates.FooterAttribution
	if in.Render.FooterAttribution != nil {
		footerAttribution = &templates.FooterAttribution{
			Name: in.Render.FooterAttribution.Name,
			URL:  in.Render.FooterAttribution.URL,
		}
	}
	rawSite := templates.SiteData{
		Title:             cfg.Site.Title,
		Description:       cfg.Site.Description,
		Author:            cfg.Site.Author,
		Theme:             cfg.Theme,
		BaseURL:           cfg.Site.BaseURL,
		BasePath:          basePath,
		Image:             cfg.Site.Image,
		TOC:               cfg.Features.TOC,
		Graph:             cfg.Features.Graph,
		Search:            cfg.Features.Search,
		RSS:               cfg.Features.RSS,
		HeadExtra:         cfg.Site.HeadExtra,
		FooterAttribution: footerAttribution,
	}
	return cfg, safeSiteData(rawSite), nil
}

// parseConfig turns the optional shared config object into a validated Config,
// falling back to defaults when absent. An empty or null object renders the
// default site.
func parseConfig(raw json.RawMessage) (*config.Config, error) {
	if len(bytes.TrimSpace(raw)) == 0 || string(bytes.TrimSpace(raw)) == "null" {
		return config.Default(), nil
	}
	cfg, err := config.Parse(raw)
	if err != nil {
		return nil, inputErrorf("invalid config: %v", err)
	}
	return cfg, nil
}

// basePathFromURL derives the internal link base path ("" or "/prefix" with no
// trailing slash) from the canonical absolute baseURL's path component.
func basePathFromURL(baseURL string) (string, error) {
	if baseURL == "" {
		return "", nil
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", inputErrorf("invalid config.site.baseURL: %v", err)
	}
	return strings.TrimSuffix(parsed.Path, "/"), nil
}

// safeSiteData protects text/template interpolation at the renderer trust
// boundary. headExtra deliberately remains raw because it is an explicit
// trusted Leafpress configuration escape hatch, matching the CLI. Nav is set
// later from BuildNavigation, whose parts are already escaped.
func safeSiteData(site templates.SiteData) templates.SiteData {
	site.Title = html.EscapeString(site.Title)
	site.Description = html.EscapeString(site.Description)
	site.Author = html.EscapeString(site.Author)
	site.BaseURL = html.EscapeString(site.BaseURL)
	site.Image = html.EscapeString(site.Image)
	if site.FooterAttribution != nil {
		copy := *site.FooterAttribution
		copy.Name = html.EscapeString(copy.Name)
		copy.URL = html.EscapeString(copy.URL)
		site.FooterAttribution = &copy
	}
	return site
}

func validateFooterAttribution(attribution *FooterAttribution) error {
	if attribution == nil {
		return nil
	}
	if strings.TrimSpace(attribution.Name) == "" {
		return inputErrorf("render.footerAttribution.name is required")
	}
	if attribution.URL == "" {
		return nil
	}
	parsed, err := url.Parse(attribution.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return inputErrorf("render.footerAttribution.url must be an absolute http(s) URL")
	}
	return nil
}

func renderArtifacts(
	tmpl *templates.Templates,
	safeSite templates.SiteData,
	cfg *config.Config,
	pages []*content.Page,
	resolver *content.LinkResolver,
	hasOrigin bool,
) ([]OutputArtifact, error) {
	artifactPages := make([]*content.Page, 0, len(pages))
	for _, page := range pages {
		copy := *page
		copy.Title = html.UnescapeString(copy.Title)
		copy.Description = html.UnescapeString(copy.Description)
		copy.Image = html.UnescapeString(copy.Image)
		artifactPages = append(artifactPages, &copy)
	}
	rawSite := safeSite
	rawSite.Title = html.UnescapeString(rawSite.Title)
	rawSite.Description = html.UnescapeString(rawSite.Description)
	rawSite.Author = html.UnescapeString(rawSite.Author)
	rawSite.BaseURL = cfg.Site.BaseURL
	rawSite.Image = html.UnescapeString(rawSite.Image)

	// search-index.json is always emitted: full-text search UI and hover link
	// previews share it. cfg.Features.Search only controls the search UI chrome/JS.
	graphJSON, searchJSON, err := sitegen.GraphSearch(
		artifactPages, resolver, safeSite.BasePath, cfg.Features.Graph, true,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate graph/search artifacts: %w", err)
	}
	artifacts := make([]OutputArtifact, 0, 6)
	if graphJSON != "" {
		artifacts = append(artifacts, OutputArtifact{
			Path: "graph.json", Content: graphJSON, ContentType: "application/json",
		})
	}
	if searchJSON != "" {
		artifacts = append(artifacts, OutputArtifact{
			Path: "search-index.json", Content: searchJSON, ContentType: "application/json",
		})
	}
	// robots.txt is always emitted; Robots() omits the Sitemap directive when
	// baseURL is empty. sitemap.xml and feed.xml require an absolute origin.
	artifacts = append(artifacts,
		OutputArtifact{Path: "robots.txt", Content: sitegen.Robots(cfg.Site.BaseURL), ContentType: "text/plain; charset=utf-8"},
	)
	if hasOrigin {
		artifacts = append(artifacts,
			OutputArtifact{Path: "sitemap.xml", Content: sitegen.Sitemap(artifactPages, cfg.Site.BaseURL), ContentType: "application/xml"},
		)
	}
	notFound, err := sitegen.NotFound(tmpl, safeSite)
	if err != nil {
		return nil, fmt.Errorf("failed to render 404 artifact: %w", err)
	}
	artifacts = append(artifacts, OutputArtifact{
		Path: "404.html", Content: notFound, ContentType: "text/html; charset=utf-8",
	})
	if cfg.Features.RSS && hasOrigin {
		artifacts = append(artifacts, OutputArtifact{
			Path:        "feed.xml",
			Content:     sitegen.RSS(artifactPages, rawSite, cfg.Site.BaseURL, time.Time{}),
			ContentType: "application/rss+xml",
		})
	}
	// Every generated artifact is text; asset artifacts (base64) are
	// appended by the caller.
	for i := range artifacts {
		artifacts[i].Encoding = "utf8"
	}
	return artifacts, nil
}

// buildPages converts input pages into content pages.
func buildPages(in []InputPage) ([]*content.Page, error) {
	pages := make([]*content.Page, 0, len(in))
	seen := make(map[string]bool, len(in))
	for i, ip := range in {
		slug := strings.Trim(ip.Slug, "/")
		// Only the garden-home index page may have an empty slug (the
		// CLI's root _index.md).
		if slug == "" && !ip.IsIndex {
			return nil, inputErrorf("pages[%d].slug is required", i)
		}
		if strings.ContainsAny(slug, unsafeSlugChars) || hasDotSegment(slug) {
			return nil, inputErrorf("pages[%d].slug is invalid: %q", i, ip.Slug)
		}
		if seen[slug] {
			if slug == "" {
				return nil, inputErrorf("duplicate garden home index page")
			}
			return nil, inputErrorf("duplicate page slug: %q", slug)
		}
		seen[slug] = true

		switch ip.Sort {
		case "", "date", "title", "growth":
		default:
			return nil, inputErrorf("pages[%d].sort must be one of date, title, growth; got %q", i, ip.Sort)
		}

		for _, tag := range ip.Tags {
			if !tagRegex.MatchString(tag) {
				return nil, inputErrorf("pages[%d] has invalid tag %q: tags may only contain letters, digits, underscores, and hyphens", i, tag)
			}
		}

		// Growth flows into class/data attributes unescaped; restrict it to
		// the known enum.
		switch ip.Growth {
		case "", "seedling", "budding", "evergreen":
		default:
			return nil, inputErrorf("pages[%d].growth must be one of seedling, budding, evergreen; got %q", i, ip.Growth)
		}

		created, err := parseTime(ip.CreatedAt)
		if err != nil {
			return nil, inputErrorf("pages[%d].createdAt: %v", i, err)
		}
		updated, err := parseTime(ip.UpdatedAt)
		if err != nil {
			return nil, inputErrorf("pages[%d].updatedAt: %v", i, err)
		}
		if ip.ReadingTime != nil && *ip.ReadingTime <= 0 {
			return nil, inputErrorf("pages[%d].readingTime must be a positive integer", i)
		}

		// Title and description reach <title>, <h1>, and og:/twitter: meta
		// attributes through text/template — escape at the input boundary.
		title := html.EscapeString(ip.Title)
		if title == "" && slug != "" {
			if ip.IsIndex {
				// Mirror the scanner's generateTitleFromSlug fallback for
				// _index.md files without a frontmatter title. (An untitled
				// root index falls back to the garden title at render time.)
				title = html.EscapeString(titleFromSlug(path.Base(slug)))
			} else {
				title = html.EscapeString(slug)
			}
		}

		permalink := "/" + slug + "/"
		if slug == "" {
			permalink = "/"
		}
		pages = append(pages, &content.Page{
			Title:               title,
			Description:         html.EscapeString(ip.Description),
			Date:                created,
			Created:             created,
			Modified:            updated,
			Tags:                ip.Tags,
			Growth:              ip.Growth,
			TOC:                 ip.TOC,
			Image:               html.EscapeString(ip.Image),
			Slug:                slug,
			Permalink:           permalink,
			RawContent:          ip.Markdown,
			IsIndex:             ip.IsIndex,
			SectionSort:         ip.Sort,
			ShowList:            ip.ShowList,
			ReadingTimeOverride: ip.ReadingTime,
		})
	}
	return pages, nil
}

// titleFromSlug turns a slug segment into a display title, matching the
// scanner's generateTitleFromSlug (hyphens to spaces, each word capitalized).
func titleFromSlug(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "-", " "))
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

// parseTime parses an optional RFC3339 timestamp ("" means unset).
func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("must be RFC3339: %v", err)
	}
	return t, nil
}

// sortPages orders pages for the index using the same semantics as the
// CLI's section sort (cli/internal/build). sort.SliceStable keeps output
// deterministic when keys compare equal.
func sortPages(pages []*content.Page, sortBy string) {
	switch sortBy {
	case "title":
		sort.SliceStable(pages, func(i, j int) bool {
			return pages[i].Title < pages[j].Title
		})
	case "growth":
		growthOrder := map[string]int{"seedling": 0, "budding": 1, "evergreen": 2, "": 3}
		sort.SliceStable(pages, func(i, j int) bool {
			return growthOrder[pages[i].Growth] < growthOrder[pages[j].Growth]
		})
	default: // date - display date logic (modified if present, otherwise created)
		sort.SliceStable(pages, func(i, j int) bool {
			dateI := pages[i].Date
			if pages[i].HasModified() {
				dateI = pages[i].Modified
			}
			dateJ := pages[j].Date
			if pages[j].HasModified() {
				dateJ = pages[j].Modified
			}
			return dateI.After(dateJ)
		})
	}
}
