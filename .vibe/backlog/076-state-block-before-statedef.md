---
status: todo
---
# Tolerate a `[State]` block before any `[Statedef]`

## Description
4 characters fail to load because a `.cns`/`.st` file has a `[State ...]` block before the first `[Statedef]`. Do what Ikemen GO does with it (skip or attach), cited. Ground the behavior in Ikemen GO's source (commit `54d4148a`).

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO compatibility, verified against Ikemen GO (roadmap decision `038`).

## Acceptance Criteria
- [ ] The real-world form loads as in Ikemen GO
- [ ] Round trip byte-exact

## Notes
Found by the conformance harness (item 030, `docs/conformance.md`, 2026-11-17).
