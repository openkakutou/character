---
status: done
---
# Air Parse Frame Line Integer Fields Reject Trailing Garbage

## Description
`air.Parse`'s `parseFrameLine` (in `air/parser.go`) uses `strconv.Atoi` on each comma-separated frame-line field, which requires the *entire* field to be a valid integer. A field with trailing non-numeric content after a valid leading integer (e.g. `143 0` as one field, from a line like `18765, 143 0, 0, 2 ,,A`) is rejected as `air: line N: malformed frame line %q: invalid image: strconv.Atoi: parsing "143 0": invalid syntax`. Found in 14 of 717 real characters in the local corpus (`~/workspace/ikemen-quick-versus/chars`) — always the identical literal frame line `18765, 143 0, 0, 2 ,,A`, copy-pasted across otherwise-unrelated characters (different authors, different franchises), which strongly suggests real MUGEN/Ikemen engines tolerate this specific value at runtime via a more permissive, C-style leading-integer-prefix number scan (like `atoi`/`sscanf("%d", ...)`) rather than strict whole-field parsing — otherwise this widely-shared broken value would not still be circulating in working character releases.

## Acceptance Criteria
- [ ] A frame-line integer field (Group, Image, X, Y, Time) with a valid leading integer followed by non-numeric trailing content (e.g. `143 0`) parses using the leading integer, ignoring the trailing content, instead of returning an error
- [ ] A field that is *not* a valid integer at all (no usable leading digits) still returns the existing descriptive, line-numbered error
- [ ] `~/workspace/ikemen-quick-versus/chars/Dragon Ball/Goku/Goku.def` (referencing `goku.air`, failing today at line 5520 on `18765, 143 0, 0, 2 ,,A`) parses successfully after the fix
- [ ] `~/workspace/ikemen-quick-versus/chars/Guilty Gear/axl-kofa/axl-kofa.def` and `~/workspace/ikemen-quick-versus/chars/Guilty Gear/baiken-kofa/baiken-kofa.def` (same literal broken line in their own `.air` files) also parse successfully after the fix

## Notes
This item is inferred from corpus evidence (a broken value that still "works" across many unrelated real releases), not from documented spec — verify the exact tolerance real MUGEN/Ikemen applies (e.g. by checking Ikemen GO's own number-parsing source) during implementation before generalizing this behavior, to avoid silently accepting genuinely corrupted data elsewhere. If real engines turn out to reject this too, downgrade this item to "leave as an error" and close it as not applicable.
