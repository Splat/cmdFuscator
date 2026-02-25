package charinsert

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"cmdFuscator/engine/modifiers"
	"cmdFuscator/models"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

// cfg builds a json.RawMessage from explicit field values so test cases stay
// readable without raw JSON strings everywhere.
func cfg(appliesTo []string, probability string, characters []string, offset string) json.RawMessage {
	c := Config{
		BaseModifierConfig: models.BaseModifierConfig{
			AppliesTo:   appliesTo,
			Probability: probability,
		},
		Characters: characters,
		Offset:     offset,
	}
	b, err := json.Marshal(c)
	if err != nil {
		panic("cfg helper: " + err.Error())
	}
	return b
}

// tok is a shorthand token constructor.
func tok(typ models.TokenType, val string) models.Token {
	return models.Token{Type: typ, Value: val}
}

// countInserted counts how many runes in got are not present (by position) in
// original.  It is used to verify that exactly one character was inserted.
func countInserted(original, got string) int {
	o := []rune(original)
	g := []rune(got)
	return len(g) - len(o)
}

// ─── modifier interface ───────────────────────────────────────────────────────

func TestName(t *testing.T) {
	m := &CharacterInsertion{}
	if m.Name() != "CharacterInsertion" {
		t.Errorf("Name() = %q, want %q", m.Name(), "CharacterInsertion")
	}
}

func TestDescription(t *testing.T) {
	m := &CharacterInsertion{}
	if m.Description() == "" {
		t.Error("Description() must not be empty")
	}
}

// ─── config unmarshalling ─────────────────────────────────────────────────────

func TestApply_InvalidJSON(t *testing.T) {
	m := &CharacterInsertion{}
	_, err := m.Apply(
		[]models.Token{tok(models.TokenTypeArgument, "-urlcache")},
		json.RawMessage(`not valid json`),
	)
	if err == nil {
		t.Fatal("Apply with invalid JSON config should return an error")
	}
}

var m = &CharacterInsertion{}
var input = []models.Token{
	tok(models.TokenTypeArgument, "-urlcache"),
	tok(models.TokenTypeValue, "output.bin"),
}

// ─── probability || offset = "invalid" (not an integer) ──────────────────────
func TestApply_InvalidProbabilityOffset(t *testing.T) {
	c := cfg(
		[]string{"argument", "value"},
		"0.0",
		[]string{"\u200c"}, // zero-width non-joiner
		"invalid",
	)
	_, err := m.Apply(input, c)

	var numError *strconv.NumError

	if !errors.As(err, &numError) {
		t.Errorf("expected *strconv.NumError, got %T: %v", err, err)
	}

	c = cfg(
		[]string{"argument", "value"},
		"invalid",
		[]string{"\u200c"}, // zero-width non-joiner
		"0",
	)
	_, err = m.Apply(input, c)
	if !errors.As(err, &numError) {
		t.Errorf("expected *strconv.NumError, got %T: %v", err, err)
	}
}

// ─── probability = 0.0 (never fires) ─────────────────────────────────────────

