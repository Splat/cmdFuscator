// Package models defines the data structures that map to the ArgFuscator JSON
// profile format (format version 2.0). These types are shared across the loader,
// engine, and any future de-obfuscator packages.
package models

import "encoding/json"

// ─── Token ───────────────────────────────────────────────────────────────────

// TokenType is the semantic role of a single element in a command line.
// It drives which modifiers are allowed to act on each token.
type TokenType string

const (
	TokenTypeCommand  TokenType = "command"  // the executable itself
	TokenTypeArgument TokenType = "argument" // a flag/switch, e.g. -urlcache
	TokenTypeValue    TokenType = "value"    // a plain value for a preceding argument
	TokenTypePath     TokenType = "path"     // a file-system path
	TokenTypeURL      TokenType = "url"      // a URL
)

// Token is the unit that the engine and modifiers operate on.
// The parser produces a []Token from a raw command string; the renderer
// joins them back into an output string after modification.
type Token struct {
	Type  TokenType
	Value string
}

// ─── Profile file ─────────────────────────────────────────────────────────────

// ProfileFile is the root of a single JSON model file.
// The Name field is derived from the filename (e.g. "certutil" from certutil.json)
// and is not present in the JSON itself.
type ProfileFile struct {
	Name     string    `json:"-"`
	Versions Versions  `json:"versions"`
	Profiles []Profile `json:"profiles"`
}

// Versions holds format metadata from the JSON file header.
type Versions struct {
	ArgFuscator string `json:"argfuscator"`
	Format      string `json:"format"`
}

// ─── Profile ──────────────────────────────────────────────────────────────────

// Profile is a single OS/version-specific configuration for one executable.
// A ProfileFile may contain multiple Profiles (one per OS or version).
type Profile struct {
	ExecutableVersion     string            `json:"executableVersion"`
	Platform              string            `json:"platform"`              // "windows" | "linux" | "macos"
	OperatingSystem       string            `json:"operatingSystem"`       // "Windows" | "Ubuntu" | "macOS"
	OperatingSystemVersion string           `json:"operatingSystemVersion"`
	Alias                 []string          `json:"alias"`
	Parameters            ProfileParameters `json:"parameters"`
}

// ProfileParameters bundles the command template, known arguments, and modifier
// configuration for a single Profile.
type ProfileParameters struct {
	// Command is the example command template; each element is a typed token.
	Command []CommandElement `json:"command"`

	// Arguments lists the known flags for this executable and how many values
	// each flag consumes. Used by the Shorthands modifier and the tokenizer.
	Arguments []ArgumentDefinition `json:"arguments"`

	// Modifiers maps modifier name (e.g. "RandomCase") to its raw JSON config.
	// Using json.RawMessage lets each modifier unmarshal its own extra fields
	// without requiring a union type here.
	Modifiers map[string]json.RawMessage `json:"modifiers"`
}

// ─── Command element ──────────────────────────────────────────────────────────

// CommandElement represents one token in the command template.
// Exactly one of the five fields will be non-empty; the rest are zero.
type CommandElement struct {
	Command  string `json:"command,omitempty"`
	Argument string `json:"argument,omitempty"`
	Value    string `json:"value,omitempty"`
	Path     string `json:"path,omitempty"`
	URL      string `json:"url,omitempty"`
}

// Type returns the TokenType that corresponds to whichever field is set.
func (c CommandElement) Type() TokenType {
	switch {
	case c.Command != "":
		return TokenTypeCommand
	case c.Argument != "":
		return TokenTypeArgument
	case c.Path != "":
		return TokenTypePath
	case c.URL != "":
		return TokenTypeURL
	default:
		return TokenTypeValue
	}
}

// StringValue returns the non-empty string value of this element.
func (c CommandElement) StringValue() string {
	switch {
	case c.Command != "":
		return c.Command
	case c.Argument != "":
		return c.Argument
	case c.Path != "":
		return c.Path
	case c.URL != "":
		return c.URL
	default:
		return c.Value
	}
}

// ToToken converts this CommandElement into a Token.
func (c CommandElement) ToToken() Token {
	return Token{Type: c.Type(), Value: c.StringValue()}
}

// ─── Argument definition ──────────────────────────────────────────────────────

// ArgumentDefinition describes one logical argument: the set of equivalent
// flag spellings (e.g. ["-f", "/f", "--file"]) and how many values it consumes.
// Used by the Shorthands modifier to expand or contract flag abbreviations, and
// by the tokenizer to correctly classify value tokens.
type ArgumentDefinition struct {
	Flags      []string `json:"flags"`
	ValueCount int      `json:"valueCount"`
}

// ─── Base modifier config ─────────────────────────────────────────────────────

// BaseModifierConfig holds the two fields that every modifier config must have.
// Embed this in modifier-specific config structs, then unmarshal from the
// json.RawMessage stored in ProfileParameters.Modifiers.
//
// Example:
//
//	type RandomCaseConfig struct {
//	    models.BaseModifierConfig
//	}
//	var cfg RandomCaseConfig
//	json.Unmarshal(rawMsg, &cfg)
type BaseModifierConfig struct {
	// AppliesTo is the set of TokenTypes this modifier should act on.
	// Values match the TokenType constants: "command", "argument", "value", "path", "url".
	AppliesTo []string `json:"AppliesTo"`

	// Probability is a string in [0.0, 1.0] controlling how often the modifier
	// fires on each eligible token. Parse with strconv.ParseFloat.
	Probability string `json:"Probability"`
}
