package web

import "embed"

// FS embeds all static assets and HTML templates.
//
//go:embed static/* templates/*.html templates/components/*.html
var FS embed.FS
