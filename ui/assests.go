package ui

import "embed"

// The web UI used to display the data
//go:embed index.html
var Content embed.FS
