---
status: todo
---
# Decisions and docs the parsers no longer follow

## Description
Found by the audit of 2026-11-21 (medium). Accepted decisions and `docs/api.md` describe errors the code no longer raises:
- Decision 012, ADR 026's context, the `cns.Parse` doc (`cns/parser.go:69-70`) and `docs/api.md:788`: a `[State]` before any Statedef is rejected; it is skipped, as Ikemen GO does (compiler.go:7228-7238).
- Decision 023: boolean header fields stay strictly typed; since item 046 they go to `HeaderExprs`.
- `docs/api.md:155-157`: a "Known gap" fixed by item 066; `docs/api.md:297-301`: errors item 055 removed.
- `docs/api.md` struct blocks and `.vibe/models.md` lack `MugenVersion`, `IkemenVersion`, `Movelists`, `movelistFiles`, `TriggerAll`, `TriggerGroups`, `NumberRaw`, and `LoadFS` is undocumented.

## Acceptance Criteria
- [ ] New decisions supersede 012 (the clause) and amend 023, citing the Ikemen GO evidence
- [ ] `docs/api.md` and `.vibe/models.md` describe today's behaviour and types
