---
status: todo
---
# Air Parse Rejects Colon-Prefixed Comment Line In Frame Data

## Description
`air.Parse`'s frame-line reader treats a bracket-less content line inside a `[Begin Action N]` block as an attempted frame line and rejects it with `air: line N: malformed frame line %q` when it isn't a valid comma-separated field list. A real-world `.air` file (King of Fighters "Mai (98)", from the local corpus at `~/workspace/ikemen-quick-versus/chars`) has a line starting with `:` instead of the standard `;` comment marker (e.g. `: Kagerou No Mai (RBS version)`), which is not recognized as a comment and is instead misread as a malformed frame line, at `98Mai.air` line 2445. Discovered while runtime-verifying item 047's fix against this same fixture — a different root cause, not fixed as part of that item, following the same split-out precedent as item 052.

## Acceptance Criteria
- [ ] Determine whether real MUGEN/Ikemen engines actually tolerate `:` as an alternate comment marker, or whether this is simply malformed/typo'd content in this one file that real engines also reject/ignore at runtime (check Ikemen GO's own comment-line source if accessible)
- [ ] If tolerated by real engines: `air.Parse` recognizes a line starting with `:` the same way it already recognizes `;`-prefixed comment lines
- [ ] If not tolerated (i.e. this is genuinely corrupt content real engines also choke on or silently skip via a different mechanism): document the finding and downgrade/close this item accordingly rather than silently accepting arbitrary garbage
- [ ] `~/workspace/ikemen-quick-versus/chars/King of Fighters/Mai (98)/Mai.def` (referencing `98Mai.air`) parses successfully after the fix, if the fix is applicable

## Notes
Surfaced by item 047's runtime verification step, not from a corpus scan — only confirmed in this one file so far. Verify prevalence across the corpus during implementation before generalizing.
