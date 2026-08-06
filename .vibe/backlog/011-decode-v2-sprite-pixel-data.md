---
status: in_progress
depends_on: [010]
---
# Decode v2 Sprite Pixel Data

## Description
Implement pixel decoding for sff v2 sprites, handling both raw literal and PNG-encoded formats as indicated by the per-sprite flag from item 010. Testable in isolation against known v2 fixtures of each encoding.

## Acceptance Criteria
- [ ] A raw-encoded v2 sprite fixture decodes to the expected pixel buffer
- [ ] A PNG-encoded v2 sprite fixture decodes to the expected pixel buffer
- [ ] An unrecognized encoding flag returns a descriptive error rather than panicking

## Notes
Independent decoding unit from the table parsing in 010, but requires it to locate sprite data and determine the encoding flag.
