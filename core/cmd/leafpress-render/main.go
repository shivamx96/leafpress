// Command leafpress-render is a pure stdin→stdout JSON bridge that renders
// a garden (a set of published pages) into full HTML documents, indexes,
// theme/user CSS, and canonical Leafpress site artifacts. No filesystem,
// network, or database access.
//
// Exit codes: 0 success (warnings allowed), 1 invalid input, 2 internal
// render failure. Only JSON is ever written to stdout.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/shivamx96/leafpress/core/render"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr))
}

// run is main minus the process boundary, so tests can drive the full
// bridge contract (exit codes, stdout purity, stderr messages) in-process.
func run(stdin io.Reader, stdout, stderr io.Writer) int {
	raw, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(stderr, "leafpress-render: failed to read stdin: %v\n", err)
		return 1
	}

	out, err := render.Run(raw)
	if err != nil {
		fmt.Fprintf(stderr, "leafpress-render: %v\n", err)
		return exitCode(err)
	}

	if err := json.NewEncoder(stdout).Encode(out); err != nil {
		fmt.Fprintf(stderr, "leafpress-render: failed to encode output: %v\n", err)
		return 2
	}
	return 0
}

// exitCode maps a render failure onto the exit-code contract: invalid
// input is the caller's bug (1), anything else is ours (2).
func exitCode(err error) int {
	var inputErr *render.InputError
	if errors.As(err, &inputErr) {
		return 1
	}
	return 2
}
