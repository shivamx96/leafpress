package fonts_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/cli/internal/fonts"
	"github.com/shivamx96/leafpress/cli/internal/fonts/fontstest"
)

func resolve(t *testing.T, root string, server *fontstest.Server, download bool, families ...string) fonts.Result {
	t.Helper()
	result, err := fonts.Resolve(root, families, fonts.Options{
		Download:         download,
		NoDownloadReason: "downloads are off",
		Client:           server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestResolveDownloadsOnceThenReadsTheLock(t *testing.T) {
	root := t.TempDir()
	server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true, Italic: true})

	first := resolve(t, root, server, true, "Test Serif")
	if len(first.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", first.Warnings)
	}
	// Variable family with italics: two styles, Latin and Latin Extended.
	if len(first.Faces) != 4 {
		t.Fatalf("got %d faces, want 4: %+v", len(first.Faces), first.Faces)
	}
	for _, face := range first.Faces {
		if face.Family != "Test Serif" || face.Weight != "300 800" || face.Display != "swap" {
			t.Errorf("unexpected face %+v", face)
		}
		if face.UnicodeRange != fontstest.LatinRange && face.UnicodeRange != fontstest.LatinExtRange {
			t.Errorf("face outside Latin subsets: %+v", face)
		}
		if _, err := os.Stat(filepath.Join(root, face.File)); err != nil {
			t.Errorf("face file not saved: %v", err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "static/fonts/test-serif/OFL.txt")); err != nil {
		t.Errorf("license not saved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, fonts.LockFile)); err != nil {
		t.Fatalf("lock not written: %v", err)
	}

	requests := server.Requests.Load()
	second := resolve(t, root, server, true, "Test Serif")
	if server.Requests.Load() != requests {
		t.Error("a family already in the lock must not be downloaded again")
	}
	if fmt.Sprint(second.Faces) != fmt.Sprint(first.Faces) {
		t.Errorf("lock faces differ from downloaded faces:\n%+v\n%+v", second.Faces, first.Faces)
	}
	offline := resolve(t, root, server, false, "Test Serif")
	if len(offline.Faces) != 4 || len(offline.Warnings) != 0 {
		t.Errorf("a locked family must resolve without downloads: %+v", offline)
	}
}

func TestResolveStaticFamilyDownloadsThemeWeights(t *testing.T) {
	root := t.TempDir()
	server := fontstest.New(t, fontstest.Family{Name: "Test Mono", Weights: []string{"100", "400", "700", "900"}})

	result := resolve(t, root, server, true, "Test Mono")
	weights := map[string]bool{}
	for _, face := range result.Faces {
		weights[face.Weight] = true
	}
	if len(weights) != 2 || !weights["400"] || !weights["700"] {
		t.Errorf("static family should download only the theme weights 400 and 700, got %v", weights)
	}
}

func TestResolveWithoutDownloadsWarns(t *testing.T) {
	server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true})
	result := resolve(t, t.TempDir(), server, false, "Test Serif")
	if len(result.Faces) != 0 || !strings.Contains(result.Warnings["Test Serif"], "downloads are off") {
		t.Errorf("got %+v", result)
	}
	if server.Requests.Load() != 0 {
		t.Error("no request may be made when downloads are off")
	}
}

func TestResolveSuggestsTheClosestFamily(t *testing.T) {
	server := fontstest.New(t, fontstest.Family{Name: "Playfair Display", Variable: true})
	for typo, want := range map[string]string{
		"Playfiar Display": "Playfair Display",
		"playfair display": "Playfair Display",
	} {
		result := resolve(t, t.TempDir(), server, true, typo)
		if !strings.Contains(result.Warnings[typo], fmt.Sprintf("did you mean %q", want)) {
			t.Errorf("%s: warning %q should suggest %q", typo, result.Warnings[typo], want)
		}
	}
	result := resolve(t, t.TempDir(), server, true, "Completely Unknown")
	if w := result.Warnings["Completely Unknown"]; !strings.Contains(w, "is not a Google Fonts family") || strings.Contains(w, "did you mean") {
		t.Errorf("unknown family warning = %q", w)
	}
}

func TestResolveRefusesEditedFiles(t *testing.T) {
	root := t.TempDir()
	server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true})
	first := resolve(t, root, server, true, "Test Serif")
	if err := os.WriteFile(filepath.Join(root, first.Faces[0].File), []byte("wOF2 edited"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := fonts.Resolve(root, []string{"Test Serif"}, fonts.Options{Download: true, Client: server.Client()})
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("an edited font file must stop the build, got %v", err)
	}
}

