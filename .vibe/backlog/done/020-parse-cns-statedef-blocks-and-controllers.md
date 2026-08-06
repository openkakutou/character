---
status: done
depends_on: [019]
---
# Parse .cns StateDef Blocks and Basic Controllers

## Description
Implement a parser for `.cns` text format covering `[Statedef N]` blocks and basic state controller (`[State N]`) blocks into the `StateDef`/`Controller` model (item 019), without evaluating trigger expressions.

## Acceptance Criteria
- [ ] Parsing a valid `.cns` sample with multiple Statedef/State blocks produces correctly structured `StateDef` and `Controller` data
- [ ] Trigger and parameter values are captured as raw data (strings/key-value pairs), not evaluated
- [ ] Malformed block headers return a descriptive error identifying the offending line

## Notes
Depends on 019 for the target model.
