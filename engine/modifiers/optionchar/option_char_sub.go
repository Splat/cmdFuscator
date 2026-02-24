// Package optionchar implements the OptionCharSubstitution obfuscation modifier.
//
// Technique: replace the leading option character of a flag (typically '-' or '/')
// with a visually similar Unicode character drawn from the profile's
// OutputOptionChars list.
//
// Example (Windows):  -urlcache  →  /urlcache  or  –urlcache  (en-dash)
//
// ArgFuscator reference: src/Modifiers/OptionCharSubstitution.ts
// Applies to token types: argument, url, value
package optionchar

import (
	"encoding/json"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

func init() {
	modifiers.Register(&OptionCharSubstitution{})
}

// OptionCharSubstitution swaps the leading flag character for an alternative.
type OptionCharSubstitution struct{}

func (o *OptionCharSubstitution) Name() string { return "OptionCharSubstitution" }
func (o *OptionCharSubstitution) Description() string {
	return "Replace - or / with a lookalike Unicode option char"
}

// Config holds OptionCharSubstitution-specific config fields.
type Config struct {
	models.BaseModifierConfig
	// OutputOptionChars is the set of replacement characters to choose from,
	// e.g. ["/", "-", "–", "—", "−"].
	OutputOptionChars []string `json:"OutputOptionChars"`
}

// Apply implements modifiers.Modifier.
//
// TODO: Implement this method.
//
// Steps:
//  1. Unmarshal cfg into a Config struct.
//  2. Parse Probability.
//  3. For each eligible token:
//     a. Check whether the first rune is '-' or '/'.
//     b. Roll rand.Float64(); if < probability, pick a random entry from
//        Config.OutputOptionChars and replace the leading character.
//  4. Return updated tokens.
//
// Note: some entries in OutputOptionChars are multi-byte UTF-8; use []rune
// indexing rather than []byte to avoid corrupting multi-byte characters.
func (o *OptionCharSubstitution) Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error) {
	return tokens, modifiers.ErrNotImplemented
}
