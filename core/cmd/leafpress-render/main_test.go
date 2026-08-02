package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/shivamx96/leafpress/core/render"
)

// hostileGarden carries author-controlled HTML/script in every field that
// reaches output, so stdout-purity assertions run against worst-case input.
const hostileGarden = `{
  "render": {"slug": "shivam"},
  "config": {"site": {
    "title": "<script>alert('t')</script> Garden",
    "baseURL": "https://example.com/g/shivam"
  }},
  "content": {"pages": [
    {
      "slug": "alpha",
      "title": "Alpha </textarea><img src=x onerror=alert(1)>",
      "markdown": "Linking to [[beta]].\n\n<script>alert('body')</script>",
      "tags": ["systems"],
      "createdAt": "2026-05-16T10:00:00Z"
    },
    {
      "slug": "beta",
      "title": "Beta",
      "markdown": "# Heading\n\nBeta content."
    }
  ]}
}`

// decodeSingleJSON asserts stdout holds exactly one JSON object and nothing
// else — the half of the bridge contract exit codes don't cover.
func decodeSingleJSON(t *testing.T, stdout *bytes.Buffer) *render.Output {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	var out render.Output
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", err, stdout.String())
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		t.Fatalf("stdout carries more than one JSON value (next decode: %v)\nstdout: %q", err, stdout.String())
	}
	return &out
}

func TestValidInputExitsZeroWithPureJSONStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(strings.NewReader(hostileGarden), &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %q", code, stderr.String())
	}
	out := decodeSingleJSON(t, &stdout)
	if len(out.Pages) != 2 || out.Index == "" || out.CSS == "" || len(out.Artifacts) == 0 {
		t.Fatalf("output incomplete: %d pages, index %d bytes, css %d bytes, %d artifacts",
			len(out.Pages), len(out.Index), len(out.CSS), len(out.Artifacts))
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr not empty on success: %q", stderr.String())
	}
}

func TestMalformedJSONExitsOne(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(strings.NewReader(`{"garden": <not json`), &stdout, &stderr); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout not empty on failure: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "leafpress-render:") {
		t.Fatalf("stderr missing error message: %q", stderr.String())
	}
}

func TestInvalidInputExitsOne(t *testing.T) {
	// render.slug now defaults with a warning, so the exit-1 contract is
	// exercised with genuinely invalid input: an unsupported contract version
	// and a page slug that escapes the garden route via a "." path segment.
	cases := map[string]string{
		"unsupported contractVersion": `{"contractVersion": 3, "render": {"slug": "g"}, "content": {"pages": []}}`,
		"dot-segment page slug":       `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "essays/../secret", "markdown": "x"}]}}`,
		// A v1 payload must fail loudly, not silently render an empty site.
		"v1 envelope shape": `{"garden": {"slug": "g"}, "pages": [{"slug": "a", "markdown": "hi"}]}`,
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run(strings.NewReader(input), &stdout, &stderr); code != 1 {
				t.Fatalf("exit code = %d, want 1; stderr: %q", code, stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout not empty on failure: %q", stdout.String())
			}
		})
	}
}

func TestStdinReadFailureExitsOne(t *testing.T) {
	var stdout, stderr bytes.Buffer
	failing := iotest.ErrReader(errors.New("pipe broke"))
	if code := run(failing, &stdout, &stderr); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout not empty on failure: %q", stdout.String())
	}
}

func TestExitCodeMapping(t *testing.T) {
	_, inputErr := render.Run([]byte(`not json`))
	if inputErr == nil {
		t.Fatal("expected render.Run to reject malformed JSON")
	}
	if got := exitCode(inputErr); got != 1 {
		t.Fatalf("exitCode(InputError) = %d, want 1", got)
	}
	if got := exitCode(errors.New("internal failure")); got != 2 {
		t.Fatalf("exitCode(internal error) = %d, want 2", got)
	}
}

// TestBinarySmoke builds and execs the real binary once: the process-level
// contract hosted consumers depend on, not just the in-process run function.
func TestBinarySmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in -short mode")
	}
	bin := filepath.Join(t.TempDir(), "leafpress-render")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(hostileGarden)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("binary exited non-zero: %v; stderr: %q", err, stderr.String())
	}
	decodeSingleJSON(t, &stdout)

	bad := exec.Command(bin)
	bad.Stdin = strings.NewReader(`not json`)
	err := bad.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("invalid input: got %v, want exit code 1", err)
	}
}
