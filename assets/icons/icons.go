// Package icons embeds the application artwork and its pre-scaled variants.
package icons

import "embed"

// Files contains the icon source files used at build time.
//
//go:embed AppIcon.iconset/*.png
var Files embed.FS
