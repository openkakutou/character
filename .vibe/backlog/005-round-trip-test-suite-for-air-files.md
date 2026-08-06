---
status: todo
depends_on: [002, 004]
---
# Round-Trip Test Suite for .air Files

## Description
Build a round-trip test suite proving that parsing a real `.air` file and re-serializing it reproduces the original content byte-for-byte (or semantically identical, including comments and line ordering, where byte-exact isn't feasible). Required by CLAUDE.md's testing conventions for all write-path code, since noisy diffs on save would hurt community collaboration on character files.

## Acceptance Criteria
- [ ] At least one realistic multi-action `.air` fixture round-trips through parse → serialize with no loss of frames, Clsn boxes, or `Loopstart` markers
- [ ] Comments present in the source file are preserved in the round-tripped output
- [ ] A round-trip diff (or explicit semantic-equivalence check) is asserted in a test, not just eyeballed

## Notes
Depends on both the parser (002) and the serializer (004).
