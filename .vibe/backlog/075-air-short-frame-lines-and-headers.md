---
status: todo
---
# Accept the `.air` lines Ikemen GO tolerates

## Description
26 characters fail to load on an `.air` frame line with fewer than 5 fields, 4 more on a malformed action header or Clsn line, all loaded by Ikemen GO. Match Ikemen GO's tolerance (missing fields read as 0 by its `Atoi` path, see `src/anim.go` `ReadAnimFrame`) without hiding genuinely broken data. Ground the behavior in Ikemen GO's source (commit `54d4148a`).

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO compatibility, verified against Ikemen GO (roadmap decision `038`).

## Acceptance Criteria
- [ ] The real-world forms load with Ikemen GO's values
- [ ] Round trip byte-exact
- [ ] A line Ikemen GO itself rejects still errors

## Notes
Found by the conformance harness (item 030, `docs/conformance.md`, 2026-11-17).
