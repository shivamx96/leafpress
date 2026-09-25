package content

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var outputTagNameRegex = regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)

// ValidateSlug checks the shared filesystem and URL contract. Empty denotes
// the site root. Keep Unicode names, but reject ambiguous URL syntax and names
// that cannot be materialized consistently on supported operating systems.
func ValidateSlug(slug string) error {
	if slug == "" {
		return nil
	}
	if !utf8.ValidString(slug) {
		return fmt.Errorf("slug must be valid UTF-8")
	}
	for _, r := range slug {
		if unicode.IsSpace(r) || unicode.IsControl(r) || strings.ContainsRune("?#%&\"'<>\\:*|", r) {
			return fmt.Errorf("slug contains unsupported character %q; use letters, numbers, hyphens, underscores, or dots", r)
		}
	}
	for _, segment := range strings.Split(slug, "/") {
		if segment == "" || segment == "." || segment == ".." || strings.HasSuffix(segment, ".") {
			return fmt.Errorf("slug contains invalid path segment %q", segment)
		}
		// Windows device names stay reserved even with a filename extension.
		base, _, _ := strings.Cut(strings.ToUpper(segment), ".")
		reserved := base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || base == "CONIN$" || base == "CONOUT$"
		if strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT") {
			suffix := strings.TrimPrefix(strings.TrimPrefix(base, "COM"), "LPT")
			reserved = reserved || (len([]rune(suffix)) == 1 && strings.ContainsAny(suffix, "123456789¹²³"))
		}
		if reserved {
			return fmt.Errorf("slug contains reserved Windows path segment %q", segment)
		}
	}
	return nil
}

// ValidateOutputRoutes rejects page sets whose generated HTML would claim the
// same URL more than once. Besides duplicate page slugs, this accounts for
// section indexes synthesized for directories and tag pages synthesized from
// metadata.
func ValidateOutputRoutes(pages []*Page) error {
	claims := make(map[string]string)
	claim := func(route, owner string) error {
		route = strings.Trim(route, "/")
		if err := ValidateSlug(route); err != nil {
			return fmt.Errorf("%s has invalid output route %q: %w", owner, route, err)
		}
		key := strings.ToLower(route)
		if previous, exists := claims[key]; exists && previous != owner {
			return fmt.Errorf("output route %q is claimed by both %s and %s", displayRoute(route), previous, owner)
		}
		claims[key] = owner
		return nil
	}

	indexBySection := make(map[string]bool)
	for _, page := range pages {
		if page == nil {
			continue
		}
		if err := ValidateSlug(page.Slug); err != nil {
			err = fmt.Errorf("%s has invalid slug %q: %w", pageRouteOwner(page), page.Slug, err)
			if page.SourcePath != "" {
				return fmt.Errorf("%w; rename the file or folder, or set a different slug in the page's frontmatter", err)
			}
			return err
		}
		if err := claim(page.Slug, pageRouteOwner(page)); err != nil {
			if page.SourcePath != "" {
				// Cleaned filenames can meet: "My Note.md" and "My-Note.md".
				return fmt.Errorf("%w; rename one of the files or give one a different slug in its frontmatter", err)
			}
			return err
		}
		if page.IsIndex {
			indexBySection[page.Slug] = true
		}
	}

	// A direct child causes Leafpress to synthesize its parent section route
	// unless an explicit _index page already owns it.
	for _, page := range pages {
		if page == nil || page.IsIndex {
			continue
		}
		section := path.Dir(filepathToSlash(page.Slug))
		if section == "." || indexBySection[section] {
			continue
		}
		if err := claim(section, fmt.Sprintf("generated section %q", section)); err != nil {
			return err
		}
	}

	seenTags := make(map[string]bool)
	for _, page := range pages {
		if page == nil {
			continue
		}
		for _, tag := range page.Tags {
			if !outputTagNameRegex.MatchString(tag) {
				return fmt.Errorf("%s has invalid tag %q: tags may only contain letters, digits, underscores, and hyphens", pageRouteOwner(page), tag)
			}
			tag = strings.ToLower(tag)
			if seenTags[tag] {
				continue
			}
			seenTags[tag] = true
			if err := claim("tags", "generated tag index"); err != nil {
				return err
			}
			if err := claim(path.Join("tags", tag), fmt.Sprintf("generated tag %q", tag)); err != nil {
				return err
			}
		}
	}

	return nil
}

func pageRouteOwner(page *Page) string {
	if page.SourcePath != "" {
		return fmt.Sprintf("page %q", filepathToSlash(page.SourcePath))
	}
	return fmt.Sprintf("page with slug %q", page.Slug)
}

func displayRoute(route string) string {
	if route == "" {
		return "/"
	}
	return "/" + route + "/"
}

func filepathToSlash(value string) string {
	return strings.ReplaceAll(value, `\`, "/")
}