func TestResolveDownloadsAgainWhenFilesAreMissing(t *testing.T) {
	root := t.TempDir()
	server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true})
	resolve(t, root, server, true, "Test Serif")
	if err := os.RemoveAll(filepath.Join(root, "static/fonts/test-serif")); err != nil {
		t.Fatal(err)
	}
	result := resolve(t, root, server, true, "Test Serif")
	if len(result.Faces) != 2 || len(result.Warnings) != 0 {
		t.Fatalf("deleted family should download again: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(root, result.Faces[0].File)); err != nil {
		t.Error(err)
	}
}

func TestResolveLeavesUnmanagedDirectoriesAlone(t *testing.T) {
	root := t.TempDir()
	own := filepath.Join(root, "static/fonts/test-serif/mine.woff2")
	if err := os.MkdirAll(filepath.Dir(own), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(own, []byte("mine"), 0644); err != nil {
		t.Fatal(err)
	}
	server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true})
	result := resolve(t, root, server, true, "Test Serif")
	if !strings.Contains(result.Warnings["Test Serif"], "was not downloaded by leafpress") {
		t.Errorf("warning = %q", result.Warnings["Test Serif"])
	}
	if data, err := os.ReadFile(own); err != nil || string(data) != "mine" {
		t.Error("an existing directory must not be touched")
	}
}

func TestResolveRejectsUntrustedResponses(t *testing.T) {
	for name, css := range map[string]func(*fontstest.Server) string{
		"foreign host": func(s *fontstest.Server) string {
			return "/* latin */\n@font-face {\n  font-style: normal;\n  font-weight: 400;\n  src: url(https://evil.example/x.woff2) format('woff2');\n  unicode-range: U+0000-00FF;\n}\n"
		},
		"css injection": func(s *fontstest.Server) string {
			return "/* latin */\n@font-face {\n  font-style: normal;\n  font-weight: 400;\n  src: url(" + s.URL + "/files/x.woff2) format('woff2');\n  unicode-range: U+0000-00FF} body{color:red;\n}\n"
		},
		"no latin": func(s *fontstest.Server) string {
			return "/* cyrillic */\n@font-face {\n  font-style: normal;\n  font-weight: 400;\n  src: url(" + s.URL + "/files/x.woff2) format('woff2');\n  unicode-range: U+0400-045F;\n}\n"
		},
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true})
			server.CSS = func(string) string { return css(server) }
			result := resolve(t, root, server, true, "Test Serif")
			if len(result.Faces) != 0 || !strings.Contains(result.Warnings["Test Serif"], "could not download") {
				t.Fatalf("got %+v", result)
			}
			entries, _ := os.ReadDir(filepath.Join(root, "static/fonts"))
			if len(entries) != 0 {
				t.Errorf("a failed download must leave nothing behind, found %v", entries)
			}
		})
	}
}

func TestResolveReportsNetworkFailures(t *testing.T) {
	server := fontstest.New(t, fontstest.Family{Name: "Test Serif", Variable: true})
	client := server.Client()
	server.Close()
	result, err := fonts.Resolve(t.TempDir(), []string{"Test Serif"}, fonts.Options{Download: true, Client: client})
	if err != nil {
		t.Fatal(err)
	}
	if w := result.Warnings["Test Serif"]; !strings.Contains(w, "could not download") || !strings.Contains(w, "later build") {
		t.Errorf("warning = %q", w)
	}
}

func TestReadLockRejectsPathsOutsideTheFontsDirectory(t *testing.T) {
	root := t.TempDir()
	lock := filepath.Join(root, fonts.LockFile)
	if err := os.MkdirAll(filepath.Dir(lock), 0755); err != nil {
		t.Fatal(err)
	}
	body := `{"version":1,"families":{"X":{"licenseFile":"static/fonts/x/OFL.txt","faces":[{"file":"static/fonts/../../secret.woff2"}]}}}`
	if err := os.WriteFile(lock, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := fonts.ReadLock(root); err == nil {
		t.Fatal("a lock pointing outside static/fonts must be rejected")
	}
}
