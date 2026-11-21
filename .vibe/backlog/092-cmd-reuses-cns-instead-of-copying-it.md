---
status: todo
---
# `cmd` reuses `cns` instead of copying it

## Description
Found by the audit of 2026-11-21 (medium). `cmd` copies from `cns` the header-attempt patterns (`cmd/parser.go:40-41`, "duplicated here rather than exported") and the edit machinery (`editPlan`, `patchLine`, `pair` / `align`, `render`, the apply loop: `cmd/edit.go:159-298, 451-474`). The copies have already drifted (`p.repl[i] != ""` against `, ok`). A header spelling `cns` learns would be missed by `cmd.ensureImplicitStatedef`, which then adds a second `[Statedef -1]`.

## Acceptance Criteria
- [ ] `cns` exposes the header classifier (or an internal package holds it) and `cmd` uses it
- [ ] The line-edit helpers live once (e.g. `internal/lineedit`) and both packages use them
