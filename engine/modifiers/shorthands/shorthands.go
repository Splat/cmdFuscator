// Package shorthands implements the Shorthands obfuscation modifier.
//
// Technique: replace a full flag name with a unique prefix abbreviation that
// the target executable still accepts. For example, PowerShell accepts
// "-NonInteractive" → "-NonI", "-No", "-N" (as long as the prefix is unambiguous).
//
// ArgFuscator reference: src/Modifiers/Shorthands.ts
// Applies to token types: argument
package shorthands

import (
	"encoding/json"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

func init() {
	modifiers.Register(&Shorthands{})
}

// Shorthands replaces long flag names with their shortest unambiguous prefix.
type Shorthands struct{}

func (s *Shorthands) Name() string        { return "Shorthands" }
func (s *Shorthands) Description() string { return "Abbreviate flags to shortest unambiguous prefix" }

// Config holds Shorthands-specific config fields.
type Config struct {
	models.BaseModifierConfig
}

// Apply implements modifiers.Modifier.
//
// TODO: Implement this method.
//
// Steps:
//  1. Unmarshal cfg into a Config struct.
//  2. Parse Probability.
//  3. Build an index of all known flags from the profile's Arguments list
//     (you will need to pass the profile through or pre-process it; consider
//     whether the profile should be injected via a constructor or method).
//  4. For each eligible argument token:
//     a. Strip the leading option char (-, /, --).
//     b. Find all known flags that start with the same prefix.
//     c. If exactly one flag matches a given prefix, that prefix is unambiguous.
//     d. Roll probability; if triggered, replace the token value with the
//        shortest unambiguous prefix (re-adding the original option char).
//  5. Return updated tokens.
//
// Design note: you may need to refactor the Apply signature or store the
// ArgumentDefinition list on the struct to make the profile data available here.
func (s *Shorthands) Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error) {
	return tokens, modifiers.ErrNotImplemented
}
