---
status: done
---
# Read Ikemen GO move list files

## Description
Ikemen GO characters can declare a move list in `.def` `[Files]` with `movelist` (and `movelist0`, `movelist1`… per language), each pointing to a text file the pause menu shows (Ikemen GO `src/char.go:3425-3452`, `loadMovelists`, commit `54d4148a728153c2097a0fb899a06812b2d3e193`). `def.Parse` ignores these keys today. Parse the keys, load the referenced files in `Load`/`LoadBytes` (missing file: non-fatal, like `.cmd`), and expose them on `Character` and through the WASM `load` result, keeping the text and its glyph tokens as written. `game`'s move list screen consumes it (game#042).

🔒 MUST: Ikemen GO behavior grounded in its source (key names, index meaning, path resolution, encoding); MUGEN has no move list, so MUGEN characters keep loading unchanged.

## Acceptance Criteria
- [x] A real Ikemen GO character with a move list exposes its text, per declared key
- [x] A missing or unreadable move list file is reported on `Character` without failing the load
- [x] Round trip: saving the `.def` keeps the movelist keys byte-exact
