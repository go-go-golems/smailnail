package smailnaildocs

import "embed"

const Dir = "."

// Files embeds the canonical JavaScript-facing API documentation assets.
//
//go:embed *.js
var Files embed.FS
