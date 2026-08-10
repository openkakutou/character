// Package zss defines the pure-data model for Ikemen GO Lua-like state
// script (.zss) files: Script and its ordered Blocks.
//
// This is the read-path surface — the stable vocabulary a library consumer
// (editor, engine) works with. It carries no scripting-language parsing,
// file I/O, or write-only (format-preservation) logic; per CLAUDE.md's
// read/write separation constraint, that lives elsewhere.
//
// A .zss file is Ikemen GO's Lua-like alternative to .cns: a character uses
// either .cns or .zss, never both (see roadmap decision 012). Following the
// same "unevaluated data" principle cns.Controller already applies to
// trigger conditions, a Block's script body — the Lua-like statements
// between one block header and the next — is kept as raw, unevaluated text.
// This package parses .zss structure only; actually executing the scripting
// language is engine's responsibility. See
// .vibe/decisions/027-zss-block-structure-header-parsed-body-raw.md.
package zss

// BlockKind identifies which .zss top-level block header a Block represents.
type BlockKind string

const (
	// BlockKindStatedef is a "[Statedef N; ...]" state definition block.
	BlockKindStatedef BlockKind = "Statedef"
	// BlockKindFunction is a "[Function Name(...) ...]" helper function
	// block.
	BlockKindFunction BlockKind = "Function"
)

// Block is a single top-level .zss block: a "[Statedef N; ...]" state
// definition or a "[Function Name(...) ret]" helper function, together with
// its script body kept as unevaluated raw text — see the package doc
// comment.
type Block struct {
	// Kind reports whether this is a Statedef or a Function block.
	Kind BlockKind `json:"kind"`

	// Number is the state number this block defines (the N in
	// "[Statedef N; ...]"). Unused (zero) for a Function block.
	Number int `json:"number,omitempty"`
	// HeaderParams are a Statedef header's semicolon-separated
	// "key: value" fields (e.g. "type", "physics", "movetype", "anim",
	// "ctrl"), verbatim and unevaluated, keyed by lowercase field name —
	// mirroring cns.Controller.Parameters one format up. Unused (nil) for
	// a Function block.
	HeaderParams map[string]string `json:"headerParams,omitempty"`

	// Name is this block's function name (the "[Function Name(...)"
	// part). Empty for a Statedef block.
	Name string `json:"name,omitempty"`
	// Params are a Function header's parenthesized parameter names, in
	// declaration order. Empty for a Statedef block, or a Function
	// declared with no parameters.
	Params []string `json:"params,omitempty"`
	// Ret are a Function header's declared return variable name(s) (the
	// text after the closing parenthesis, e.g. "ret" or
	// "xangl,yangl,zangl"), in declaration order. Empty for a Statedef
	// block, or a Function declaring no return value.
	Ret []string `json:"ret,omitempty"`

	// Body is this block's full script text: every line between this
	// block's header and the next block header (or end of file), kept
	// verbatim as raw, unevaluated Lua-like source. See the package doc
	// comment.
	Body string `json:"body"`
}

// Script is the top-level result of parsing a .zss file: its Blocks, in
// file order, plus any Preamble content.
type Script struct {
	// Preamble is the raw content preceding the first block header —
	// typically a file banner/comment block — preserved verbatim rather
	// than discarded. Empty when the file starts directly with a block
	// header, or contains no blocks at all.
	Preamble string `json:"preamble,omitempty"`
	// Blocks are this script's Statedef/Function blocks, in file order.
	Blocks []Block `json:"blocks,omitempty"`
}
