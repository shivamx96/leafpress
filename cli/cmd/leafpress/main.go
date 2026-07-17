package main

import (
	"os"
	"runtime/debug"

	"github.com/shivamx96/leafpress/cli/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Execute(effectiveVersion()); err != nil {
		os.Exit(1)
	}
}

// effectiveVersion keeps release-archive linker flags authoritative and
// falls back to module metadata for binaries built by `go install ...@vX`.
func effectiveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		return selectVersion(version, info.Main.Version)
	}
	return version
}

func selectVersion(linkedVersion, moduleVersion string) string {
	if linkedVersion != "dev" {
		return linkedVersion
	}
	if moduleVersion != "" && moduleVersion != "(devel)" {
		return moduleVersion
	}
	return linkedVersion
}
