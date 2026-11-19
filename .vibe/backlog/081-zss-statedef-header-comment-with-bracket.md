---
status: todo
---
# Zss Statedef Header Comment with Bracket

## Description
`zss.Parse` ends a multi-line `[StateDef …]` header inside a `#` comment that holds `;` or `]` (kfm_zss `kfm.zss:212-222`). State 200's header then swallows state 210's, and "state 200" runs state 210's body. Ikemen GO strips comments before it reads the header (compiler.go stateCompileZSS / scanStateDef, commit `54d4148a`).

🔒 MUST: Ikemen GO compatibility (roadmap decision `038`).

## Acceptance Criteria
- [ ] A `#` comment inside a multi-line Statedef header, holding `;` or `]`, does not end the header
- [ ] Ikemen GO's `kfm.zss` gives state 200 its own header and body
- [ ] A test uses the real `kfm.zss` lines

## Notes
Found by engine#040 (2026-11-19). Blocks the engine end-to-end scenario where the `.zss` Kung Fu Man does a light punch.
