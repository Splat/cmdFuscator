// Package data embeds the bundled ArgFuscator-compatible JSON profile files.
// Import this package and pass ModelFS to loader.LoadFS or tui.New.
//
// To add more profiles, place *.json files in data/models/ and rebuild.
package data

import "embed"

// ModelFS contains all JSON profile files from data/models/.
// The directory tree within the FS is preserved: files are at "models/<name>.json".
//
//go:embed models/*.json
var ModelFS embed.FS
