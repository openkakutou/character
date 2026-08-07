---
status: in_progress
---
# Decode v2 LZ5 Pixel Format

## Description
Extend `DecodeV2Sprite` (`sff/v2_decoder.go`) with a `V2FormatLZ5` case, producing a `V2Image` of raw palette-index bytes (`BytesPerPixel: 1`). Port the bit-unpacking loop from `decodeLZ5.mjs` (dictionary back-references plus literal runs, driven by a control-bit byte) down to index bytes only, dropping its fused palette-lookup step, per the same split as item 024.

## Acceptance Criteria
- [ ] `DecodeV2Sprite(V2FormatLZ5, ...)` decodes LZ5-encoded data into the correct row-major index buffer for a synthetic fixture built in-test
- [ ] Malformed/truncated LZ5 data returns a descriptive error instead of panicking or reading out of bounds
- [ ] `V2FormatLZ5` is removed from the "unsupported format" error path/tests it currently falls under

## Notes
Reference: `decodeLZ5.mjs` in `ikemen-launcher/sff-extractor`. Cross-validated against `Lz5Decode` in the real engine, `ikemen-engine/Ikemen-GO` (`src/image.go`) — same algorithm, already in Go, useful as a second source if anything is ambiguous when porting.
