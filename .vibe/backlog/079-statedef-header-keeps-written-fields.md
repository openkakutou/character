---
status: todo
---
# Statedef Header Keeps Written Fields

## Description
`cns.StateDef` cannot tell an omitted `ctrl` from `ctrl = 0`, nor an omitted `anim` from `anim = 0`. It also drops `velset`. Ikemen GO treats every header field except the persist flags as optional, and evaluates each written one as an expression at state entry (compiler.go:7020-7130, bytecode.go:5444-5512, commit `54d4148a`). An omitted `ctrl` keeps the fighter's control. engine needs to know which fields were written, with their source text, as engine's `.zss` header conversion already provides (engine#055).

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO compatibility (roadmap decision `038`).

## Acceptance Criteria
- [ ] A parsed StateDef records which header fields were written, with each field's raw source text (anim, ctrl, velset, poweradd, juggle, facep2, sprpriority, the persist flags)
- [ ] The existing typed fields keep their current values, so current consumers do not change
- [ ] The new data survives the JSON shape the WebAssembly build returns

## Notes
Needed by engine#055 (2026-11-19): common1.cns states 0 and 20 omit `ctrl`, and the engine currently removes control when they are entered without `ctrl = 1`.
