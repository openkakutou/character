---
status: done
---
# Thread Full CharacterInfo Fields Through LoadBytes JSON Contract

## Description
`def.CharacterInfo` already parses richer fields (author, etc.) than what `LoadBytes`'s JSON contract currently exposes to WASM consumers — this gap was noted as a known limitation blocking `character-viewer-web`'s characteristics panel from showing full metadata. Thread every `CharacterInfo` field through to the JSON contract consumed by WASM.

## Acceptance Criteria
- [ ] Every field already parsed by `def.CharacterInfo` is present in `LoadBytes`'s JSON output
- [ ] Existing WASM consumers relying on the current (partial) contract are unaffected — this is additive, not a breaking rename
- [ ] A character `.def` missing optional metadata fields still loads successfully, with those fields empty/zero rather than erroring

## Notes
Raised as a known gap in `character-viewer-web`'s item 004 (Characteristics Panel).
