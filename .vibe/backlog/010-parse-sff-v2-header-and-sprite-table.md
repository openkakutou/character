---
status: in_progress
depends_on: [006]
---
# Parse .sff v2 Header and Sprite Table

## Description
Parse the Ikemen-compatible `.sff` v2 binary header and sprite index table into the `Sprite`/`SpriteGroup` model (item 006). Covers the v2-specific table layout: palette bank references, per-sprite flags, and links to shared sprite data.

## Acceptance Criteria
- [ ] Parsing a valid v2 `.sff` header yields correct sprite count, palette bank count, and per-sprite flags
- [ ] The v2 index table correctly resolves palette bank and shared-data references
- [ ] Malformed or truncated header data returns a descriptive error rather than panicking

## Notes
v2 is needed for Ikemen GO compatibility, per CLAUDE.md's stack requirements. Independent of the v1 items (007–009).
