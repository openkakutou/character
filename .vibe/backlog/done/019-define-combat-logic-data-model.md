---
status: done
---
# Define Combat-Logic Data Model

## Description
Define the minimal pure-data model for the `cns` package: `StateDef` (state number, physics/type flags, transition targets) and `Controller` (a state controller: type, trigger conditions, parameters) as scaffolding — not a full CNS expression engine. Establishes the read-path vocabulary for item 020.

## Acceptance Criteria
- [ ] `StateDef` and `Controller` structs are exported with documented fields covering the common `.cns` Statedef header and controller shape
- [ ] Zero-value structs compile and behave predictably, covered by a zero-value test
- [ ] No expression parsing/evaluation logic is included — parameters and triggers are stored as pure data, not evaluated

## Notes
Explicitly scoped as scaffolding per CLAUDE.md ("`.cns` can wait" — lowest priority, minimal viable model).
