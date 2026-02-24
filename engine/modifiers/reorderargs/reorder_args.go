// Package reorderargs implements the ReorderArgs obfuscation modifier.
//
// Technique: shuffle the order of command-line arguments. Many command-line
// parsers (getopt, flag, etc.) accept arguments in any order; reordering them
// defeats position-based signatures without changing program behaviour.
//
// ArgFuscator reference: src/Modifiers/ReorderArgs.ts
// Applies to token types: argument (and its associated value tokens)
package reorderargs

import (
	"encoding/json"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

func init() {
	modifiers.Register(&ReorderArgs{})
}

// ReorderArgs shuffles argument tokens (keeping flag–value pairs together).
type ReorderArgs struct{}

func (r *ReorderArgs) Name() string        { return "ReorderArgs" }
func (r *ReorderArgs) Description() string { return "Shuffle argument order (keeps flag–value pairs)" }

// Config holds ReorderArgs-specific config fields.
type Config struct {
	models.BaseModifierConfig
}

// Apply implements modifiers.Modifier.
//
// TODO: Implement this method.
//
// Steps:
//  1. Unmarshal cfg into a Config struct.
//  2. Parse Probability; if rand.Float64() >= probability, return unchanged.
//  3. Separate the command token (index 0) from the argument tokens.
//  4. Group argument tokens into (flag, value…) pairs using the ValueCount
//     information from the profile's ArgumentDefinitions.
//     (You may need to pass ArgumentDefinitions through the engine; consider
//     adding them to the Config or using a wrapper struct.)
//  5. Shuffle the pairs with rand.Shuffle.
//  6. Flatten back to a token slice: [command] + [shuffled pairs…].
//  7. Return updated tokens.
//
// Edge case: tokens that are not recognised flags should be treated as
// standalone argument groups (no associated value tokens).
func (r *ReorderArgs) Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error) {
	return tokens, modifiers.ErrNotImplemented
}
