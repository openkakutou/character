---
status: in_progress
---
# Define Sprite and SpriteGroup Data Model

## Description
Define the pure-data model for the `sff` package: `Sprite` (group, image index, width, height, axis offset, palette reference) and `SpriteGroup` (a collection of `Sprite`s sharing a group index). No binary parsing yet — establishes the read-path vocabulary that both the sff v1 (item 007) and v2 (item 010) parsers will populate.

## Acceptance Criteria
- [ ] `Sprite` and `SpriteGroup` structs are exported with documented fields covering group/image indexing, dimensions, offset, and palette reference
- [ ] Zero-value structs compile and behave predictably, covered by a zero-value test
- [ ] The model is version-agnostic (no v1/v2-specific fields leaking into the shared read-path type)

## Notes
No dependency — can proceed independently of the `air` package, though both feed into item 013's cross-format integration.
