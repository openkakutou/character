---
status: todo
---
# Accept `[Statedef N ...]` headers followed by extra text

## Description
Real characters write `[StateDef 620 - Air Akimbo Blaze]`; `cns.Parse` rejects the header, so 25 characters fail to load in the engine's conformance run while Ikemen GO loads them. Read the number and ignore the rest as Ikemen GO does, keeping the text byte-exact on save. Ground the behavior in Ikemen GO's source (commit `54d4148a`).

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO compatibility, verified against Ikemen GO (roadmap decision `038`).

## Acceptance Criteria
- [ ] The header form parses with its number
- [ ] Round trip keeps the original header text
- [ ] A header with no number still errors

## Notes
Found by the conformance harness (item 030, `docs/conformance.md`, 2026-11-17).
