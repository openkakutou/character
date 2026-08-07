---
date: 2026-08-07
status: accepted
---
# Palette resolution API shape (backlog item 026)

**Context:** Adding a helper that resolves a decoded index buffer (`PCXImage`/`V2Image`) plus a 256-entry RGBA palette into final pixel colors, plus the v1/v2 palette-byte conversions it depends on.

**Decision:**
- New exported type `Palette [256]color.RGBA` in `sff/palette.go`, kept separate from `PCXImage`/`V2Image`/`Sprite` — an opt-in helper, not a field added to those read-path types, per the backlog item's own note and consistent with `DecodePCX`/`DecodeV2Sprite` already being separate from `Load`.
- `ResolvePixels(indices []byte, palette Palette, rule AlphaRule) []color.RGBA` takes an `AlphaRule` (`AlphaForceTransparentAtIndexZero` or `AlphaLiteral`) rather than a boolean, naming the two rules the reference project applies depending on decode path (PCX/PNG8 vs RLE8/LZ5).
- v1: `DecodeV1Palette(data []byte) (Palette, error)` decodes a raw 768-byte RGB block (opaque alpha). A separate `ResolveV1Palette(table *V1SpriteTable, r io.ReaderAt, i int) (Palette, error)` locates sprite `i`'s owning palette by walking backward to the nearest sprite (including `i`) with `SharedPalette == false`, mirroring `testdata/gen/main.go`'s `findOwnPalette`.
- v2: `DecodeV2Palette(data []byte) (Palette, error)` decodes raw already-RGBA bank color data. A separate `ResolveV2Palette(table *V2SpriteTable, r io.ReaderAt, index int) (Palette, error)` follows a zero-length bank's `LinkedIndex` chain (cycle/bounds-guarded, mirroring `resolveV1Pixels`'s linked-sprite chain), rather than the bare `decodeV2PaletteColors` in isolation.

**Reason:** Empirically verified against the real fixtures in `sff/testdata/files` (a small throwaway script comparing each v1 sprite's `Offset+Length` against the next sprite's subheader offset) that a v1 sprite's trailing 768-byte embedded palette sits *after* `Offset+Length` — outside the byte range `DecodePCX` reads — not as a suffix of the same blob `DecodePCX` is handed. This ruled out the simpler design of extracting the palette as "the last 768 bytes of the same raw slice passed to `DecodePCX`": that slice never contains it. The two-level split (raw decode vs. table/link-aware resolve) mirrors the existing `DecodePCX`/`Load` and `DecodeV2Sprite`/`Load` pattern and gives callers a function usable directly with a parsed table and a sprite index, without first having to know where in the file the owning palette bytes live.

**Rejected alternatives:**
- Extracting the v1 palette as a suffix of `DecodePCX`'s own input slice — contradicted by the empirical file-layout check above.
- Adding a `Palette [256]color.RGBA` field directly to `PCXImage`/`V2Image` — rejected per the backlog item's explicit note to keep this resolution step separate from the pure-data read-path types.
- A single `ResolvePixels(..., forceTransparentAtZero bool)` boolean flag instead of a named `AlphaRule` type — rejected for the same reason `V2Format*`/`StateType` etc. use named constants elsewhere in the codebase rather than booleans for a fixed, meaningful set of choices.
