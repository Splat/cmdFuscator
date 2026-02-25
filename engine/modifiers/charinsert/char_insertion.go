// Package charinsert implements the CharacterInsertion obfuscation modifier.
//
// Technique: insert one or more invisible or non-printing Unicode characters
// (from the profile's Characters list) at a configurable offset within each
// eligible token. The parser on the target system ignores these characters, but
// they defeat string-match signatures.
//
// ArgFuscator reference: src/Modifiers/CharacterInsertion.ts
// Applies to token types: argument, value
package charinsert

import (
	"encoding/json"
	"fmt"
	"strconv"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

func init() {
	modifiers.Register(&CharacterInsertion{})
}

// CharacterInsertion inserts invisible Unicode codepoints into token values.
type CharacterInsertion struct{}

func (c *CharacterInsertion) Name() string { return "CharacterInsertion" }
func (c *CharacterInsertion) Description() string {
	return "Insert invisible Unicode characters into tokens"
}

// Config holds CharacterInsertion-specific config fields.
type Config struct {
	models.BaseModifierConfig
	// Characters is the pool of Unicode characters to sample from.
	// Each entry is a single-character string (possibly multi-byte UTF-8).
	Characters []string `json:"Characters"`
	// Offset is a string integer controlling insertion position within the token.
	// "2" means insert after the 2nd character.
	Offset string `json:"Offset"`
}

// Apply implements modifiers.Modifier.
//
// TODO: Implement this method.
//
// Steps:
//  1. Unmarshal cfg into a Config struct. (DONE)
//  2. Parse Probability and Offset (strconv.Atoi for Offset).
//  3. For each eligible token:
//     a. Roll probability; skip if not triggered.
//     b. Pick a random character from Config.Characters.
//     c. Insert it at position Offset within the rune slice of token.Value
//     (clamp Offset to len(runes) if the token is shorter).
//  4. Return updated tokens.
//
// Safety note: the Characters list in the profiles can be very long (hundreds
// of entries). Use rand.Intn(len(cfg.Characters)) to pick one.
func (c *CharacterInsertion) Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error) {
	cfgM := &Config{}
	if err := json.Unmarshal(cfg, cfgM); err != nil {
		return tokens, fmt.Errorf("unmarshal config: %w", err)
	}

	if probability, err := strconv.ParseFloat(cfgM.Probability, 64); err != nil {
		return tokens, fmt.Errorf("parse probability: %w", err)
	} else if probability < 0 || probability > 1 {
		return tokens, fmt.Errorf("probability must be between 0 and 1: %w", err)
	} else {
		fmt.Println("probability:", probability)
	}

	if offset, err := strconv.Atoi(cfgM.Offset); err != nil {
		return tokens, fmt.Errorf("parse offset: %w", err)
	} else {
		fmt.Println("offset:", offset)
	}

	return tokens, modifiers.ErrNotImplemented
}
