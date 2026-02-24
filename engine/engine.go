// Package engine orchestrates the obfuscation pipeline:
//
//  1. Parse  – turn a raw command string into a typed []models.Token
//  2. Modify – apply each enabled Modifier in registration order
//  3. Render – join the modified tokens back into an output string
//
// The Parse and Render steps are stubbed; implement them here once you are
// comfortable with the modifier implementations.
package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"

	// Side-effect imports register all built-in modifiers at package init time.
	_ "cmdFuscator/engine/modifiers/all"
)

// Engine is the top-level obfuscation coordinator. Create one with New() and
// reuse it across calls — it is safe for concurrent use once constructed.
type Engine struct{}

// New returns a ready-to-use Engine. All modifiers registered via
// modifiers.Register() (typically via init() in each modifier file) are
// available automatically.
func New() *Engine {
	return &Engine{}
}

// ObfuscateResult holds both the output command and a per-modifier summary
// so the TUI can show which techniques were actually applied.
type ObfuscateResult struct {
	Output  string
	Applied []string // names of modifiers that ran without error
	Skipped []string // names of modifiers that returned ErrNotImplemented
	Errors  map[string]error
}

// Obfuscate runs the full pipeline against command using the first profile in pf
// that matches the host platform (or the first profile if none match).
//
// enabled is a set of modifier names the user has toggled on in the TUI;
// modifiers absent from the map, or mapped to false, are skipped.
func (e *Engine) Obfuscate(command string, pf *models.ProfileFile, enabled map[string]bool) (ObfuscateResult, error) {
	if pf == nil || len(pf.Profiles) == 0 {
		return ObfuscateResult{}, errors.New("engine: no profiles available")
	}

	profile := pickProfile(pf)

	// ── Step 1: Tokenize ─────────────────────────────────────────────────────
	// TODO: implement Tokenize in tokenize.go.
	// It should use profile.Parameters.Arguments to identify flags and their
	// value counts, then classify each whitespace-separated token.
	tokens, err := Tokenize(command, profile)
	if err != nil {
		return ObfuscateResult{}, fmt.Errorf("engine: tokenize: %w", err)
	}

	// ── Step 2: Apply modifiers ───────────────────────────────────────────────
	result := ObfuscateResult{Errors: make(map[string]error)}

	for _, mod := range modifiers.All() {
		if !enabled[mod.Name()] {
			continue
		}

		rawCfg, hasCfg := profile.Parameters.Modifiers[mod.Name()]
		if !hasCfg {
			// Profile does not define this modifier; silently skip.
			continue
		}

		modified, err := mod.Apply(tokens, rawCfg)
		if err != nil {
			if errors.Is(err, modifiers.ErrNotImplemented) {
				result.Skipped = append(result.Skipped, mod.Name())
			} else {
				result.Errors[mod.Name()] = err
			}
			// Leave tokens unchanged and continue with remaining modifiers.
			continue
		}

		tokens = modified
		result.Applied = append(result.Applied, mod.Name())
	}

	// ── Step 3: Render ────────────────────────────────────────────────────────
	// TODO: implement Render in render.go.
	// It should reconstruct quoting and spacing correctly rather than just
	// joining with spaces.
	result.Output = Render(tokens)

	return result, nil
}

// ─── Stubs (implement these) ──────────────────────────────────────────────────

// Tokenize parses a raw command string into a slice of typed tokens.
//
// TODO: Implement this function.
//
// Guidance:
//   - Split on whitespace (but respect quoted strings).
//   - The first token is always TokenTypeCommand.
//   - Use profile.Parameters.Arguments to identify known flags; the token
//     immediately following a flag with ValueCount > 0 is TokenTypeValue.
//   - Tokens that look like file paths (contain / or \) → TokenTypePath.
//   - Tokens that start with http:// or https:// → TokenTypeURL.
//   - Everything else → TokenTypeArgument or TokenTypeValue depending on context.
func Tokenize(command string, profile models.Profile) ([]models.Token, error) {
	// Minimal fallback: split on whitespace, label first token as command,
	// rest as argument. Replace this with a proper implementation.
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, errors.New("tokenize: empty command")
	}

	tokens := make([]models.Token, len(parts))
	tokens[0] = models.Token{Type: models.TokenTypeCommand, Value: parts[0]}
	for i, p := range parts[1:] {
		tokens[i+1] = models.Token{Type: models.TokenTypeArgument, Value: p}
	}

	return tokens, nil
}

// Render joins a token slice back into a command string.
//
// TODO: Implement this function.
//
// Guidance:
//   - Insert spaces between tokens (matching the original spacing where possible).
//   - Re-apply quoting when a token value contains spaces.
//   - This is the inverse of Tokenize; the round-trip should be lossless for
//     unmodified tokens.
func Render(tokens []models.Token) string {
	parts := make([]string, len(tokens))
	for i, t := range tokens {
		parts[i] = t.Value
	}
	return strings.Join(parts, " ")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// pickProfile selects the most relevant Profile from a ProfileFile.
// Currently returns the first profile; extend this to match the host OS if desired.
func pickProfile(pf *models.ProfileFile) models.Profile {
	return pf.Profiles[0]
}

// ModifierSummary returns a []ModifierInfo describing all registered modifiers
// and whether each one is enabled, for use by the TUI options panel.
func ModifierSummary(enabled map[string]bool) []ModifierInfo {
	all := modifiers.All()
	out := make([]ModifierInfo, len(all))
	for i, m := range all {
		out[i] = ModifierInfo{
			Name:        m.Name(),
			Description: m.Description(),
			Enabled:     enabled[m.Name()],
		}
	}
	return out
}

// ModifierInfo is a view-model used by the TUI to render the options panel.
type ModifierInfo struct {
	Name        string
	Description string
	Enabled     bool
}

// DefaultEnabled returns a map with every registered modifier enabled.
// Call this when a new executable is selected in the TUI to reset options.
func DefaultEnabled(pf *models.ProfileFile) map[string]bool {
	m := make(map[string]bool)
	if pf == nil || len(pf.Profiles) == 0 {
		return m
	}
	profile := pickProfile(pf)
	for name := range profile.Parameters.Modifiers {
		m[name] = true
	}
	return m
}

// ConfigFor extracts and unmarshals the modifier config for the given name from
// a profile. Returns the raw json.RawMessage; the modifier itself is responsible
// for unmarshaling into its own config struct.
func ConfigFor(profile models.Profile, modifierName string) (json.RawMessage, bool) {
	raw, ok := profile.Parameters.Modifiers[modifierName]
	return raw, ok
}
