// Package randomcase implements the RandomCase obfuscation modifier.
//
// Technique: for each character in an eligible token, flip its case
// (upper → lower, lower → upper) with the probability specified in the profile.
//
// ArgFuscator reference: src/Modifiers/RandomCase.ts
// Applies to token types: command, argument, value, path
package randomcase

import (
	"encoding/json"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

func init() {
	modifiers.Register(&RandomCase{})
}

// RandomCase flips the case of individual characters with a given probability.
type RandomCase struct{}

func (r *RandomCase) Name() string        { return "RandomCase" }
func (r *RandomCase) Description() string { return "Randomly flip UPPER/lower case per character" }

// Config holds the config fields for this modifier. Embed BaseModifierConfig to
// pick up AppliesTo and Probability automatically.
type Config struct {
	models.BaseModifierConfig
}

// Apply implements modifiers.Modifier.
//
// TODO: Implement this method.
//
// Steps:
//  1. Unmarshal cfg into a Config struct.
//  2. Parse Config.Probability with strconv.ParseFloat.
//  3. For each token whose Type is in Config.AppliesTo:
//     a. Iterate over each rune in token.Value.
//     b. Call rand.Float64(); if < probability, flip the rune's case
//        (use unicode.ToUpper / unicode.ToLower as appropriate).
//     c. Rebuild token.Value from the modified runes.
//  4. Return the updated token slice.
//
// Hint: unicode.IsUpper(r) / unicode.IsLower(r) tell you the current case.
// Hint: use a strings.Builder or []rune for efficient string construction.
func (r *RandomCase) Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error) {
	return tokens, modifiers.ErrNotImplemented
}
