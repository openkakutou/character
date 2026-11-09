---
status: done
---
# Accept symbolic and relative Statedef numbers in .zss headers

## Description
Validating item 060 against the 10 `.zss` files Ikemen GO itself ships (`data/`) showed 5 rejected by `zss.Parse` as "malformed block header": the state number can be a named constant (`[StateDef const(StateDizzyBirdsHelper); ...]` in dizzy/guardbreak/tag.zss) or relative (`[StateDef +1]` in system/demo.zss). Such a character still loads (non-fatal per-file error, ADR 032) but those scripts are unusable. Accept both forms, keeping the raw number token so engine can resolve it; amend ADR 027 for the model change.

## Acceptance Criteria
- [ ] `[StateDef const(Name); ...]` and `[StateDef +1]` parse, the raw number token is kept
- [ ] The 10 Ikemen GO `data/*.zss` files parse and round-trip byte-exact
- [ ] Unmodified save stays byte-exact; JSON contract documented

## Notes
High priority: blocks real Ikemen GO compatibility (MUST). Source of the failing headers: ikemen-engine/Ikemen-GO@063a060 `data/`.
