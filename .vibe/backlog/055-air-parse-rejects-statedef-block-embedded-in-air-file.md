---
status: todo
---
# air.Parse Rejects a Statedef Block Embedded in an .air File

## Description
A real `.air` file (One Piece "Luffy", `~/workspace/ikemen-quick-versus/chars/One Piece/Luffy/LUFFY.AIR`) contains a `[Statedef 1060]` header partway through the file, inside what is otherwise a normal animation file. `air.Parse` treats any bracket line that isn't `[Begin Action N]` as a malformed action header and errors (`air: line 33: malformed action header "[Statedef 1060]"`), so the whole file fails to load. Surfaced while runtime-verifying backlog item 050's case-insensitive file resolution fix — item 050's own fix works correctly (the file is now found and opened despite the `.def` referencing it under different casing), this is a separate, pre-existing `air.Parse` gap, same precedent as items 052/054.

## Acceptance Criteria
- [ ] A bracket line inside an `.air` file that isn't a recognized `[Begin Action N]` header is tolerated (e.g. skipped) rather than treated as a malformed action header, mirroring how `cns.Parse`/`def.Parse` already tolerate unrecognized section headers
- [ ] A genuinely malformed `[Begin Action ...]` header (e.g. non-numeric or missing action number) still returns the existing descriptive, line-numbered error
- [ ] `~/workspace/ikemen-quick-versus/chars/One Piece/Luffy/Luffy.def` loads successfully end-to-end via `character.Load` after the fix

## Notes
None.
