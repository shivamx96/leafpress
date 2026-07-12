// Command leafpress-render is a pure stdin→stdout JSON bridge that renders
// a garden (a set of published pages) into full HTML documents, an index
// page, and theme CSS. No filesystem, network, or database access.
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
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "leafpress-render: failed to read stdin: %v\n", err)
		os.Exit(1)
	}

	out, err := render.Run(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "leafpress-render: %v\n", err)
		var inputErr *render.InputError
		if errors.As(err, &inputErr) {
			os.Exit(1)
		}
		os.Exit(2)
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "leafpress-render: failed to encode output: %v\n", err)
		os.Exit(2)
	}
}
