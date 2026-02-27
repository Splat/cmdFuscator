# cmdFuscator

A terminal UI (TUI) port of [ArgFuscator.net](https://argfuscator.net) - a tool for
obfuscating command-line arguments to evade signature-based detection.

Built in Go as an educational security research project.

> **Source repo being ported:** https://github.com/wietze/ArgFuscator.net

---

## Project Overview

cmdFuscator reads ArgFuscator-compatible JSON profile files to learn which obfuscation techniques apply to a given executable, then presents an interactive terminal interface for applying those techniques to a command you type.

A de-obfuscator is planned as a future addition. Both the obfuscator and de-obfuscator are designed to be importable as Go module packages as well as usable via the CLI.

## Architecture

The module root exposes public packages (`models`, `loader`, `engine`) that can be
imported by external tools. The TUI lives entirely under `cmd/cmdfuscator/` and is
not part of the public API. A future de-obfuscator will live alongside `engine/` as a peer package at the module root.

```
cmdFuscator/
├── main.go                             # Module doc file (package cmdfuscator)
├── go.mod / go.sum
├── cmd/
│   └── cmdfuscator/
│       ├── main.go                     # TUI entry point
│       └── tui/
│           ├── app.go                  # Bubbletea model (View / Update / Init)
│           ├── styles.go               # Lipgloss style definitions
│           └── keys.go                 # Key binding definitions
├── data/
│   ├── data.go                         # go:embed declaration (exports ModelFS)
│   └── models/
│       ├── bash.json
│       ├── certutil.json
│       └── powershell.json             # add more from ArgFuscator repo here
├── models/
│   └── models.go                       # Token, Profile, ProfileFile, etc.
├── loader/
│   └── loader.go                       # LoadFS, IndexByName, GroupByPlatform
└── engine/
    ├── engine.go                       # Obfuscate(); Tokenize + Render stubs
    └── modifiers/
        ├── modifier.go                 # Modifier interface + registry
        ├── all/
        │   └── all.go                  # Blank imports to register all modifiers
        ├── charinsert/
        │   └── char_insertion.go       # STUB – TODO
        ├── filepath/
        │   └── file_path.go            # STUB – TODO
        ├── optionchar/
        │   └── option_char_sub.go      # STUB – TODO
        ├── quoteinsert/
        │   └── quote_insertion.go      # STUB – TODO
        ├── randomcase/
        │   └── random_case.go          # STUB – TODO
        ├── regex/
        │   └── regex.go                # STUB – TODO
        ├── reorderargs/
        │   └── reorder_args.go         # STUB – TODO
        ├── sed/
        │   └── sed.go                  # STUB – TODO
        ├── shorthands/
        │   └── shorthands.go           # STUB – TODO
        └── urltransform/
            └── url_transformer.go      # STUB – TODO
```

### Package Import Paths

| Package                              | Import path                       |
|--------------------------------------|-----------------------------------|
| Module doc / root                    | `cmdFuscator`                     |
| Embedded profile data                | `cmdFuscator/data`                |
| Data types                           | `cmdFuscator/models`              |
| Profile loader                       | `cmdFuscator/loader`              |
| Obfuscation engine                   | `cmdFuscator/engine`              |
| Modifier interface + registry        | `cmdFuscator/engine/modifiers`    |
| TUI (CLI only, not a library export) | `cmdFuscator/cmd/cmdfuscator/tui` |

## JSON Model Format

Each file in `data/models/` follows this schema:

```json
{
  "versions": { "argfuscator": "2.0", "format": "2.0" },
  "profiles": [{
    "executableVersion": "...",
    "platform": "windows|linux|macos",
    "operatingSystem": "Windows|Ubuntu|macOS",
    "operatingSystemVersion": "...",
    "alias": ["alt-name"],
    "parameters": {
      "command": [
        {"command": "certutil.exe"},
        {"argument": "-urlcache"},
        {"url": "https://example.com"},
        {"path": "output.txt"}
      ],
      "arguments": [],
      "modifiers": {
        "RandomCase":             { "AppliesTo": ["argument","value"], "Probability": "0.5" },
        "QuoteInsertion":         { "AppliesTo": ["path","url"],       "Probability": "0.5" },
        "OptionCharSubstitution": { "AppliesTo": ["argument"],         "Probability": "0.5",
                                    "OutputOptionChars": ["/","-","–"] },
        "Sed":                    { "AppliesTo": ["argument"],         "Probability": "0.5",
                                    "SedStatements": "s/a/ᵃ/i\ns/e/ᵉ/i" },
        "FilePathTransformer":    { "AppliesTo": ["path"],             "Probability": "0.5",
                                    "PathTraversal": true, "SubstituteSlashes": true },
        "CharacterInsertion":     { "AppliesTo": ["argument"],         "Probability": "0.5",
                                    "Characters": ["…"], "Offset": "2" }
      }
    }
  }]
}
```

### Token Types (`AppliesTo` values)

| Token Type | Meaning                                            |
|------------|----------------------------------------------------|
| `command`  | The executable name (e.g. `certutil`, `curl`, etc) |
| `argument` | A flag/switch (e.g. `-urlcache`, `--log-level`)    |
| `value`    | A value for a preceding argument                   |
| `path`     | A file-system path argument                        |
| `url`      | A URL argument                                     |

## Implementing Modifiers

Each modifier stub in `engine/modifiers/<name>/` has:

- A struct that implements the `Modifier` interface
- `Name() string` – must match the JSON key exactly (e.g. `"RandomCase"`)
- `Description() string` – shown in the TUI options panel
- `Apply(tokens []models.Token, cfg json.RawMessage) ([]models.Token, error)`
  – the function you implement; `cfg` is the raw modifier config from the JSON profile

The engine calls `Apply()` on each enabled modifier in sequence. Stubs return
`modifiers.ErrNotImplemented`; the engine skips them gracefully and reports them
in the TUI status bar.

New modifiers self-register via `init()`:

```go
func init() { modifiers.Register(&MyModifier{}) }
```

Then add a blank import to `engine/modifiers/all/all.go`.

### Good Go Practices to Apply

- Use `errors.New` / `fmt.Errorf("...: %w", err)` for error wrapping
- Prefer value receivers for small structs, pointer receivers when mutating
- Use `strings.Builder` for efficient string construction
- Use `math/rand` with a seeded source for randomness
- Write table-driven tests in `_test.go` files alongside each modifier

## Implementation Guide

This section is a curriculum for working through the stubs in a logical order. Each phase builds on the last and introduces progressively more interesting Go concepts.

### What to Implement

| File                             | What to implement                                       |
|----------------------------------|---------------------------------------------------------|
| `engine/engine.go`               | `Tokenize()` — parse command string into typed tokens   |
| `engine/engine.go`               | `Render()` — join tokens back into a command string     |
| `engine/modifiers/randomcase/`   | Probabilistic per-character case flip (**implemented**) |
| `engine/modifiers/quoteinsert/`  | Insert empty `""` or `''` inside tokens                 |
| `engine/modifiers/optionchar/`   | Replace `-` with `–`, `/`, `—`, etc.                    |
| `engine/modifiers/sed/`          | Parse `s/a/ᵃ/i` rules and apply per-char substitution   |
| `engine/modifiers/filepath/`     | Path traversal, slash substitution, extra separators    |
| `engine/modifiers/charinsert/`   | Insert invisible Unicode codepoints at a fixed offset  (**implemented**) |
| `engine/modifiers/shorthands/`   | Abbreviate flags to shortest unambiguous prefix         |
| `engine/modifiers/urltransform/` | Hex/octal IP encoding, URL path traversal               |
| `engine/modifiers/reorderargs/`  | Shuffle flag–value pairs while keeping them grouped     |
| `engine/modifiers/regex/`        | Regex find-and-replace substitutions                    |

Each stub has detailed guidance comments. The TUI gracefully labels unimplemented
modifiers as "not implemented" in the status bar without crashing.

---

### Phase 1 — Foundation: Tokenize and Render

**Files:** `engine/engine.go`

Start here. Everything else depends on the token representation being correct.
No randomness, no config parsing — just pure string → struct → string.

**`Tokenize(command string, profile models.Profile) ([]models.Token, error)`**

1. Split the input on whitespace (but respect quoted strings — a value like `"hello world"` is one token).
2. The first token is always `TokenTypeCommand`.
3. Walk `profile.Parameters.Arguments` to build a map of known flags → `ValueCount`.
4. For each remaining token:
   - If it matches a known flag → `TokenTypeArgument`; consume the next N tokens as `TokenTypeValue`.
   - If it starts with `http://` or `https://` → `TokenTypeURL`.
   - If it contains `/` or `\` (and isn't a flag) → `TokenTypePath`.
   - Otherwise → `TokenTypeValue`.

**`Render(tokens []models.Token) string`**

Join tokens with spaces. Re-quote any value that contains a space.
The invariant `Render(Tokenize(cmd)) == cmd` should hold for unmodified input.

**Go concepts introduced:** `strings.Fields`, `strings.Builder`, slice operations, map lookups, `strconv`.

**Tests to write (`engine/engine_test.go`):**

```go
// Table-driven round-trip test
var cases = []struct{
    input    string
    wantTokens []models.Token
}{
    {"certutil.exe -urlcache -f https://x.com out.bin", [...]},
    {"bash -c id", [...]},
}
```

---

### Phase 2 — Stateless String Transformations

These modifiers touch individual characters or token boundaries. They are the easiest
to reason about because the output is deterministic once you fix the random seed.

#### 2a. `RandomCase`

- Iterate over the rune slice of each eligible token value.
- For each rune, roll `rand.Float64()`. If `< probability`, flip case with `unicode.ToUpper` / `unicode.ToLower`.
- Rebuild the string with `strings.Builder`.

**Go concepts introduced:** `[]rune` vs `[]byte`, `unicode` package, `math/rand`.

#### 2b. `QuoteInsertion`

- Pick a random insertion position between index 1 and `len(runes)-1`.
- Insert `""` or `''` (chosen randomly) at that position.

**Go concepts introduced:** Slice insertion (`append(s[:i], append([]T{x}, s[i:]...)...)`).

#### 2c. `OptionCharSubstitution`

- Check whether `runes[0]` is `-` or `/`.
- If so, pick a random entry from `cfg.OutputOptionChars` and replace the first rune.

**Go concepts introduced:** `json.Unmarshal` into a typed config struct, multi-byte UTF-8 rune indexing.

**Tests to write (per-modifier `_test.go`):**

```go
// Seed rand so output is deterministic, then assert exact output.
// Also test that tokens NOT in AppliesTo are never modified.
// Also test that Probability=0.0 always returns input unchanged.
// Also test that Probability=1.0 always transforms every eligible token.
```

---

### Phase 3 — Rule-Based Transformations

#### 3a. `Sed`

Parse `SedStatements` (newline-delimited `s/<from>/<to>/i` rules) into a
`map[rune]string` substitution table, then apply it per-character with probability.

- Split on `\n` to get individual rules.
- For each rule, the character after `s` is the delimiter. Split on it: `[from, to]`.
- The `/i` flag means both `unicode.ToUpper(from)` and `unicode.ToLower(from)` map to `to`.
- Apply the table: for each eligible rune, if it exists in the map and probability fires, replace it.

**Go concepts introduced:** String parsing without `regexp`, `rune` → `string` maps.

#### 3b. `Regex`

Compile each rule's `Pattern` with `regexp.Compile`, then call `re.ReplaceAllString` on each eligible token value.

**Go concepts introduced:** `regexp` package, error handling for user-supplied patterns.

---

### Phase 4 — Structural Path and URL Transformations

These modifiers require parsing structured values (file paths, URLs) rather than
treating tokens as opaque strings.

#### 4a. `FilePathTransformer`

Use `strings.Split` on `/` and `\` to get path components, then:
- **SubstituteSlashes:** randomly swap `/` for `\` and vice versa.
- **PathTraversal:** insert `./` or `.\` between two random adjacent components.
- **ExtraSlashes:** double one random separator.

**Go concepts introduced:** `path/filepath`, platform-aware separator handling.

#### 4b. `UrlTransformer`

Parse with `net/url.Parse`, then inspect `u.Hostname()`:
- If `net.ParseIP(host)` succeeds, encode the IP in one of three alternate forms:
  - **Integer:** pack four octets into `uint32` with `binary.BigEndian`, format as `%d`.
  - **Hex:** same `uint32`, format as `0x%08x`.
  - **Octal:** format each octet as `0%o` and rejoin with `.`.
- Reconstruct the URL string with the modified host.

**Go concepts introduced:** `net/url`, `net.IP`, `encoding/binary`, format verbs.

---

### Phase 5 — Argument-Aware Transformations

These modifiers need to understand the relationship between flags and their values,
making them the most structurally complex.

#### 5a. `Shorthands`

1. Build an index of all known flags from `profile.Parameters.Arguments`.
2. For each argument token, strip its leading option char and find the matching flag entry.
3. Find the shortest prefix of that flag's canonical form that is unambiguous (no other known flag shares it).
4. Replace the token's value with `<option-char> + shortest-prefix`.

**Go concepts introduced:** Prefix matching, data-driven lookup tables, handling ambiguity.

#### 5b. `ReorderArgs`

1. Separate the command token (index 0) from the rest.
2. Group remaining tokens into `(flag, value...)` pairs using the `ArgumentDefinitions` ValueCount.
3. Shuffle the pairs with `rand.Shuffle`.
4. Flatten: `[command] + [group1 tokens...] + [group2 tokens...] + ...`.

**Go concepts introduced:** `rand.Shuffle`, grouping slices by a data-driven rule.

---

### Testing Strategy

#### Unit tests (one `_test.go` per modifier package)

Write table-driven tests. Every test case should specify:

| Field        | Purpose                                                          |
|--------------|------------------------------------------------------------------|
| `name`       | Description shown on failure                                     |
| `input`      | `[]models.Token` before the modifier runs                        |
| `cfg`        | Raw JSON config (use `json.RawMessage(...)` literals)            |
| `want`       | Expected `[]models.Token` after the modifier runs                |
| `wantErr`    | Whether an error is expected                                     |

Seed `math/rand` in test setup so randomized modifiers produce deterministic output:

```go
// In TestMain or individual tests:
rand.Seed(42)
```

#### Property tests (for probabilistic modifiers)

Even without a fixed seed you can assert structural invariants:

- The token count never changes (modifiers only mutate values, not add/remove tokens).
- `TokenTypeCommand` tokens are never modified unless `"command"` is in `AppliesTo`.
- The original string is recoverable when `Probability = "0.0"`.
- With `Probability = "1.0"`, every eligible token is different from the input (for case-flipping modifiers).

#### Integration tests (`engine/engine_test.go`)

Test the full pipeline end to end:

```go
result, err := eng.Obfuscate("certutil.exe -urlcache -f https://x.com out.bin", profile, enabled)
// assert no error, output is non-empty, output differs from input
```

#### Profile parsing tests (`loader/loader_test.go`)

Load every file in `data/models/` and assert no parse errors. This catches JSON
schema drift early:

```go
profiles, err := loader.LoadFS(data.ModelFS)
assert.NoError(t, err)
assert.NotEmpty(t, profiles)
```

---

## Adding More Profiles

Download JSON files from:
https://github.com/wietze/ArgFuscator.net/tree/main/models

Place them in `data/models/`. They are embedded at compile time via `go:embed`
in `data/data.go`.

## Running the Project

```bash
go mod tidy                       # fetch dependencies
go run ./cmd/cmdfuscator          # launch TUI
# or build a binary
go build -o cmdfuscator ./cmd/cmdfuscator && ./cmdfuscator
```

## Dependencies

| Package                              | Role                           |
|--------------------------------------|--------------------------------|
| `github.com/charmbracelet/bubbletea` | TUI event loop                 |
| `github.com/charmbracelet/lipgloss`  | Terminal styling and layout    |
| `github.com/charmbracelet/bubbles`   | textinput and viewport widgets |

## TUI Key Bindings

| Key           | Action                         |
|---------------|--------------------------------|
| `Tab`         | Cycle focus between panels     |
| `Up` / `Down` | Navigate list / options        |
| `Space`       | Toggle modifier on/off         |
| `Enter`       | Apply obfuscation              |
| `c`           | Copy output to clipboard       |
| `r`           | Reset / clear output           |
| `/`           | Focus search bar in sidebar    |
| `Esc`         | Cancel search                  |
| `q` / `^C`    | Quit                           |

## License

This project is intended for educational and authorized security research purposes only.
The obfuscation profiles are derived from the ArgFuscator.net project (GPL-3.0).
