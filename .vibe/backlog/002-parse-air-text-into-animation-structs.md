---
status: todo
depends_on: [001]
---
# Parse .air Text Into Animation Structs

## Description
Implement a parser that reads MUGEN/Ikemen `.air` text format into the `Animation`/`Frame` model defined in item 001. Must handle `[Begin Action N]` headers, per-frame lines (group, image, x, y, time[, flip][, blendmode]), `Clsn2Default`/`Clsn1Default` declarations, indexed `Clsn[i]` box lines, and the `Loopstart` marker.

## Acceptance Criteria
- [ ] Parsing a valid multi-action `.air` sample produces `Animation` structs with correct action numbers, frame sequences, and Clsn boxes
- [ ] `Loopstart` is correctly recorded so looping animations are distinguishable from non-looping ones
- [ ] Frame optional fields (flip, blend mode) default correctly when omitted from a line

## Notes
Depends on 001 for the target model. Error handling for malformed input is split into item 003 to keep this item focused on the happy path.
