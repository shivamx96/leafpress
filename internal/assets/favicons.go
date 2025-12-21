package assets

import (
	_ "embed"
)

//go:embed favicon.ico
var FaviconICO []byte

//go:embed favicon.svg
var FaviconSVG []byte

//go:embed favicon-96x96.png
var FaviconPNG []byte
