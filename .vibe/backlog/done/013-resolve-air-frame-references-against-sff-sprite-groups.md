---
status: done
depends_on: [002, 006]
---
# Resolve .air Frame References Against .sff Sprite Groups

## Description
Implement cross-format resolution linking each `Frame`'s (group, image) reference in a parsed `Animation` to the corresponding `Sprite` in a loaded `SpriteGroup` collection, per CLAUDE.md's constraint that `.air` and `.sff` are interdependent formats. A frame referencing a missing sprite must fail explicitly rather than silently rendering blank.

## Acceptance Criteria
- [ ] Given a matching `Animation` and `SpriteGroup` set, every `Frame` reference resolves to the correct `Sprite`
- [ ] A `Frame` referencing a non-existent (group, image) pair returns a descriptive error rather than being silently skipped
- [ ] Resolution works against sprites loaded from either sff v1 or v2, with no version-specific branching required by the caller

## Notes
First true cross-package integration point; depends on both the `air` parser (002) and the `sff` data model (006).
