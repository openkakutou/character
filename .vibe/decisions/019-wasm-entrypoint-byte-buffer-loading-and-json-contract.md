---
date: 2026-08-08
status: accepted
---
# WASM entrypoint loads characters from in-memory byte buffers, returns the whole Character graph as one JSON blob

**Context:** Item 033 adds a WASM entrypoint so a browser consumer (`character-viewer-web`) can load a MUGEN/Ikemen character without a Go toolchain. The existing `character.Load(path string)` reads `.def`/`.air`/`.sff`/`.cns` from the filesystem by path, resolving referenced files relative to the `.def` file's own directory — a browser has no such filesystem, so this shape doesn't fit a JS caller, which will already have fetched/selected the four files' bytes itself.

**Decision:**
- Add `character.LoadBytes(defBytes, airBytes, sffBytes, cnsBytes []byte) (*Character, error)` in the root package: a plain, fully unit-testable Go function (no `syscall/js` involved) that parses each buffer directly via the existing `def.Parse`/`air.Parse`/`sff.Load`/`cns.Parse`, with no file I/O or path resolution. `cmd/wasm/main.go` (build-tag gated `js && wasm`) is thin glue: it calls `LoadBytes`, marshals the result via `encoding/json`, and exposes it to JS as a single global function returning the JSON string (or a JS-visible error).
- `LoadBytes` returns `(nil, error)` on any stage failure, with each error wrapped to name which stage failed — mirroring `Load`'s existing per-stage message pattern in `load.go` (`"character: parsing character definition bytes: %w"`, `"...animation bytes..."`, `"...sprite bytes..."`, `"...combat logic bytes..."`, swapping `file %q` for `bytes` since there is no path) rather than inventing a new prefix convention. This is the only error text that ever reaches a JS `.message`.
- `LoadBytes` normalizes every slice/map on the returned `Character` graph to non-nil (empty `[]`/`{}` rather than `null`) before returning, so the JSON a JS caller sees is always iterable without a null-check — a guarantee specific to this JSON/JS-facing entrypoint, not retrofitted onto `Parse`/`Load`'s existing Go-idiomatic nil slices.
- Explicit `json:"..."` tags are added to every exported field on `Character`/`air.Animation`/`air.Frame`/`air.ClsnBox`/`sff.SpriteGroup`/`sff.Sprite`/`cns.StateDef`/`cns.Controller` to pin the wire contract deliberately rather than let it ride on Go's default field-name reflection.
- The whole `Character` graph is returned as one JSON blob from a single call — no per-sub-resource (animations-only, sprites-only, ...) entrypoints yet. Sprite pixel/palette color data is out of scope for this item (`sff.Sprite` metadata only — dimensions, axis, palette index).

**Reason:** Keeps the browser-facing surface minimal and reuses every existing parser unchanged; only a new assembly function and thin glue are added. Byte buffers (not paths) match what a browser environment actually has. Whole-graph JSON keeps the JS contract simple for a first entrypoint — splitting into per-resource calls can be added later if a real usage pattern (e.g. lazy-loading large sprite sheets) demands it, without this decision blocking that.

**Rejected alternatives:**
- Extending `character.Load` itself to accept an `io.Reader`/byte-buffer variant in place of a path — rejected because `Load`'s contract (resolve referenced files relative to the `.def` file's directory) has no meaning without a filesystem; a separate function is clearer than overloading one with two incompatible input shapes.
- Per-sub-resource WASM calls (`loadAnimations`, `loadSprites`, ...) from the start — rejected as premature: no consumer usage pattern yet justifies the added surface, and a single call is simpler to document and test first.
- Returning resolved/decoded sprite pixel colors (RGBA) alongside metadata in this same call — rejected as out of scope: it requires keeping parsed file state alive across calls (or re-parsing per sprite) and a separate design conversation about how pixel data crosses the JS boundary (likely per-sprite, on demand) — left for a future item once `character-viewer-web` actually needs to render sprites.
