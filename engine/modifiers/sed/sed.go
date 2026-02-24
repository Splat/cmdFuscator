// Package sed implements the Sed obfuscation modifier.
//
// Technique: apply sed-style substitution statements of the form
//
//	s/<char>/<replacement>/i
//
// to eligible token values. The substitution replaces single characters with
// visually similar Unicode lookalikes (e.g. 'a' → 'ᵃ', 'e' → 'ᵉ').
//
// ArgFuscator reference: src/Modifiers/Sed.ts
// Applies to token types: argument, value
package sed

import (
	"encoding/json"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

func init() {
	modifiers.Register(&Sed{})
}

// Sed applies sed-like character substitution rules to token values.
type Sed struct{}

func (s *Sed) Name() string        { return "Sed" }
func (s *Sed) Description() string { return "Replace chars with Unicode lookalikes via sed rules" }

// Config holds Sed-specific config fields.
type Config struct {
	models.BaseModifierConfig
	// SedStatements is a newline-delimited list of substitution rules.
	// Each rule has the form: s/<from>/<to>/i
	// The /i flag means case-insensitive matching.
	SedStatements string `json:"SedStatements"`
}

// Apply implements modifiers.Modifier.
//
// TODO: Implement this method.
//
// Steps:
//  1. Unmarshal cfg into a Config struct.
//  2. Parse Probability.
//  3. Parse Config.SedStatements: split on newlines, then parse each
//     "s/<from>/<to>/i" rule into a (from, to) pair.
//     - The delimiter after 's' can be any character (it's the char after 's').
//     - Parts: split the rule on the delimiter to get [from, to].
//     - The trailing /i flag means match both upper and lower case of <from>.
//  4. Build a substitution map: rune → replacement string.
//  5. For each eligible token:
//     a. Roll probability per character.
//     b. If triggered and the character has a substitution, apply it.
//  6. Return updated tokens.
//
// Example rule: "s/a/ᵃ/i" → replace 'a' or 'A' with 'ᵃ'
func (s *Sed) Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error) {
	return tokens, modifiers.ErrNotImplemented
}
