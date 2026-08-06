---
status: todo
depends_on: [014, 016]
---
# Wire .def as Entry Point in Root Package

## Description
Extend the root package so a `.def` file can be loaded as the entry point: parse it into `CharacterInfo` (item 016), resolve the referenced `.sff`/`.air` file paths, and assemble a fully populated `Character` struct (building on item 014's air+sff assembly).

## Acceptance Criteria
- [ ] Loading a `.def` fixture with valid sprite/animation references produces a fully populated `Character` with resolved animations and sprites
- [ ] A `.def` referencing a missing `.sff`/`.air` file returns a descriptive error rather than panicking
- [ ] The public entry-point function accepts a `.def` path and returns a `Character`, matching the read-path API design goal in CLAUDE.md

## Notes
First end-to-end "load a real character" capability; depends on both the air+sff integration (014) and the `.def` parser (016). Does not require the `.def` write path (017) since loading is read-only.
