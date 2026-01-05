package main

import (
	"os"

	"github.com/shivamx96/leafpress/cli/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Execute(version); err != nil {
		os.Exit(1)
	}
}
