---
status: done
---
# Update CLAUDE.md Project Context To Match Current Roadmap Naming

## Description
This repo's CLAUDE.md still describes a generically-named `editor` and a loosely-scoped future `engine`, both superseded by roadmap decisions since (`character-editor` naming, `engine`'s settled combat-simulation scope, the `sff` extraction). Low-priority documentation cleanup — no code change.

## Acceptance Criteria
- [x] "Project context" section names `character-editor` instead of generic `editor`
- [x] `engine`'s description matches its actual settled scope (combat simulation, not "a game engine in the spirit of Ikemen GO")
- [x] Mentions the `sff` extraction and this repo's dependency on it

## Notes
Purely descriptive; can be picked up whenever, not blocking anything.

## Resolved
2026-08-09: Picked up ahead of schedule at the Product Owner's prompt right after backlog item 035 (the `sff` extraction migration) shipped, rather than waiting. Rewrote every `<!-- keep -->` section (`Project overview`, `Project context`, `Scope of this repo`, `Design constraints`) to reflect: `character-viewer-web`/`character-editor` by name instead of generic `editor`; `engine`'s settled combat-simulation scope (roadmap decision `008`) instead of "a game engine in the spirit of Ikemen GO"; the `.sff` format now handled by the external `sff` repo dependency, not an internal sub-package. Also rewrote the (non-`keep`) `Architecture` section, which had drifted much further than this item's own scope — every format sub-package was still marked "not yet implemented" despite `def`/`air`/`cns` all being complete, and the tree diagram still listed a `sff/` sub-package that no longer exists locally — and fixed a matching stale reference in the `Review agents` table (`vibe:review-solid`'s rationale).
