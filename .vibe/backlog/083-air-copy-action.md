---
status: todo
---
# Read `Copy Action N` in `.air` actions

## Description
Ikemen GO accepts a `Copy Action N` line in an action (and in a lifebar's `[Begin Action]` blocks): the action takes the frames of action N (`src/anim.go` `compileAction`, @54d4148a). `air.Parse` rejects the line today, so `lifebar` drops such a block with a warning (its `shortcut:` in `parse.go`), and characters using it lose the action.

🔒 MUST: as Ikemen GO resolves it (roadmap decision `041`).

## Acceptance Criteria
- [ ] An action with `Copy Action N` gets action N's frames, as in Ikemen GO
- [ ] A copy of a missing action behaves as in Ikemen GO
- [ ] `lifebar` can drop its `shortcut:` once released

## Notes
2026-11-21 audit: `Parse` already resolves `Copy Action` (v0.23.0); the edit path does not, see item 085. Close this item once `lifebar` drops its shortcut.
