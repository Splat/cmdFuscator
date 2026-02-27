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
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"unicode"

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
//     (use unicode.ToUpper / unicode.ToLower as appropriate).
//     c. Rebuild token.Value from the modified runes.
//  4. Return the updated token slice.
//
// Hint: unicode.IsUpper(r) / unicode.IsLower(r) tell you the current case.
// Hint: use a strings.Builder or []rune for efficient string construction.
func (r *RandomCase) Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error) {
	out := make([]models.Token, len(tokens)) // the eventual return value
	copy(out, tokens)                        // make a copy of the input tokens for no op situations

	cfgM := &Config{}
	if err := json.Unmarshal(cfg, cfgM); err != nil {
		return tokens, fmt.Errorf("unmarshal config: %w", err)
	}

	probability, err := strconv.ParseFloat(cfgM.Probability, 64)
	if err != nil {
		return tokens, fmt.Errorf("parse probability: %w", err)
	} else if probability < 0 || probability > 1 {
		return tokens, fmt.Errorf("probability must be between 0 and 1")
	}

	for idx := range tokens {
		if !slices.Contains(cfgM.AppliesTo, string(tokens[idx].Type)) {
			continue // only apply to tokens of the specified types from config
		}
		runes := []rune(tokens[idx].Value)
		for charIdx, r := range runes {
			if rand.Float64() < probability { // flip this character's case with given probability
				if unicode.IsUpper(r) {
					runes[charIdx] = unicode.ToLower(runes[charIdx])
				} else {
					runes[charIdx] = unicode.ToUpper(runes[charIdx])
				}
			}
		}
		out[idx].Value = string(runes) // rebuild the token with the modified runes as a string
	}

	return out, nil
}
