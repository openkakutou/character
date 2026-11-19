---
status: done
---
# Keep Trigger Group Numbers

## Description
`cns.Controller.Triggers` is a flat list. It loses the `triggerall` / `trigger1` / `trigger2` … numbering, so a consumer cannot evaluate MUGEN's rule: every `triggerall` AND (`trigger1` OR `trigger2` OR …), each numbered group being the AND of its own lines. Ikemen GO keeps the groups (compiler.go stateCompileCNS, trigger handling). Most real characters rely on OR between groups. For example, `common1.cns:128-133` changes the walk animation on trigger1 OR trigger2, so the engine never plays it.

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO compatibility (roadmap decision `038`).

## Acceptance Criteria
- [x] A parsed controller exposes its `triggerall` lines and its numbered groups in order, each group with its own lines
- [x] The existing flat `Triggers` field keeps its current value, so current consumers do not change
- [x] The new data survives the JSON shape the WebAssembly build returns, and `.cmd` controllers get it too

## Notes
Needed by engine#061 (2026-11-19). The engine's decision 003 had recorded the flattening as a known gap.
