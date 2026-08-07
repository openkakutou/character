---
status: todo
---
# Decode v2 RLE5 Pixel Format

## Description
Extend `DecodeV2Sprite` (`sff/v2_decoder.go`) with a `V2FormatRLE5` case, producing a `V2Image` of raw palette-index bytes (`BytesPerPixel: 1`), matching the shape items 024/025 already established for RLE8/LZ5. Unlike those two, the reference project (`ikemen-launcher/sff-extractor`) has **no working RLE5 decoder to port** — `decodeSpriteBuffer.mjs` explicitly throws `TODO RLE5` for this format. This item has to derive the encoding directly from the MUGEN/Ikemen `.sff` v2 format documentation instead of porting known-good reference behavior.

## Acceptance Criteria
- [ ] `DecodeV2Sprite(V2FormatRLE5, ...)` decodes RLE5-encoded data into the correct row-major index buffer, verified against at least one real RLE5-encoded sprite (not only hand-built synthetic data), since no independent reference implementation exists to cross-check against
- [ ] Malformed/truncated RLE5 data returns a descriptive error instead of panicking or reading out of bounds
- [ ] `V2FormatRLE5` is removed from the "unsupported format" error path/tests it currently falls under
- [ ] The format's bit-level layout as implemented is documented in the function's doc comment, since there's no upstream source to point readers at for the "why" the way items 024/025 can

## Notes
Because there's no reference decoder to validate against (unlike RLE8/LZ5), this item carries more correctness risk than usual — per `.vibe/decisions/006-sff-v2-pixel-decode-shape-and-scope.md`'s own reasoning ("an unverified compressed format is worse than an honest error"), prefer finding at least one real character file that actually uses RLE5 to test against over shipping an unverified guess. If no such real fixture can be found/verified, consider documenting the gap and keeping the existing "unsupported format" error instead of a silently-wrong decoder.
