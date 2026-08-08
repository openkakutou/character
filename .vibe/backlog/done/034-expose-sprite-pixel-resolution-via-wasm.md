---
status: done
---
# Expose Sprite Pixel Resolution Via WASM

## Description
The WASM entrypoint added in item 033 (`OpenKakutouCharacter.load`) exposes sprite *metadata* only (group, image, width, height, axis, palette index) — no decoded pixel/color data, per `.vibe/decisions/019-wasm-entrypoint-byte-buffer-loading-and-json-contract.md`, which explicitly deferred this: "left for a future item once character-viewer-web actually needs to render sprites." That consumer now needs it — its sprite browser, palette picker, and animation player backlog items (`character-viewer-web` items 005/006/007) are blocked on this.

Add a WASM-callable function resolving a specific sprite's decoded pixel buffer through its palette, reusing the existing Go-side `ResolvePixels`/`ResolveV1Palette`/`ResolveV2Palette` machinery, with an optional external override palette (mirroring the `.act`-file override those functions already accept). Design the JS-facing call shape deliberately (likely per-sprite, on demand, given a loaded character handle or the previously loaded byte buffers) rather than bloating the existing whole-graph `load` JSON blob with every sprite's full pixel data up front.

## Acceptance Criteria
- [x] A JS caller can request a specific sprite's (group, image) resolved pixel data (e.g. RGBA buffer, or PNG-encoded bytes) plus its width/height, without a Go toolchain
- [x] The call supports an optional external palette override, matching `ResolveV1Palette`/`ResolveV2Palette`'s existing `override *Palette` parameter
- [x] A request for a sprite that doesn't exist, or a malformed/corrupt pixel format, returns a clear JS-visible error instead of throwing or crashing the page
- [x] Verified by a Node.js smoke test (same pattern as `cmd/wasm/smoke.mjs`) loading the built module and requesting real sprite pixel data

## Notes
Cross-repo blocker for `character-viewer-web` — see its backlog items 005 (Sprite Browser), 006 (Palette Picker), 007 (Animation Player), all of which depend on this landing first. Design the call shape before implementing: whole `Character` graph is currently stateless per `load` call (no persisted handle across calls) — decide whether this new call re-parses from the original byte buffers each time or requires the JS side to keep some parsed-state handle alive.
