package templates

import (
	"strings"
	"testing"
)

// mermaidClientScript renders the shared client bundle and returns it.
func mermaidClientScript(t *testing.T) string {
	t.Helper()
	tmpl, err := New()
	if err != nil {
		t.Fatal(err)
	}
	_, script, err := tmpl.ClientScriptAsset(SiteData{})
	if err != nil {
		t.Fatal(err)
	}
	return script
}

// The hardening is only real if a diagram cannot undo it from its own source.
// mermaid ignores directive keys named in `secure`, so every key carrying the
// hardening has to appear there.
func TestMermaidLocksHardeningKeys(t *testing.T) {
	script := mermaidClientScript(t)

	secureAt := strings.Index(script, "secure: [")
	if secureAt < 0 {
		t.Fatalf("mermaid init does not declare a secure list:\n%s", script)
	}
	secureList := script[secureAt:]
	secureList = secureList[:strings.Index(secureList, "]")]

	// Every key the init sets for security reasons must be locked. flowchart
	// and sequence are locked wholesale because mermaid enforces the list at
	// top-level key granularity, which is the only way to pin the label flags
	// nested inside them.
	for _, key := range []string{
		"secure", "securityLevel", "htmlLabels",
		"flowchart", "sequence", "legacyMathML", "forceLegacyMathML",
	} {
		if !strings.Contains(secureList, "'"+key+"'") {
			t.Errorf("secure list does not lock %q:\n%s", key, secureList)
		}
	}
}

func TestMermaidInitDisablesHTMLLabelsAndMath(t *testing.T) {
	script := mermaidClientScript(t)

	for _, want := range []string{
		"securityLevel: 'strict'",
		"htmlLabels: false",
		"useHtmlLabels: false",
		"legacyMathML: false",
		"forceLegacyMathML: false",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("mermaid init is missing %q", want)
		}
	}
	if strings.Contains(script, "htmlLabels: true") {
		t.Error("mermaid init enables HTML labels somewhere")
	}
}

// The script must come from the vendored copy. A CDN URL would defeat the
// point of embedding 3.5MB of mermaid in the binary.
func TestMermaidLoadsVendoredScriptOnly(t *testing.T) {
	script := mermaidClientScript(t)

	if !strings.Contains(script, "'/static/leafpress/mermaid/mermaid.min.js'") {
		t.Error("mermaid is not loaded from the vendored path")
	}
	for _, host := range []string{"cdn.jsdelivr.net", "unpkg.com", "cdnjs", "//cdn."} {
		if strings.Contains(script, host) {
			t.Errorf("client script references a CDN (%s)", host)
		}
	}
}

// Diagrams take their look from the active preset's --lp-diagram-* custom
// properties instead of mermaid's stock palette, and follow the reader's
// theme toggle by rendering again rather than inverting the SVG.
func TestMermaidThemedFromPresetVariables(t *testing.T) {
	script := mermaidClientScript(t)

	for _, want := range []string{
		"theme: 'base'",
		"themeVariables: themeVariables()",
		"'--lp-diagram-font'",
		"'--lp-diagram-surface'",
		"'--lp-diagram-primary'",
		"'--lp-diagram-primary-border'",
		"'--lp-diagram-secondary'",
		"'--lp-diagram-tertiary'",
		"'--lp-diagram-line'",
		"'--lp-diagram-text'",
		"document.addEventListener('lp:themechange', render)",
		"document.dispatchEvent(new CustomEvent('lp:themechange'",
		"mermaid.run({ nodes: diagrams })",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("client script is missing %q", want)
		}
	}
	if strings.Contains(script, "theme: 'default'") {
		t.Error("mermaid init still uses the stock default theme")
	}
}
