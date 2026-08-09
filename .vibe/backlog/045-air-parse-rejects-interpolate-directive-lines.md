---
status: todo
---
# Air Parse Rejects Interpolate Directive Lines

## Description
`air.Parse` (in `air/parser.go`) only recognizes frame lines, `Clsn1/2(Default)` declarations, and the `loopstart` marker as valid non-header content inside a `[Begin Action N]` block. Ikemen GO's `.air` format also supports animation interpolation directive lines (`Interpolate Offset`, `Interpolate Blend`, `Interpolate Scale`, `Interpolate Angle`), which real-world files use between frame lines to smoothly transition a property across an animation. `air.Parse` currently misreads these as a frame line and rejects them with `air: line N: malformed frame line %q: expected at least 5 comma-separated fields`. Found in 23 of 717 real characters in the local corpus (`~/workspace/ikemen-quick-versus/chars`) — the largest `air.Parse` failure class.

## Acceptance Criteria
- [ ] `air.Parse` recognizes `Interpolate Offset`, `Interpolate Blend`, `Interpolate Scale`, and `Interpolate Angle` lines inside a `[Begin Action N]` block and accepts them without erroring, the same way it already special-cases `loopstart`
- [ ] A genuinely malformed frame line (wrong field count, non-numeric field) still returns the existing descriptive, line-numbered error — this change only recognizes the specific Interpolate directive keywords, not an open-ended relaxation of frame-line validation
- [ ] `~/workspace/ikemen-quick-versus/chars/BlazBlue/Ragna/Ragna.def` (referencing `Ragna.air`, failing today at line 5210 on `Interpolate Blend`) parses successfully after the fix
- [ ] Existing `Animation`/`Frame` data for files without Interpolate directives is unchanged

## Notes
It is not yet decided (nor required by this item) whether Interpolate directive content needs to be represented on the `Animation`/`Frame` read-path model — per the existing "read-path model can't hold everything yet, so drop what it can't represent" pattern already used elsewhere in this package (see `.vibe/decisions/`), simply recognizing and skipping the line (like `loopstart` line handling, which also isn't stored beyond `LoopStart`) is sufficient to satisfy this item; representing the interpolation data itself can be a separate, later item if a consumer needs it. Second real fixture area: `~/workspace/ikemen-quick-versus/chars/Darkstalkers/Donovan/Donovan.def`.
