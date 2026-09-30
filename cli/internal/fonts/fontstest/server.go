// Package fontstest serves a fake Google Fonts API for tests, so font
// downloads can be exercised without network access.
package fontstest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/shivamx96/leafpress/cli/internal/fonts"
)

// Family describes one family the fake service knows.
type Family struct {
	Name string
	// Variable families report a weight axis; static ones list Weights.
	Variable bool
	Weights  []string
	Italic   bool
}

// Server is a running fake. CSS lets a test replace the stylesheet body,
// BeforeCSS runs while a download is in progress, and Requests counts every
// request received.
type Server struct {
	*httptest.Server
	Families  map[string]Family
	CSS       func(family string) string
	BeforeCSS func(family string)
	Requests  atomic.Int64
}

// LatinRange and LatinExtRange are the unicode ranges the fake serves.
const (
	LatinRange    = "U+0000-00FF, U+0131"
	LatinExtRange = "U+0100-02BA, U+1E00-1E9F"
)

// New starts a fake serving the given families.
func New(t *testing.T, families ...Family) *Server {
	t.Helper()
	s := &Server{Families: map[string]Family{}}
	for _, family := range families {
		s.Families[family.Name] = family
	}
	s.Server = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.Close)
	return s
}

// Client returns a fonts client pointed at the fake.
func (s *Server) Client() *fonts.Client {
	return &fonts.Client{
		HTTP:        s.Server.Client(),
		MetadataURL: s.URL + "/metadata/fonts",
		CSSURL:      s.URL + "/css2",
		LicenseURL:  s.URL + "/repo",
		FilePrefix:  s.URL + "/files/",
		UserAgent:   "fontstest",
	}
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	s.Requests.Add(1)
	switch {
	case r.URL.Path == "/metadata/fonts":
		var names []string
		for name := range s.Families {
			names = append(names, fmt.Sprintf(`{"family":%q}`, name))
		}
		fmt.Fprintf(w, `)]}'{"familyMetadataList":[%s]}`, strings.Join(names, ","))
	case strings.HasPrefix(r.URL.Path, "/metadata/fonts/"):
		family, ok := s.Families[strings.TrimPrefix(r.URL.Path, "/metadata/fonts/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, metadataJSON(family))
	case r.URL.Path == "/css2":
		// Read the raw query: the CSS2 API takes unescaped ";" separators,
		// which net/url's query parser rejects.
		raw, _, _ := strings.Cut(strings.TrimPrefix(r.URL.RawQuery, "family="), "&")
		param, _ := url.QueryUnescape(raw)
		name, axes, _ := strings.Cut(param, ":")
		if _, ok := s.Families[name]; !ok {
			http.Error(w, "unknown family", http.StatusBadRequest)
			return
		}
		if s.BeforeCSS != nil {
			s.BeforeCSS(name)
		}
		if s.CSS != nil {
			fmt.Fprint(w, s.CSS(name))
			return
		}
		fmt.Fprint(w, s.stylesheet(s.Families[name], requestedFaces(axes)))
	case strings.HasPrefix(r.URL.Path, "/files/"):
		fmt.Fprint(w, "wOF2"+r.URL.Path)
	case strings.HasPrefix(r.URL.Path, "/repo/ofl/"):
		fmt.Fprint(w, "SIL Open Font License, Version 1.1\n")
	default:
		http.NotFound(w, r)
	}
}

func metadataJSON(family Family) string {
	var keys []string
	weights := family.Weights
	if family.Variable {
		weights = []string{"400", "700"}
	}
	for _, weight := range weights {
		keys = append(keys, fmt.Sprintf(`"%s":{}`, weight))
		if family.Italic {
			keys = append(keys, fmt.Sprintf(`"%si":{}`, weight))
		}
	}
	axes := "[]"
	if family.Variable {
		axes = `[{"tag":"wght","min":300,"max":800}]`
	}
	return fmt.Sprintf(`)]}'{"family":%q,"axes":%s,"fonts":{%s},"license":"ofl"}`, family.Name, axes, strings.Join(keys, ","))
}

type styleWeight struct{ style, weight string }

// requestedFaces parses a CSS2 axis spec such as "ital,wght@0,400;1,700" or
// "wght@300..800" into the styles and weights to serve.
func requestedFaces(axes string) []styleWeight {
	names, values, _ := strings.Cut(axes, "@")
	italAxis := strings.HasPrefix(names, "ital,")
	var faces []styleWeight
	for _, tuple := range strings.Split(values, ";") {
		style, weight := "normal", tuple
		if italAxis {
			ital, w, _ := strings.Cut(tuple, ",")
			weight = w
			if ital == "1" {
				style = "italic"
			}
		}
		faces = append(faces, styleWeight{style, strings.ReplaceAll(weight, "..", " ")})
	}
	return faces
}

// stylesheet returns CSS shaped like the CSS2 API: one face per requested
// style and weight for each subset, including a subset leafpress must skip.
func (s *Server) stylesheet(family Family, requested []styleWeight) string {
	slug := url.PathEscape(fonts.Slug(family.Name))
	subsets := []struct{ name, rng string }{
		{"cyrillic", "U+0400-045F"},
		{"latin-ext", LatinExtRange},
		{"latin", LatinRange},
	}
	var b strings.Builder
	for _, face := range requested {
		for _, subset := range subsets {
			fmt.Fprintf(&b, "/* %s */\n@font-face {\n  font-family: '%s';\n  font-style: %s;\n  font-weight: %s;\n  font-display: swap;\n  src: url(%s/files/%s-%s-%s-%s.woff2) format('woff2');\n  unicode-range: %s;\n}\n",
				subset.name, family.Name, face.style, face.weight, s.URL, slug, face.style, strings.ReplaceAll(face.weight, " ", "-"), subset.name, subset.rng)
		}
	}
	return b.String()
}
