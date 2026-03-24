package main

import "embed"

//go:embed agents rules schemas CLAUDE.md .claude
var EmbeddedFiles embed.FS
