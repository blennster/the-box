package web

import (
	"embed"
)

//go:embed *
var FS embed.FS

//go:embed static
var Static embed.FS
