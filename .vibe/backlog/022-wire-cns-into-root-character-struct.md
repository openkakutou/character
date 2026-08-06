---
status: todo
depends_on: [018, 021]
---
# Wire .cns Into Root Character Struct

## Description
Extend the root `Character` struct and the `.def`-driven loading path (item 018) to also parse and attach a character's `.cns` combat-logic data, completing the full read/write surface across all four MUGEN/Ikemen file formats.

## Acceptance Criteria
- [ ] Loading a `.def` fixture that references a `.cns` file produces a `Character` with populated combat-logic data
- [ ] A `.def` referencing a missing `.cns` file returns a descriptive error rather than panicking
- [ ] No write-only `cns` type leaks into the public `Character` API, consistent with the separation maintained for `air`/`sff`/`def`

## Notes
Final integration item closing out the full roadmap described in CLAUDE.md.
