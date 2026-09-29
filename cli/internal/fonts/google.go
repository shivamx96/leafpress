package fonts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Client downloads families from Google Fonts. The URLs are fields so tests
// can point the client at a local server.
type Client struct {
	HTTP *http.Client

	// MetadataURL serves per-family metadata at MetadataURL/<family> and the
	// full family list at MetadataURL itself.
	MetadataURL string
	// CSSURL is the Google Fonts CSS2 API endpoint.
	CSSURL string
	// LicenseURL is the root of the google/fonts repository, which holds
	// each family's license file.
	LicenseURL string
	// FilePrefix is the only URL prefix font files may be fetched from.
	FilePrefix string
	// UserAgent selects the CSS2 response format; the API returns woff2 with
	// per-subset unicode ranges only to modern browsers.
	UserAgent string
}

// Google returns a client for the public Google Fonts service.
func Google() *Client {
	return &Client{
		HTTP:        &http.Client{Timeout: time.Minute},
		MetadataURL: "https://fonts.google.com/metadata/fonts",
		CSSURL:      "https://fonts.googleapis.com/css2",
		LicenseURL:  "https://raw.githubusercontent.com/google/fonts/main",
		FilePrefix:  "https://fonts.gstatic.com/",
		UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36 leafpress",
	}
}

const (
	maxFontFileBytes = 5 << 20
	maxLicenseBytes  = 256 << 10
	maxMetadataBytes = 1 << 20
	maxListBytes     = 16 << 20
	maxCSSBytes      = 1 << 20
)

// subsets are the character sets leafpress downloads, matching the bundled
// fonts: other scripts fall back to the reader's system fonts.
var subsets = []string{"latin", "latin-ext"}

// staticWeights are the weights leafpress themes use. For families without
// a weight axis only these are downloaded, not every static cut.
var staticWeights = []string{"400", "500", "600", "700"}

// NotFoundError reports a family name Google Fonts does not know.
type NotFoundError struct {
	Family     string
	Suggestion string
}

func (e *NotFoundError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%q is not a Google Fonts family (did you mean %q?)", e.Family, e.Suggestion)
	}
	return fmt.Sprintf("%q is not a Google Fonts family", e.Family)
}

