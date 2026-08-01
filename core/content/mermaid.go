package content

import "strings"

// mermaidClassMarker is the class the renderer emits on Mermaid diagram divs.
// Both the CLI and the embedded renderer detect diagram usage via this marker
// so mermaid.min.js is only materialized when needed.
const mermaidClassMarker = `class="mermaid"`

// UsesMermaid reports whether any page's rendered HTML contains a Mermaid
// diagram. Call only after markdown has been rendered into HTMLContent.
func UsesMermaid(pages []*Page) bool {
	for _, p := range pages {
		if p != nil && strings.Contains(p.HTMLContent, mermaidClassMarker) {
			return true
		}
	}
	return false
}
