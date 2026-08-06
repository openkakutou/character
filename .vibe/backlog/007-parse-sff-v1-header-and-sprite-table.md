---
status: todo
depends_on: [006]
---
# Parse .sff v1 Header and Sprite Table

## Description
Parse the MUGEN `.sff` v1 binary header and sprite index table into the `Sprite`/`SpriteGroup` model (item 006). Covers the (group, image) → file-offset table and shared-palette linkage used by v1 files, without yet decoding the actual pixel data (item 008).

## Acceptance Criteria
- [ ] Parsing a valid v1 `.sff` header yields the correct sprite count, version, and palette-sharing metadata
- [ ] The (group, image) index table correctly resolves to per-sprite offsets
- [ ] Malformed or truncated header data returns a descriptive error rather than panicking

## Notes
v1 is the historical MUGEN format most existing community characters use, per the user's priority decision to sequence v1 before v2.
