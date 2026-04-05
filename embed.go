package main

import "embed"

//go:embed agents rules schemas references CLAUDE.md .claude
var EmbeddedFiles embed.FS
