// Package all imports every built-in modifier so their init() functions fire
// and register themselves with the modifiers registry. Import this package
// with a blank identifier in engine.go:
//
//	import _ "cmdFuscator/engine/modifiers/all"
package all

import (
	_ "cmdFuscator/engine/modifiers/charinsert"
	_ "cmdFuscator/engine/modifiers/filepath"
	_ "cmdFuscator/engine/modifiers/optionchar"
	_ "cmdFuscator/engine/modifiers/quoteinsert"
	_ "cmdFuscator/engine/modifiers/randomcase"
	_ "cmdFuscator/engine/modifiers/regex"
	_ "cmdFuscator/engine/modifiers/reorderargs"
	_ "cmdFuscator/engine/modifiers/sed"
	_ "cmdFuscator/engine/modifiers/shorthands"
	_ "cmdFuscator/engine/modifiers/urltransform"
)
