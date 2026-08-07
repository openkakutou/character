---
status: done
---
# Decode v2 RLE8 Pixel Format

## Description
Extend `DecodeV2Sprite` (`sff/v2_decoder.go`) with a `V2FormatRLE8` case, producing a `V2Image` of raw palette-index bytes (`BytesPerPixel: 1`), mirroring `decodeV2Raw`'s output shape rather than the reference project's fused palette-application approach. Port the run-length algorithm from `decodeRLE8.mjs` (run marker `0b01xxxxxx` in the low 6 bits of a byte, literal bytes otherwise) but strip its palette lookup, keeping this package's existing "decode to indices, resolve color separately" split (`.vibe/decisions/006-sff-v2-pixel-decode-shape-and-scope.md`).

## Acceptance Criteria
- [ ] `DecodeV2Sprite(V2FormatRLE8, ...)` decodes RLE8-encoded data into the correct row-major index buffer for a synthetic fixture built in-test
- [ ] Malformed/truncated RLE8 data (too short, run overruns image bounds) returns a descriptive error instead of panicking
- [ ] `V2FormatRLE8` is removed from the "unsupported format" error path/tests it currently falls under

## Notes
Reference: `decodeRLE8.mjs` in `ikemen-launcher/sff-extractor`. Cross-validated against `Rle8Decode` in the real engine, `ikemen-engine/Ikemen-GO` (`src/image.go`) — same algorithm, already in Go, useful as a second source if anything is ambiguous when porting. `RLE5` is handled separately (item 030): `sff-extractor` has no implementation for it, but `Ikemen-GO` does (`Rle5Decode`, same file).
