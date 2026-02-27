// Package modifiers defines the Modifier interface and the global registry
// used by the engine to discover and apply obfuscation techniques.
//
// To add a new modifier:
//  1. Create a new file (e.g. my_technique.go) in this package.
//  2. Define a struct that implements the Modifier interface.
//  3. Call Register(New<MyTechnique>()) in an init() function in that file.
package modifiers

import (
	"encoding/json"
	"fmt"

	"cmdFuscator/models"
)

// ─── Interface ────────────────────────────────────────────────────────────────

// Modifier is the contract every obfuscation technique must satisfy.
//
// The engine iterates over a ordered list of registered Modifiers, calls
// CanApply to decide whether the modifier should run for the current profile,
// and then calls Apply with the current token slice and the raw JSON config
// extracted from the profile.
type Modifier interface {
	// Name returns the exact key used in the JSON profile's "modifiers" object,
	// e.g. "RandomCase". The registry is keyed on this value.
	Name() string

	// Description is a short human-readable summary shown in the TUI options panel.
	Description() string

	// Apply transforms tokens according to the technique's rules.
	// cfg is the raw JSON config for this modifier from the profile; unmarshal
	// it into a modifier-specific struct that embeds models.BaseModifierConfig.
	// Return the (possibly modified) token slice and any error.
	Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error)
}

// ─── Registry ─────────────────────────────────────────────────────────────────

// registry holds all modifiers indexed by Name().
var registry = map[string]Modifier{}

// order preserves insertion order so the engine applies modifiers consistently.
var order []string

// Register adds a Modifier to the global registry. It panics if a modifier with
// the same name has already been registered (catches copy-paste mistakes at
// startup rather than silently at runtime).
func Register(m Modifier) {
	if _, exists := registry[m.Name()]; exists {
		panic(fmt.Sprintf("modifiers: duplicate registration for %q", m.Name()))
	}
	registry[m.Name()] = m
	order = append(order, m.Name())
}

// All returns every registered Modifier in registration order.
func All() []Modifier {
	out := make([]Modifier, 0, len(order))
	for _, name := range order {
		out = append(out, registry[name])
	}
	return out
}

// Get looks up a Modifier by name. The second return value is false when the
// name is not registered.
func Get(name string) (Modifier, bool) {
	m, ok := registry[name]
	return m, ok
}

// ParseConfig parses a modifier config from a raw JSON message.
func ParseConfig(raw json.RawMessage) (models.BaseModifierConfig, error) {
	var cfg models.BaseModifierConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// ─── Sentinel error ───────────────────────────────────────────────────────────

// ErrNotImplemented is returned by stub Apply() methods to signal that the
// user has not yet written the implementation for that modifier.
var ErrNotImplemented = fmt.Errorf("not implemented")
