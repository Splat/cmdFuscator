// Package regex implements the Regex obfuscation modifier.
//
// Technique: apply regular-expression based find-and-replace operations to
// eligible token values. The replacement patterns are defined in the profile's
// modifier config.
//
// ArgFuscator reference: src/Modifiers/Regex.ts
// Applies to token types: argument, value
package regex

import (
	"encoding/json"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

func init() {
	modifiers.Register(&Regex{})
}

// Regex applies profile-defined regex substitutions to token values.
type Regex struct{}

func (r *Regex) Name() string        { return "Regex" }
func (r *Regex) Description() string { return "Apply regex find-and-replace substitutions" }

// Rule is a single regex substitution rule.
type Rule struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
}

// Config holds Regex-specific config fields.
// Inspect actual profile JSON to verify the exact field names.
type Config struct {
	models.BaseModifierConfig
	Rules []Rule `json:"rules"`
}

// Apply implements modifiers.Modifier.
//
// TODO: Implement this method.
//
// Steps:
//  1. Unmarshal cfg into a Config struct.
//  2. Parse Probability.
//  3. Compile each Rule.Pattern with regexp.MustCompile (or Compile + handle error).
//  4. For each eligible token:
//     a. Roll probability; skip if not triggered.
//     b. Apply each compiled regex in order using regexp.Regexp.ReplaceAllString.
//  5. Return updated tokens.
//
// Note: since the Regex modifier config schema is partially inferred, you may
// need to adjust the Config struct after inspecting real profile files that use it.
func (r *Regex) Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error) {
	return tokens, modifiers.ErrNotImplemented
}
