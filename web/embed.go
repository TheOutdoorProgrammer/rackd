// Package web embeds the built React single-page app. The dist directory is
// produced by `npm run build` in this directory; a placeholder is committed so
// the Go build always succeeds.
package web

import "embed"

//go:embed all:dist
var Dist embed.FS
