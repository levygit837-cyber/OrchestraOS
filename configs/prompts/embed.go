package prompts

import "embed"

// FS holds the embedded prompt catalog (manifest + fragment markdown files).
//
//go:embed manifest.json fragments
var FS embed.FS
