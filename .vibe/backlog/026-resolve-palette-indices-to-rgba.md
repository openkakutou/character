---
status: todo
---
# Resolve Palette Index Buffers To RGBA

## Description
Add a new function (e.g. `sff/palette.go`) resolving a decoded index buffer (`PCXImage`/`V2Image` with `BytesPerPixel: 1`) plus a 256-entry RGBA palette into final pixel colors, reproducing two distinct alpha rules observed in the reference project depending on the caller's decode path: forced transparency at palette index 0 (used by PCX and PNG8, per `decodePCX.mjs`/`decodePNG8.mjs`) versus the palette's own literal stored alpha with no override (used by RLE8/LZ5, per `decodeRLE8.mjs`/`decodeLZ5.mjs`). Also add the two palette-byte conversions this depends on: v1's embedded 256×RGB palette (3 bytes/color, alpha always opaque — currently unread; `DecodePCX` only reads the PCX-encoded length, not the 768-byte palette block that follows a non-shared sprite), and v2's embedded palette bank data (already RGBA in the file, resolving a zero-length bank via its `LinkedIndex`: copy the previous bank, or the first bank if `LinkedIndex` is 0).

## Acceptance Criteria
- [ ] A function resolves an index buffer + RGBA palette + alpha-rule selection into a final RGBA pixel buffer
- [ ] Palette index 0 renders `(0,0,0,0)` under the "forced transparency" rule regardless of the palette's own stored value there
- [ ] The "literal alpha" rule uses the palette's own alpha byte unmodified, including at index 0
- [ ] v1's 768-byte trailing RGB palette block is read and converted to opaque RGBA
- [ ] v2's zero-length palette bank correctly resolves to its linked bank's colors (or the first bank when `LinkedIndex` is 0)

## Notes
References: `convertPaletteRGBtoRGBA.mjs`, `extractPalettesFromSFFV2.mjs` in `ikemen-launcher/sff-extractor`. Keep this resolution step separate from `PCXImage`/`V2Image`/`Sprite` (read-path pure-data types) — an explicit opt-in helper, mirroring how `DecodePCX`/`DecodeV2Sprite` are already separate from `Load`.
