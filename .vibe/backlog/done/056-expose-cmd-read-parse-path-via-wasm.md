---
status: done
depends_on: [036, 039]
---
# Expose `.cmd` Read/Parse Path Via WASM

## Description
The WASM bridge (`cmd/wasm/main.go`) exposes `saveCmd` (item 039) but no way to parse an existing `.cmd` file's raw bytes into the structured `cmd.CommandFile` JSON shape from JS — `.cmd` isn't wired into `character.LoadBytes`/`load`, unlike `.def`/`.air`/`.cns`. A JS consumer wanting to display an existing character's commands (e.g. `character-editor`'s command editor, item 008) currently has no way to obtain that structured data at all: it can save an edited `CommandFile`, but can't first read the one already on disk.

Add a `loadCmd` (or similarly named) WASM export wrapping the existing `cmd.Parse`/`cmd` package read path, returning the same `{ commandFile, error }`-style JSON contract `saveCmd` already consumes, so a caller can round-trip: parse → edit → `saveCmd`.

## Acceptance Criteria
- [x] WASM exports a parse entrypoint accepting raw `.cmd` file bytes and returning the `cmd.CommandFile` structure as JSON (remap, defaults, commands, states)
- [x] Both MUGEN-style and Ikemen GO-style `.cmd` syntax variants parse correctly through the new entrypoint, same corpus as item 036
- [x] A malformed `.cmd` file returns a descriptive error instead of a WASM panic, never throws
- [x] Verified by a Node-based smoke test, same pattern as `cmd/wasm/smoke.mjs`

## Notes
Blocks `character-editor` backlog item 008 (Command Editor) — its "view an existing command" acceptance criterion has no WASM surface to read from without this.
