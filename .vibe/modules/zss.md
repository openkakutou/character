# Module: zss

**Role:** Ikemen GO Lua-like state script (`.zss`) files — an alternative to `.cns` combat logic. Splits `.zss` text into an ordered list of top-level `Statedef`/`Function` blocks, parsing each header into typed fields while keeping the block's own Lua-like script body as raw, unevaluated text — this package parses `.zss` structure only, it does not execute the scripting language (that is `engine`'s future responsibility, per roadmap decision 012).

**Files:** `zss/script.go`, `zss/parser.go`, `zss/serializer.go`, `zss/document.go`, `zss/edit.go`

**Exports:**
- `Script{Preamble, Blocks}`, `Block{Kind, Number, HeaderParams, Name, Params, Ret, Body}`, `BlockKind` (`BlockKindStatedef`/`BlockKindFunction`) — the pure-data model
- `Parse(r io.Reader) (Script, error)` — reads `.zss` text into a `Script`
- `Serialize(w io.Writer, script Script) error` — writes a `Script` back out to valid `.zss` text (not byte-exact; appends a trailing newline when the last block's body lacks one)
- JSON tags: `Script` always carries `blocks`, `Block` always carries `number`/`headerParams`/`params`/`ret` (no `omitempty`; unused ones are `0`/`null`); only `preamble` and `name` are omitted when empty
- `Document{Script}`, `ParseDocument(r io.Reader) (*Document, error)`, `(*Document) Serialize(w io.Writer) error` — byte-exact round trip for unmodified content
- `(*Document) SerializeEdited(w io.Writer, script Script) error` — format-preserving edit path (item 061, ADR 033), see Edits below

**Depends on:** none (no dependency on `cns`, unlike `cmd` — a `.zss` block header shares no syntax with `.cns`'s)

**Edits (item 061, `.vibe/decisions/033-format-preserving-edits-per-format-alignment.md`):** Only changed lines are rewritten (everything else, comments and line endings included, stays byte-exact; an unchanged model writes the source verbatim); the baseline is the model re-parsed from the retained source, so mutating `Document` fields in place is safe. Blocks are aligned by LCS on identity (Statedef number, Function name), unmatched runs paired positionally so a renumbered Statedef keeps its comments. A modified block whose header fields are unchanged keeps its header lines verbatim (including wrapped headers); otherwise the header is regenerated as `Serialize` writes it. A changed `Body` (and `Preamble`) is diffed line by line (LCS), so a one-line edit rewrites one line. `#` comment/blank lines above a header belong to that block and are deleted with it; a surplus block is inserted after the nearest surviving predecessor's body (after the preamble when none). Dominant line ending for generated lines.

**Wired into:** the root `character` package as `Character.Scripts` (item 060, `.vibe/decisions/032-zss-scripts-wired-into-character-ikemen-file-selection.md`), loaded from `.def` `st`/`stN`/`stcommon` entries.
