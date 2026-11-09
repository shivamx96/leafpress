package content

import "testing"

func TestPublicationGates(t *testing.T) {
	yes, no := true, false
	for _, mode := range []string{"all", "explicit"} {
		for _, publish := range []*bool{nil, &yes, &no} {
			for _, draft := range []bool{false, true} {
				for _, preview := range []bool{false, true} {
					for _, index := range []bool{false, true} {
						p := &Page{Publish: publish, Draft: draft, IsIndex: index}
						want := (publish != nil && *publish || publish == nil && mode == "all") && (!draft || preview)
						if got := Publishable(p, mode, preview); got != want {
							t.Fatalf("mode=%s publish=%v draft=%v preview=%v index=%v: got %v want %v", mode, publish, draft, preview, index, got, want)
						}
					}
				}
			}
		}
	}
}
