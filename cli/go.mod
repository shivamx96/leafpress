module github.com/shivamx96/leafpress/cli

go 1.25.5

require (
	github.com/fsnotify/fsnotify v1.9.0
	github.com/gorilla/websocket v1.5.3
	github.com/shivamx96/leafpress/core v0.0.0
	github.com/spf13/cobra v1.10.2
	golang.org/x/text v0.32.0
)

require (
	github.com/alecthomas/chroma/v2 v2.21.1 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/yuin/goldmark v1.7.13 // indirect
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc // indirect
	golang.org/x/sys v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// drop this once core is tagged (core/vX.Y.Z) so `go install .../cli@latest` works again — release core tag first, then cli.
replace github.com/shivamx96/leafpress/core => ../core
