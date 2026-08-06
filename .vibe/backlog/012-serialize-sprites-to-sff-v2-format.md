---
status: todo
depends_on: [010, 011]
---
# Serialize Sprites to .sff v2 Format

## Description
Implement the write path for sff v2: serialize `Sprite`/`SpriteGroup` data back into a valid `.sff` v2 binary file, and prove round-trip fidelity against real fixtures.

## Acceptance Criteria
- [ ] Serializing a parsed v2 sprite set produces a file that re-parses (via items 010/011) into an equivalent structure
- [ ] A round-trip test on a realistic multi-sprite v2 fixture (mixing raw and PNG-encoded sprites) passes with no pixel data loss
- [ ] Palette bank references are preserved through the round-trip

## Notes
Completes the sff v2 read+write cycle.