type metadata struct {
	Family string `json:"family"`
	Axes   []struct {
		Tag string  `json:"tag"`
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"axes"`
	Fonts   map[string]json.RawMessage `json:"fonts"`
	License string                     `json:"license"`
}

// licenses maps Google Fonts license identifiers to the repository
// directory, file name, and a readable name.
var licenses = map[string]struct{ dir, file, name string }{
	"ofl":    {"ofl", "OFL.txt", "SIL Open Font License 1.1"},
	"apache": {"apache", "LICENSE.txt", "Apache License 2.0"},
	"ufl":    {"ufl", "UFL.txt", "Ubuntu Font Licence 1.0"},
}

// download fetches family into stageDir, naming files for their final home
// under Dir/slug. It returns the lock entry and the number of bytes saved.
func (c *Client) download(ctx context.Context, family, stageDir string) (*Family, int64, error) {
	meta, err := c.metadata(ctx, family)
	if err != nil {
		return nil, 0, err
	}
	license, ok := licenses[meta.License]
	if !ok {
		return nil, 0, fmt.Errorf("%q uses license %q, which leafpress does not recognise", family, meta.License)
	}

	css, err := c.get(ctx, c.CSSURL+"?family="+familyParam(family, cssAxes(meta))+"&display=swap", maxCSSBytes)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch stylesheet: %w", err)
	}
	faces, err := parseFaces(string(css))
	if err != nil {
		return nil, 0, err
	}

	slug := Slug(family)
	finalDir := Dir + "/" + slug
	entry := &Family{
		Source:      "Google Fonts",
		License:     license.name,
		LicenseFile: finalDir + "/" + license.file,
	}
	var total int64
	for _, face := range faces {
		if !strings.HasPrefix(face.url, c.FilePrefix) || !strings.HasSuffix(face.url, ".woff2") {
			return nil, 0, fmt.Errorf("unexpected font file URL %q", face.url)
		}
		data, err := c.get(ctx, face.url, maxFontFileBytes)
		if err != nil {
			return nil, 0, fmt.Errorf("fetch font file: %w", err)
		}
		if !hasPrefix(data, "wOF2") {
			return nil, 0, fmt.Errorf("%s is not a woff2 font", face.url)
		}
		name := fmt.Sprintf("%s-%s-%s.woff2", slug, face.style, face.subset)
		if !strings.Contains(face.weight, " ") {
			name = fmt.Sprintf("%s-%s-%s-%s.woff2", slug, face.style, face.subset, face.weight)
		}
		if err := os.WriteFile(filepath.Join(stageDir, name), data, 0644); err != nil {
			return nil, 0, err
		}
		sum := sha256.Sum256(data)
		entry.Faces = append(entry.Faces, Face{
			File:         finalDir + "/" + name,
			Weight:       face.weight,
			Style:        face.style,
			UnicodeRange: face.unicodeRange,
			SHA256:       hex.EncodeToString(sum[:]),
		})
		total += int64(len(data))
	}

	licenseText, err := c.get(ctx, c.LicenseURL+"/"+license.dir+"/"+repoDir(family)+"/"+license.file, maxLicenseBytes)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch license: %w", err)
	}
	if err := os.WriteFile(filepath.Join(stageDir, license.file), licenseText, 0644); err != nil {
		return nil, 0, err
	}
	return entry, total + int64(len(licenseText)), nil
}

func (c *Client) metadata(ctx context.Context, family string) (*metadata, error) {
	data, err := c.get(ctx, c.MetadataURL+"/"+url.PathEscape(family), maxMetadataBytes)
	var status *statusError
	if errors.As(err, &status) && status.code == http.StatusNotFound {
		return nil, &NotFoundError{Family: family, Suggestion: c.suggest(ctx, family)}
	}
	if err != nil {
		return nil, fmt.Errorf("fetch metadata: %w", err)
	}
	var meta metadata
	if err := json.Unmarshal(stripXSSIPrefix(data), &meta); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}
	if meta.Family != family {
		return nil, &NotFoundError{Family: family, Suggestion: meta.Family}
	}
	return &meta, nil
}

// suggest finds the closest Google Fonts family name for a typo. It is only
// called after a lookup fails, and returns "" if the list is unavailable.
func (c *Client) suggest(ctx context.Context, family string) string {
	data, err := c.get(ctx, c.MetadataURL, maxListBytes)
	if err != nil {
		return ""
	}
	var list struct {
		FamilyMetadataList []struct {
			Family string `json:"family"`
		} `json:"familyMetadataList"`
	}
	if json.Unmarshal(stripXSSIPrefix(data), &list) != nil {
		return ""
	}
	want := strings.ToLower(family)
	best, bestDistance := "", len(want)/3+1
	for _, candidate := range list.FamilyMetadataList {
		name := strings.ToLower(candidate.Family)
		if name == want {
			return candidate.Family
		}
		if d := levenshtein(want, name); d < bestDistance {
			best, bestDistance = candidate.Family, d
		}
	}
	return best
}

type statusError struct {
	url  string
	code int
}

func (e *statusError) Error() string {
	return fmt.Sprintf("GET %s: %s", e.url, http.StatusText(e.code))
}

func (c *Client) get(ctx context.Context, rawURL string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &statusError{url: rawURL, code: resp.StatusCode}
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("GET %s: response exceeds %d bytes", rawURL, limit)
	}
	return data, nil
}

// cssAxes builds the axis part of a CSS2 family parameter: the full weight
// range for variable families, otherwise the static weights leafpress
// themes use. Italics are included when the family has them.
func cssAxes(meta *metadata) string {
	hasItalic := false
	var available []string
	for key := range meta.Fonts {
		weight, italic := strings.CutSuffix(key, "i")
		if italic {
			hasItalic = true
		} else {
			available = append(available, weight)
		}
	}

	var weights []string
	for _, axis := range meta.Axes {
		if axis.Tag == "wght" && axis.Min < axis.Max {
			weights = []string{fmt.Sprintf("%d..%d", int(axis.Min), int(axis.Max))}
		}
	}
	if weights == nil {
		for _, weight := range staticWeights {
			if slices.Contains(available, weight) {
				weights = append(weights, weight)
			}
		}
		if weights == nil {
			// Display families often ship a single weight such as 900.
			weights = available
			slices.SortFunc(weights, func(a, b string) int {
				x, _ := strconv.Atoi(a)
				y, _ := strconv.Atoi(b)
				return x - y
			})
		}
	}
	if !hasItalic {
		return "wght@" + strings.Join(weights, ";")
	}
	var tuples []string
	for _, ital := range []string{"0", "1"} {
		for _, weight := range weights {
			if ital == "1" && !strings.Contains(weight, "..") && !italicAvailable(meta, weight) {
				continue
			}
			tuples = append(tuples, ital+","+weight)
		}
	}
	return "ital,wght@" + strings.Join(tuples, ";")
}

func italicAvailable(meta *metadata, weight string) bool {
	_, ok := meta.Fonts[weight+"i"]
	return ok
}

func familyParam(family, axes string) string {
	return strings.ReplaceAll(url.QueryEscape(family), "%20", "+") + ":" + axes
}

type cssFace struct {
	subset, style, weight, url, unicodeRange string
}

var (
	faceBlock    = regexp.MustCompile(`/\*\s*([a-z0-9-]+)\s*\*/\s*@font-face\s*\{([^}]*)\}`)
	faceProperty = regexp.MustCompile(`(?m)^\s*([a-z-]+):\s*(.+?);\s*$`)
	faceURL      = regexp.MustCompile(`url\(([^)]+)\)`)
	faceWeight   = regexp.MustCompile(`^[0-9]{1,4}( [0-9]{1,4})?$`)
	faceRange    = regexp.MustCompile(`^U\+[0-9A-Fa-f?]{1,6}(-[0-9A-Fa-f]{1,6})?(, ?U\+[0-9A-Fa-f?]{1,6}(-[0-9A-Fa-f]{1,6})?)*$`)
)

// parseFaces extracts the Latin and Latin Extended faces from a CSS2
// response. Every value is checked before it reaches a file name or the
// generated stylesheet.
func parseFaces(css string) ([]cssFace, error) {
	var faces []cssFace
	for _, block := range faceBlock.FindAllStringSubmatch(css, -1) {
		if !slices.Contains(subsets, block[1]) {
			continue
		}
		face := cssFace{subset: block[1]}
		for _, prop := range faceProperty.FindAllStringSubmatch(block[2], -1) {
			switch prop[1] {
			case "font-style":
				face.style = prop[2]
			case "font-weight":
				face.weight = prop[2]
			case "unicode-range":
				face.unicodeRange = prop[2]
			case "src":
				if m := faceURL.FindStringSubmatch(prop[2]); m != nil {
					face.url = strings.Trim(m[1], `'"`)
				}
			}
		}
		if face.style != "normal" && face.style != "italic" {
			return nil, fmt.Errorf("unexpected font-style %q in Google Fonts stylesheet", face.style)
		}
		if !faceWeight.MatchString(face.weight) {
			return nil, fmt.Errorf("unexpected font-weight %q in Google Fonts stylesheet", face.weight)
		}
		if !faceRange.MatchString(face.unicodeRange) {
			return nil, fmt.Errorf("unexpected unicode-range %q in Google Fonts stylesheet", face.unicodeRange)
		}
		if face.url == "" {
			return nil, errors.New("font face without a source in Google Fonts stylesheet")
		}
		faces = append(faces, face)
	}
	if len(faces) == 0 {
		return nil, errors.New("the family has no Latin character set")
	}
	return faces, nil
}

func stripXSSIPrefix(data []byte) []byte {
	trimmed := strings.TrimPrefix(string(data), ")]}'")
	return []byte(trimmed)
}

// Slug turns a family name into its directory and file name prefix, such as
// "Source Serif 4" into "source-serif-4".
func Slug(family string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(family) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// repoDir is the google/fonts repository directory for a family: the name
// lowercased with everything but letters and digits removed.
func repoDir(family string) string {
	return strings.ReplaceAll(Slug(family), "-", "")
}

func levenshtein(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}
