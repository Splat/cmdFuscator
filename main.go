// Package cmdfuscator provides command-line argument obfuscation primitives
// compatible with ArgFuscator.net profile files (format version 2.0).
//
// Public packages:
//
//   - [cmdFuscator/models]  – Token, Profile, and related data types
//   - [cmdFuscator/loader]  – Parse ArgFuscator JSON profile files
//   - [cmdFuscator/engine]  – Orchestrate the obfuscation pipeline
//   - [cmdFuscator/engine/modifiers] – Modifier interface and registry
//
// The TUI application lives in cmd/cmdfuscator/ and is not part of the
// public module API.
//
// Quick start:
//
//	go run ./cmd/cmdfuscator
package cmdfuscator
