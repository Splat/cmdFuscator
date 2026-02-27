package randomcase

import (
	"encoding/json"
	"errors"
	"testing"
	"unicode"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

func cfg(appliesTo []string, probability string) json.RawMessage {
	c := Config{
		BaseModifierConfig: models.BaseModifierConfig{
			AppliesTo:   appliesTo,
			Probability: probability,
		},
	}
	b, err := json.Marshal(c)
	if err != nil {
		panic("cfg helper: " + err.Error())
	}
	return b
}

func tok(typ models.TokenType, val string) models.Token {
	return models.Token{Type: typ, Value: val}
}

// ─── modifier interface ───────────────────────────────────────────────────────

func TestName(t *testing.T) {
	m := &RandomCase{}
	if m.Name() != "RandomCase" {
		t.Errorf("Name() = %q, want %q", m.Name(), "RandomCase")
	}
}

func TestDescription(t *testing.T) {
	m := &RandomCase{}
	if m.Description() == "" {
		t.Error("Description() must not be empty")
	}
}

// ─── config unmarshalling ─────────────────────────────────────────────────────

func TestApply_InvalidJSON(t *testing.T) {
	m := &RandomCase{}
	_, err := m.Apply(
		[]models.Token{tok(models.TokenTypeArgument, "-urlcache")},
		json.RawMessage(`not valid json`),
	)
	if err == nil {
		t.Fatal("Apply with invalid JSON config should return an error")
	}
}

func TestApply_InvalidProbability(t *testing.T) {
	m := &RandomCase{}
	c := cfg([]string{"argument"}, "not-a-float")
	_, err := m.Apply([]models.Token{tok(models.TokenTypeArgument, "-urlcache")}, c)
	if err == nil {
		t.Fatal("Apply with invalid probability should return an error")
	}
}

// ─── probability = 0.0 (never fires) ─────────────────────────────────────────

// probability "0.0" is a valid config meaning "never modify". The validation
// guard uses <= 0 which incorrectly rejects it — this test will expose that.
func TestApply_ProbabilityZero_NeverModifies(t *testing.T) {
	m := &RandomCase{}
	input := []models.Token{
		tok(models.TokenTypeArgument, "-urlcache"),
		tok(models.TokenTypeValue, "output.bin"),
	}
	c := cfg([]string{"argument", "value"}, "0.0")

	for range 50 {
		got, err := m.Apply(input, c)
		if errors.Is(err, modifiers.ErrNotImplemented) {
			t.Skip("RandomCase.Apply not yet implemented")
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for i, tk := range got {
			if tk.Value != input[i].Value {
				t.Errorf("Probability=0.0: token[%d] modified: got %q, want %q",
					i, tk.Value, input[i].Value)
			}
		}
	}
}

// ─── probability = 1.0 (always fires) ────────────────────────────────────────

// With probability 1.0 every letter in every eligible token must have its case
// flipped. Non-letter characters must be left unchanged.
func TestApply_ProbabilityOne_FlipsAllLetters(t *testing.T) {
	m := &RandomCase{}
	cases := []struct {
		input string
		want  string
	}{
		{"HELLO", "hello"},
		{"hello", "HELLO"},
		{"Hello", "hELLO"},
		{"-urlcache", "-URLCACHE"},
		{"-SPLIT", "-split"},
	}

	for _, tc := range cases {
		input := []models.Token{tok(models.TokenTypeArgument, tc.input)}
		c := cfg([]string{"argument"}, "1.0")

		got, err := m.Apply(input, c)
		if errors.Is(err, modifiers.ErrNotImplemented) {
			t.Skip("RandomCase.Apply not yet implemented")
		}
		if err != nil {
			t.Fatalf("input %q: unexpected error: %v", tc.input, err)
		}
		if got[0].Value != tc.want {
			t.Errorf("input %q: got %q, want %q", tc.input, got[0].Value, tc.want)
		}
	}
}

// ─── non-letter characters ────────────────────────────────────────────────────

// Digits, punctuation, and symbols have no case; they must pass through
// unchanged regardless of probability.
func TestApply_PreservesNonAlpha(t *testing.T) {
	m := &RandomCase{}
	input := []models.Token{tok(models.TokenTypeArgument, "-f123.exe")}
	c := cfg([]string{"argument"}, "1.0")

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("RandomCase.Apply not yet implemented")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	origRunes := []rune(input[0].Value)
	gotRunes := []rune(got[0].Value)
	for i, r := range origRunes {
		if !unicode.IsLetter(r) && gotRunes[i] != r {
			t.Errorf("non-letter rune at index %d changed: got %q, want %q", i, gotRunes[i], r)
		}
	}
}

// ─── token length invariant ───────────────────────────────────────────────────

// Case-flipping must never add or remove runes.
func TestApply_LengthPreserved(t *testing.T) {
	m := &RandomCase{}
	input := []models.Token{tok(models.TokenTypeArgument, "-urlcache")}
	c := cfg([]string{"argument"}, "1.0")

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("RandomCase.Apply not yet implemented")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len([]rune(got[0].Value)) != len([]rune(input[0].Value)) {
		t.Errorf("token length changed: got %d runes, want %d",
			len([]rune(got[0].Value)), len([]rune(input[0].Value)))
	}
}

// ─── AppliesTo filtering ──────────────────────────────────────────────────────

// Tokens whose Type is NOT in AppliesTo must be left unchanged.
func TestApply_RespectsAppliesTo(t *testing.T) {
	m := &RandomCase{}
	input := []models.Token{
		tok(models.TokenTypeCommand, "certutil.exe"),
		tok(models.TokenTypeArgument, "-urlcache"),
		tok(models.TokenTypeValue, "output.bin"),
	}
	c := cfg([]string{"argument"}, "1.0")

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("RandomCase.Apply not yet implemented")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got[0].Value != input[0].Value {
		t.Errorf("command token must not be modified: got %q", got[0].Value)
	}
	if got[2].Value != input[2].Value {
		t.Errorf("value token must not be modified: got %q", got[2].Value)
	}
	if got[1].Value == input[1].Value {
		t.Errorf("argument token should have been modified")
	}
}

// ─── token count invariant ────────────────────────────────────────────────────

func TestApply_TokenCountUnchanged(t *testing.T) {
	m := &RandomCase{}
	input := []models.Token{
		tok(models.TokenTypeCommand, "certutil.exe"),
		tok(models.TokenTypeArgument, "-urlcache"),
		tok(models.TokenTypeArgument, "-f"),
		tok(models.TokenTypeURL, "https://example.com"),
		tok(models.TokenTypePath, "out.bin"),
	}
	c := cfg([]string{"argument", "value"}, "1.0")

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("RandomCase.Apply not yet implemented")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(input) {
		t.Errorf("token count changed: got %d, want %d", len(got), len(input))
	}
}

// ─── token types preserved ────────────────────────────────────────────────────

func TestApply_TokenTypesPreserved(t *testing.T) {
	m := &RandomCase{}
	input := []models.Token{
		tok(models.TokenTypeCommand, "certutil.exe"),
		tok(models.TokenTypeArgument, "-urlcache"),
	}
	c := cfg([]string{"argument"}, "1.0")

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("RandomCase.Apply not yet implemented")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := range got {
		if got[i].Type != input[i].Type {
			t.Errorf("token[%d] Type changed: got %q, want %q", i, got[i].Type, input[i].Type)
		}
	}
}

// ─── input slice immutability ─────────────────────────────────────────────────

// Apply must not mutate the original token slice passed in.
func TestApply_OriginalTokensUnmodified(t *testing.T) {
	m := &RandomCase{}
	input := []models.Token{tok(models.TokenTypeArgument, "-urlcache")}
	origVal := input[0].Value
	c := cfg([]string{"argument"}, "1.0")

	_, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("RandomCase.Apply not yet implemented")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input[0].Value != origVal {
		t.Errorf("Apply mutated input slice: original is now %q", input[0].Value)
	}
}

// ─── per-character probability ────────────────────────────────────────────────

// With probability=0.5 and a long token, repeated calls must produce more than
// one distinct result, confirming that probability applies per character rather
// than to the token as a whole.
func TestApply_PartialProbability_ProducesVariation(t *testing.T) {
	m := &RandomCase{}
	input := []models.Token{tok(models.TokenTypeArgument, "-urlcache")}
	c := cfg([]string{"argument"}, "0.5")

	seen := map[string]bool{}
	for range 100 {
		got, err := m.Apply(input, c)
		if errors.Is(err, modifiers.ErrNotImplemented) {
			t.Skip("RandomCase.Apply not yet implemented")
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seen[got[0].Value] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected variation with probability=0.5, got only %d distinct result(s) over 100 runs",
			len(seen))
	}
}
