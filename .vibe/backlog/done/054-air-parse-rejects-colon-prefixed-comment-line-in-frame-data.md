---
status: done
---
# Air Parse Rejects Colon-Prefixed Comment Line In Frame Data

## Description
`air.Parse`'s frame-line reader treats a bracket-less content line inside a `[Begin Action N]` block as an attempted frame line and rejects it with `air: line N: malformed frame line %q` when it isn't a valid comma-separated field list. A real-world `.air` file (King of Fighters "Mai (98)", from the local corpus at `~/workspace/ikemen-quick-versus/chars`) has a line starting with `:` instead of the standard `;` comment marker (e.g. `: Kagerou No Mai (RBS version)`), which is not recognized as a comment and is instead misread as a malformed frame line, at `98Mai.air` line 2445. Discovered while runtime-verifying item 047's fix against this same fixture — a different root cause, not fixed as part of that item, following the same split-out precedent as item 052.

## Acceptance Criteria
- [x] Determine whether real MUGEN/Ikemen engines actually tolerate `:` as an alternate comment marker, or whether this is simply malformed/typo'd content in this one file that real engines also reject/ignore at runtime (check Ikemen GO's own comment-line source if accessible)
- [x] If tolerated by real engines: `air.Parse` recognizes a line starting with `:` the same way it already recognizes `;`-prefixed comment lines
- [x] If not tolerated (i.e. this is genuinely corrupt content real engines also choke on or silently skip via a different mechanism): document the finding and downgrade/close this item accordingly rather than silently accepting arbitrary garbage
- [x] `~/workspace/ikemen-quick-versus/chars/King of Fighters/Mai (98)/Mai.def` (referencing `98Mai.air`) parses successfully after the fix, if the fix is applicable

## Notes
Surfaced by item 047's runtime verification step, not from a corpus scan — only confirmed in this one file so far. Verify prevalence across the corpus during implementation before generalizing.

Resolution: checked Ikemen GO's own reference source (`anim.go`'s `ReadAnimFrame`/`ReadAnimation`). `:` is not literally recognized as a comment marker there either — but a content line that doesn't start with a digit or `-` is never treated as an attempted frame line in the first place: `ReadAnimFrame` returns `(nil, nil)` for it (no error), and the caller's switch falls through with no match, silently skipping the line. So the real engine tolerates this specific file via "not an attempt, so no error", not via a `:` comment rule. `air.Parse` now recognizes a line starting with `:` as a comment (mirroring `;`), which produces the identical observable outcome for this real-world quirk without loosening the stricter "attempted-but-malformed frame line" error path this repo already enforces elsewhere (e.g. `abc,0,...` still errors, unchanged). `Mai.def` now loads successfully (250 animations, 23 state defs), verified both via `go test` and a real `go run` smoke script that also confirmed an unrelated character (`kfm`) still loads correctly.
