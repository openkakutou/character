---
status: todo
depends_on: [007, 008]
---
# Serialize Sprites to .sff v1 Format

## Description
Implement the write path for sff v1: serialize `Sprite`/`SpriteGroup` data (including PCX-encoded pixel data) back into a valid `.sff` v1 binary file, and prove round-trip fidelity against real fixtures.

## Acceptance Criteria
- [ ] Serializing a parsed v1 sprite set produces a file that re-parses (via items 007/008) into an equivalent structure
- [ ] A round-trip test on a realistic multi-sprite v1 fixture passes with no pixel data loss
- [ ] Shared-palette linkage is preserved through the round-trip

## Notes
Completes the sff v1 read+write cycle before moving to v2.
