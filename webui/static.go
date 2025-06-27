package webui

import (
	"embed"
)

//go:generate cp -r ../dist ./dist
//go:embed dist
var StaticFS embed.FS
