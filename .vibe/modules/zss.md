# Module: zss

**Role:** Ikemen GO Lua-like state script (`.zss`) files — an alternative to `.cns` combat logic (a character uses either format, never both). Splits `.zss` text into an ordered list of top-level `Statedef`/`Function` blocks, parsing each header into typed fields while keeping the block's own Lua-like script body as raw, unevaluated text — this package parses `.zss` structure only, it does not execute the scripting language (that is `engine`'s future responsibility, per roadmap decision 012).

**Files:** `zss/script.go`, `zss/parser.go`, `zss/serializer.go`, `zss/document.go`

**Exports:**
- `Script{Preamble, Blocks}`, `Block{Kind, Number, HeaderParams, Name, Params, Ret, Body}`, `BlockKind` (`BlockKindStatedef`/`BlockKindFunction`) — the pure-data model
- `Parse(r io.Reader) (Script, error)` — reads `.zss` text into a `Script`
- `Serialize(w io.Writer, script Script) error` — writes a `Script` back out to valid `.zss` text (not byte-exact)
- `Document{Script}`, `ParseDocument(r io.Reader) (*Document, error)`, `(*Document) Serialize(w io.Writer) error` — byte-exact round trip for unmodified content

**Depends on:** none (no dependency on `cns`, unlike `cmd` — a `.zss` block header shares no syntax with `.cns`'s)

**Not yet wired into:** the root `character` package's `Character` struct (same status as `cmd`)
