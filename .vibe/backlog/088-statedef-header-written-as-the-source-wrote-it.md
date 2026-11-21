---
status: todo
---
# A Statedef header is written as the source wrote it

## Description
Found by the audit of 2026-11-21 (medium). `Written` and the typed header fields hold the same values (`cns/statedef.go:147`), and each side is read by different code: `headerFields` (`cns/serializer.go:81-96`) always writes all 12 fields and never `velset`, while the engine reads `velset` only from `Written` (engine `tick.go:1686`). A Statedef written fresh or inserted by `SerializeEdited` (`cns/edit.go:451`) loses `velset` and gets `ctrl = 0` / `anim = 0` / `movetype = ` for fields the source left out, where Ikemen GO keeps the current value. The round-trip tests compare through `withoutWritten()` and one pins `ctrl = 0`.

🔒 MUST: Ikemen GO behaviour (roadmap decision `038`).

## Acceptance Criteria
- [ ] Serialize writes only the header fields the model says were written, `velset` included
- [ ] Round-trip tests compare `Written` too; the `ctrl = 0` expectation goes
