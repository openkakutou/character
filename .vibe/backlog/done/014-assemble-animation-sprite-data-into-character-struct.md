---
status: done
depends_on: [013]
---
# Assemble Animation and Sprite Data Into Character Struct

## Description
Extend the root `character.go` `Character` struct to compose parsed `Animation` and `Sprite` data (via item 013's resolution) into the single unit a library consumer (editor, engine) works with, per the architecture already documented in CLAUDE.md. Must not leak any write-only (formatting-preservation) types into the exposed `Character` API.

## Acceptance Criteria
- [ ] `Character` exposes animations and sprites through the `air`/`sff` read-path types only
- [ ] A `Character` assembled from a real air+sff fixture pair exposes resolved frame→sprite references correctly
- [ ] No write-only type (e.g. a formatting-preservation AST) is reachable from the public `Character` API

## Notes
Closes out the air+sff integration phase before `def` work begins.
