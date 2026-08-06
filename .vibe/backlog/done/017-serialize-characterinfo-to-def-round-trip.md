---
status: done
depends_on: [016]
---
# Serialize CharacterInfo to .def With Round-Trip Suite

## Description
Implement the write path for the `def` package: serialize `CharacterInfo` back into valid `.def` text, preserving section ordering and comments, and prove round-trip fidelity against a real fixture — required by CLAUDE.md's testing conventions for write-path code.

## Acceptance Criteria
- [ ] Serializing a parsed `CharacterInfo` produces valid `.def` text that re-parses (via item 016) into an equivalent structure
- [ ] A realistic `.def` fixture round-trips through parse → serialize with comments and section ordering preserved
- [ ] Round-trip equivalence is asserted in a test, not just eyeballed

## Notes
Completes the `def` read+write cycle before wiring it as the library's entry point.