// When Probability is "0.0" no token should ever be modified, regardless of
// how many times Apply is called.
func TestApply_ProbabilityZero_NeverModifies(t *testing.T) {
	m := &CharacterInsertion{}
	input := []models.Token{
		tok(models.TokenTypeArgument, "-urlcache"),
		tok(models.TokenTypeValue, "output.bin"),
	}
	c := cfg(
		[]string{"argument", "value"},
		"0.0",
		[]string{"\u200c"}, // zero-width non-joiner
		"2",
	)

	for range 50 {
		got, err := m.Apply(input, c)
		if err != nil && !errors.Is(err, modifiers.ErrNotImplemented) {
			t.Fatalf("unexpected error: %v", err)
		}
		if errors.Is(err, modifiers.ErrNotImplemented) {
			t.Skip("CharacterInsertion.Apply not yet implemented")
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

// When Probability is "1.0" every eligible token must be modified.
func TestApply_ProbabilityOne_AlwaysModifies(t *testing.T) {
	m := &CharacterInsertion{}
	input := []models.Token{
		tok(models.TokenTypeArgument, "-urlcache"),
	}
	c := cfg(
		[]string{"argument"},
		"1.0",
		[]string{"\u200c"},
		"2",
	)

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("CharacterInsertion.Apply not yet implemented")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Value == input[0].Value {
		t.Errorf("Probability=1.0: eligible token was not modified")
	}
}

// ─── insertion count ──────────────────────────────────────────────────────────

// Exactly one character should be inserted per token per Apply call.
func TestApply_InsertsExactlyOneCharacter(t *testing.T) {
	m := &CharacterInsertion{}
	original := "-urlcache"
	input := []models.Token{tok(models.TokenTypeArgument, original)}
	c := cfg(
		[]string{"argument"},
		"1.0",
		[]string{"\u200c"},
		"2",
	)

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("CharacterInsertion.Apply not yet implemented")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := countInserted(original, got[0].Value)
	if n != 1 {
		t.Errorf("expected exactly 1 inserted character, got %d (value=%q)", n, got[0].Value)
	}
}

// ─── insertion offset ─────────────────────────────────────────────────────────

// The inserted character must appear at the rune position specified by Offset.
func TestApply_OffsetPosition(t *testing.T) {
	m := &CharacterInsertion{}
	ins := "\u200c" // the character we expect to find
	cases := []struct {
		name   string
		input  string
		offset string
		wantAt int // rune index where ins should appear after insertion
	}{
		{"offset 0", "-urlcache", "0", 0},
		{"offset 1", "-urlcache", "1", 1},
		{"offset 2", "-urlcache", "2", 2},
		{"offset past end clamps", "-hi", "99", 3}, // len("-hi") == 3; clamp to end
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := []models.Token{tok(models.TokenTypeArgument, tc.input)}
			c := cfg([]string{"argument"}, "1.0", []string{ins}, tc.offset)

			got, err := m.Apply(input, c)
			if errors.Is(err, modifiers.ErrNotImplemented) {
				t.Skip("CharacterInsertion.Apply not yet implemented")
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			runes := []rune(got[0].Value)
			if tc.wantAt >= len(runes) {
				t.Fatalf("result %q too short to have a rune at index %d", got[0].Value, tc.wantAt)
			}
			if string(runes[tc.wantAt]) != ins {
				t.Errorf("inserted char at wrong position: got rune[%d]=%q in %q, want %q",
					tc.wantAt, string(runes[tc.wantAt]), got[0].Value, ins)
			}
		})
	}
}

// ─── AppliesTo filtering ──────────────────────────────────────────────────────

// Tokens whose type is NOT in AppliesTo must be left unchanged.
func TestApply_RespectsAppliesTo(t *testing.T) {
	m := &CharacterInsertion{}

	// Only "argument" is eligible; "value" and "command" must not be touched.
	input := []models.Token{
		tok(models.TokenTypeCommand, "certutil.exe"),
		tok(models.TokenTypeArgument, "-urlcache"),
		tok(models.TokenTypeValue, "output.bin"),
	}
	c := cfg([]string{"argument"}, "1.0", []string{"\u200c"}, "2")

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("CharacterInsertion.Apply not yet implemented")
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
	// The argument token should have been modified (probability=1.0)
	if got[1].Value == input[1].Value {
		t.Errorf("argument token should have been modified")
	}
}

// ─── token count invariant ────────────────────────────────────────────────────

// Apply must never add or remove tokens — only mutate their values.
func TestApply_TokenCountUnchanged(t *testing.T) {
	m := &CharacterInsertion{}
	input := []models.Token{
		tok(models.TokenTypeCommand, "certutil.exe"),
		tok(models.TokenTypeArgument, "-urlcache"),
		tok(models.TokenTypeArgument, "-f"),
		tok(models.TokenTypeURL, "https://example.com"),
		tok(models.TokenTypePath, "out.bin"),
	}
	c := cfg([]string{"argument", "value"}, "1.0", []string{"\u200c"}, "1")

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("CharacterInsertion.Apply not yet implemented")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(input) {
		t.Errorf("token count changed: got %d, want %d", len(got), len(input))
	}
}

// ─── token types preserved ───────────────────────────────────────────────────

// Apply must never change a token's Type field, only its Value.
func TestApply_TokenTypesPreserved(t *testing.T) {
	m := &CharacterInsertion{}
	input := []models.Token{
		tok(models.TokenTypeCommand, "certutil.exe"),
		tok(models.TokenTypeArgument, "-urlcache"),
	}
	c := cfg([]string{"argument"}, "1.0", []string{"\u200c"}, "1")

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("CharacterInsertion.Apply not yet implemented")
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

// ─── character pool sampling ──────────────────────────────────────────────────

// With a pool of multiple characters and enough repetitions, at least two
// distinct characters should be sampled, confirming the pool is used randomly
// rather than always picking index 0.
func TestApply_SamplesFromCharacterPool(t *testing.T) {
	m := &CharacterInsertion{}
	pool := []string{"\u200c", "\u200d", "\u2060", "\u2061", "\u2062"}
	c := cfg([]string{"argument"}, "1.0", pool, "1")

	seen := map[string]bool{}
	for range 100 {
		input := []models.Token{tok(models.TokenTypeArgument, "-urlcache")}
		got, err := m.Apply(input, c)
		if errors.Is(err, modifiers.ErrNotImplemented) {
			t.Skip("CharacterInsertion.Apply not yet implemented")
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// The inserted character is at position Offset=1.
		runes := []rune(got[0].Value)
		seen[string(runes[1])] = true
	}
	if len(seen) < 2 {
		t.Errorf("only 1 distinct character sampled from pool of %d over 100 runs; expected random sampling", len(pool))
	}
}

// ─── empty characters pool ────────────────────────────────────────────────────

// An empty Characters slice is a degenerate config; Apply should either skip
// the token or return an error — but must not panic.
func TestApply_EmptyCharactersPool_DoesNotPanic(t *testing.T) {
	m := &CharacterInsertion{}
	input := []models.Token{tok(models.TokenTypeArgument, "-urlcache")}
	c := cfg([]string{"argument"}, "1.0", []string{}, "2")

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Apply panicked with empty Characters pool: %v", r)
		}
	}()

	_, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("CharacterInsertion.Apply not yet implemented")
	}
	// Any non-panic result (including an error) is acceptable here.
	_ = err
}

// ─── multibyte characters ─────────────────────────────────────────────────────

// Characters in the pool may be multi-byte UTF-8. The modifier must treat them
// as single runes, not individual bytes.
func TestApply_MultibyteCharacterInsertedAsOneRune(t *testing.T) {
	m := &CharacterInsertion{}
	ins := "\u200c" // 3-byte UTF-8 sequence
	input := []models.Token{tok(models.TokenTypeArgument, "-urlcache")}
	c := cfg([]string{"argument"}, "1.0", []string{ins}, "2")

	got, err := m.Apply(input, c)
	if errors.Is(err, modifiers.ErrNotImplemented) {
		t.Skip("CharacterInsertion.Apply not yet implemented")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	n := countInserted(input[0].Value, got[0].Value)
	if n != 1 {
		t.Errorf("expected 1 rune inserted, got %d (possible byte-level insertion)", n)
	}

	// Confirm the inserted rune is exactly our character
	if !strings.ContainsRune(got[0].Value, []rune(ins)[0]) {
		t.Errorf("result %q does not contain inserted rune %U", got[0].Value, []rune(ins)[0])
	}
}
