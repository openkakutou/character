---
status: todo
---
# One mapping from `.def` to `Character`

## Description
Found by the audit of 2026-11-21 (medium). The `def.CharacterInfo` + parts → `Character` mapping (about 20 fields) is written twice, `load.go:122-143` and `load_bytes.go:130-151`, and `Character` repeats 11 fields of `CharacterInfo` (`character.go:24`). No test compares `Load` and `LoadBytes`. Item 038 was this exact drift: a field added to `Load` only is silently zero in the WASM build.

## Acceptance Criteria
- [ ] One constructor builds `Character` for both paths (or `Character` embeds `def.CharacterInfo`)
- [ ] A test loads one fixture through `Load` and `LoadBytes` and compares the results
