---
date: 2026-08-10
status: accepted
---
# `zss` package structure: two block kinds, header fields parsed, script bodies kept fully raw

**Context:** Item 037 adds a `zss` package parsing Ikemen GO's `.zss` Lua-like state script format. A corpus scan of a real character (Melty Blood Type Lumina's "Akiha_TL", 7 `.zss` files) shows top-level content is only ever one of two block kinds — `[Statedef N; key: value; ...]` and `[Function Name(params) ret]` — never a `.cns`-style separate `[State N]` controller header, since controllers are invoked as function-call-like statements (e.g. `hitDef{...}`) directly inside a block's Lua-like body.

**Decision:** `zss.Parse` splits `.zss` text into an ordered list of top-level `Block`s at each `[Statedef ...]`/`[Function ...]` header line, parses the header itself into typed/structured fields (Statedef's number and semicolon-separated key:value header parameters; Function's name, parenthesized parameter list, and declared return variable(s)), and keeps everything else — the entire Lua-like script body between one header and the next — as a single opaque raw-text field. No attempt is made to parse the scripting language's own grammar (if/else, braces, function calls, `let`/`:=` assignment): this repo parses `.zss` structure only, per roadmap decision 012 — actually interpreting the language is `engine`'s job. Content preceding the first block header (typically a file banner/comment block) is kept verbatim as a document-level preamble rather than discarded.

**Reason:** Mirrors `cns`'s own "unevaluated data" principle (`Controller.Triggers`/`Parameters` are raw strings) one layer up: the header fields are simple, genuinely structured `key: value` data worth exposing typed/keyed, while the body is a full scripting language whose grammar this repo has no use for and no execution engine to validate against — modeling it further here would be speculative complexity with no consumer.

**Rejected alternatives:**
- *Parse the Lua-like body into statements/expressions*: rejected — `character` has no expression/execution engine (that's `engine`'s responsibility, not yet built even for the simpler `.cns` triggers), so a body parser would have no way to validate itself and no current consumer needing structured access below the block level.
- *Keep the whole block (header + body) as one opaque raw string, no structured header fields*: rejected — the header's block kind and identifying number/name (needed to index/look up a given state or function) are cheap, genuinely structured data a caller (e.g. `character-editor`, `engine`) will want without re-parsing raw text, unlike the body.
