---
status: todo
depends_on: [020]
---
# Serialize .cns Preserving Structure With Round-Trip Suite

## Description
Implement the write path for the `cns` package: serialize `StateDef`/`Controller` data back into valid `.cns` text, preserving block ordering and comments, and prove round-trip fidelity against a real fixture — required by CLAUDE.md's testing conventions for write-path code.

## Acceptance Criteria
- [ ] Serializing parsed `.cns` data produces valid `.cns` text that re-parses (via item 020) into an equivalent structure
- [ ] A realistic `.cns` fixture round-trips through parse → serialize with comments and block ordering preserved
- [ ] Round-trip equivalence is asserted in a test, not just eyeballed

## Notes
Completes the `cns` read+write cycle.
