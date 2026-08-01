package content

import "testing"

func TestUsesMermaid(t *testing.T) {
	if UsesMermaid(nil) {
		t.Error("nil pages must be false")
	}
	if UsesMermaid([]*Page{{HTMLContent: "<p>hi</p>"}}) {
		t.Error("plain HTML must not report mermaid")
	}
	if !UsesMermaid([]*Page{
		{HTMLContent: "<p>no</p>"},
		{HTMLContent: `<div class="mermaid">graph TD; A-->B</div>`},
	}) {
		t.Error("mermaid class marker must be detected")
	}
}
