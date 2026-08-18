package content

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

var outputTagNameRegex = regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)

// ValidateOutputRoutes rejects page sets whose generated HTML would claim the
// same URL more than once. Besides duplicate page slugs, this accounts for
// section indexes synthesized for directories and tag pages synthesized from
// metadata.
func ValidateOutputRoutes(pages []*Page) error {
	claims := make(map[string]string)
	claim := func(route, owner string) error {
		route = strings.Trim(route, "/")
		if previous, exists := claims[route]; exists && previous != owner {
			return fmt.Errorf("output route %q is claimed by both %s and %s", displayRoute(route), previous, owner)
		}
		claims[route] = owner
		return nil
	}

	indexBySection := make(map[string]bool)
	for _, page := range pages {
		if page == nil {
			continue
		}
		if err := claim(page.Slug, pageRouteOwner(page)); err != nil {
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
