// Package quoteinsert implements the QuoteInsertion obfuscation modifier.
//
// Technique: insert an empty pair of quotes ("" or '') at a random position
// inside a token's value. The shell ignores the empty string, but the literal
// command text looks different to signature scanners.
//
// Example:  -urlcache  →  -url""cache  or  -ur''lcache
//
// ArgFuscator reference: src/Modifiers/QuoteInsertion.ts
// Applies to token types: path, url, argument, value
package quoteinsert

import (
	"encoding/json"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

func init() {
	modifiers.Register(&QuoteInsertion{})
}

// QuoteInsertion inserts empty quote pairs inside token values.
type QuoteInsertion struct{}

func (q *QuoteInsertion) Name() string        { return "QuoteInsertion" }
func (q *QuoteInsertion) Description() string { return "Insert empty quote pairs inside tokens" }

// Config holds QuoteInsertion-specific config fields.
type Config struct {
	models.BaseModifierConfig
}

// Apply implements modifiers.Modifier.
//
// TODO: Implement this method.
//
// Steps:
//  1. Unmarshal cfg into a Config struct.
//  2. Parse Config.Probability.
//  3. For each eligible token:
//     a. Roll rand.Float64(); if >= probability, leave token unchanged.
//     b. Pick a random insertion position between 1 and len(runes)-1
//        (avoid position 0 or end to keep the token visually meaningful).
//     c. Pick a quote character at random: `"` or `'`.
//     d. Insert `""` (or `''`) at the chosen position.
//  4. Return updated tokens.
func (q *QuoteInsertion) Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error) {
	return tokens, modifiers.ErrNotImplemented
}
