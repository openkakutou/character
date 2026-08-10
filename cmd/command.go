// Package cmd defines the pure-data model for MUGEN/Ikemen GO input command
// (.cmd) files: CommandFile, CommandDefaults, and Command.
//
// This is the read-path surface — the stable vocabulary a library consumer
// (editor, engine) works with. It carries no INI parsing, file I/O, or
// write-only (format-preservation) logic; per CLAUDE.md's read/write
// separation constraint, that lives elsewhere so importing this data model
// alone never pulls in write-only dependencies.
//
// A .cmd file has two structurally distinct parts: button remapping/command
// definitions unique to this format ("[Remap]", "[Defaults]", "[Command]"
// sections), and an "always" state ("[Statedef -1]" plus its "[State ...]"
// controllers) that links a recognized command to a state change — byte-for-
// byte the same syntax .cns already parses. CommandFile.States is that
// second part, parsed via the cns package rather than reimplemented here;
// see .vibe/decisions/025-cmd-package-reuses-cns-for-state-triggering-block.md.
// The link itself needs no dedicated modeling: it already flows through
// cns.Controller's existing unevaluated Triggers strings (e.g.
// `command = "holdback"`), the same "read-path model can't hold everything
// yet" pattern this repo applies throughout.
package cmd

import "github.com/openkakutou/character/cns"

// CommandFile is a MUGEN/Ikemen GO input command (.cmd) file: optional
// button remapping, file-level command-recognition defaults, the input
// command definitions themselves, and the linked "always" state that reacts
// to them.
type CommandFile struct {
	// Remap maps a physical button ("a", "b", "c", "x", "y", "z", "s", ...)
	// to the button it is remapped to (.cmd "[Remap]" section). A nil or
	// empty Remap means no remapping is defined — an Ikemen GO extension not
	// present in every MUGEN 1.0/1.1 .cmd file.
	Remap map[string]string `json:"remap"`
	// Defaults are the file-level command-recognition defaults (.cmd
	// "[Defaults]" section) a Command falls back to when it doesn't set its
	// own Time/BufferTime.
	Defaults CommandDefaults `json:"defaults"`
	// Commands are the input command definitions (.cmd "[Command]"
	// sections), in file order.
	Commands []Command `json:"commands"`
	// States are the linked "always" state ("[Statedef -1]" and its
	// "[State ...]" controllers) that react to a recognized command, parsed
	// via the cns package — see the package doc comment.
	States []cns.StateDef `json:"states"`
}

// CommandDefaults is a .cmd file's "[Defaults]" section: the
// command-recognition window a Command falls back to when it doesn't set
// its own Time/BufferTime.
type CommandDefaults struct {
	// Time is the default input-recognition buffer window, in game ticks
	// (.cmd "command.time").
	Time int `json:"time"`
	// BufferTime is the default duration a completed command stays
	// recognized before being discarded, in game ticks (.cmd
	// "command.buffer.time").
	BufferTime int `json:"bufferTime"`
}

// Command is a single .cmd "[Command]" section: a named input sequence a
// player can perform.
type Command struct {
	// Name identifies this command (.cmd "name"), referenced by a linked
	// state controller's trigger (e.g. `command = "holdback"`).
	Name string `json:"name"`
	// Input is the raw, unevaluated MUGEN/Ikemen input-sequence expression
	// (.cmd "command", e.g. "~D, DF, F, a") that must be entered to trigger
	// this command. Stored verbatim rather than decomposed into individual
	// steps: this repo has no input-sequence grammar to evaluate it against,
	// mirroring cns.Controller's own unevaluated Triggers/Parameters.
	Input string `json:"input"`
	// Time is this command's own input-recognition buffer window override
	// (.cmd "command.time" scoped to this [Command] block); zero means "not
	// set", in which case CommandFile.Defaults.Time applies instead.
	Time int `json:"time"`
	// BufferTime is this command's own recognized-command duration override
	// (.cmd "command.buffer.time" scoped to this [Command] block); zero
	// means "not set", in which case CommandFile.Defaults.BufferTime applies
	// instead.
	BufferTime int `json:"bufferTime"`
}
