// Package filepath implements the FilePathTransformer obfuscation modifier.
//
// Technique: obfuscate file path tokens using one or more of:
//   - PathTraversal:     insert redundant ../ sequences (e.g. C:\foo → C:\.\foo)
//   - SubstituteSlashes: swap / for \ or vice versa (Windows tolerates both)
//   - ExtraSlashes:      add duplicate slashes (C:\\foo\\bar)
//
// ArgFuscator reference: src/Modifiers/FilePathTransformer.ts
// Applies to token types: path, value
package filepath

import (
	"encoding/json"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

func init() {
	modifiers.Register(&FilePathTransformer{})
}

// FilePathTransformer transforms file path tokens to evade path-based signatures.
type FilePathTransformer struct{}

func (f *FilePathTransformer) Name() string { return "FilePathTransformer" }
func (f *FilePathTransformer) Description() string {
	return "Add path traversal, swap slashes, or duplicate separators"
}

// Config holds FilePathTransformer-specific config fields.
type Config struct {
	models.BaseModifierConfig
	PathTraversal     bool `json:"PathTraversal"`
	SubstituteSlashes bool `json:"SubstituteSlashes"`
	ExtraSlashes      bool `json:"ExtraSlashes"`
}

// Apply implements modifiers.Modifier.
//
// TODO: Implement this method.
//
// Steps:
//  1. Unmarshal cfg into a Config struct.
//  2. Parse Probability.
//  3. For each eligible token:
//     a. Roll probability; if not triggered, skip.
//     b. If Config.SubstituteSlashes: randomly swap '/' ↔ '\'.
//     c. If Config.PathTraversal: insert a "./" or ".\" segment at a random
//        position between path components.
//     d. If Config.ExtraSlashes: double one or more separator characters.
//  4. Return updated tokens.
func (f *FilePathTransformer) Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error) {
	return tokens, modifiers.ErrNotImplemented
}
