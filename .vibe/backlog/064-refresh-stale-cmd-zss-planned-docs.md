---
status: todo
---
# Refresh Stale Planned Docs

## Description
CLAUDE.md still labels .cmd and .zss as planned (backlog items 036/037, sff item 035) and the layout tree lists them as planned, but cmd/ and zss/ packages exist and the README documents them.

## Acceptance Criteria
- [ ] CLAUDE.md file-format list and layout tree show cmd and zss as implemented, with the true remaining gap (Character wiring, items 059/060)
- [ ] No Planned wording remains for shipped packages
- [ ] /vibe:sync runs clean

## Notes
Verified in character/CLAUDE.md lines about .cmd and .zss.
- [ ] `docs/architecture.md:48`, `docs/api.md:1096-1097`, `.vibe/modules/cmd.md:3`, `.vibe/models.md:257, 301`, `.vibe/modules/root.md:10`, `.vibe/modules/wasm.md:9` no longer say `.cmd` / `.zss` are not wired; `models.md:257` names `Commands`, not `Scripts` (audit 2026-11-21)
