---
status: todo
depends_on: [007]
---
# Decode PCX-Encoded Sprite Pixel Data

## Description
Implement PCX RLE decoding for sff v1 sprite pixel data, producing a pure pixel buffer independent of the index-table logic in item 007. Testable in isolation against known PCX-encoded fixtures.

## Acceptance Criteria
- [ ] A known PCX-encoded sprite fixture decodes to the expected pixel buffer
- [ ] RLE run-length edge cases (e.g. runs at buffer boundaries) decode correctly
- [ ] Corrupted/truncated PCX data returns a descriptive error rather than panicking

## Notes
Independent decoding unit from the table parsing in 007, but requires it to locate sprite offsets within a real file.
