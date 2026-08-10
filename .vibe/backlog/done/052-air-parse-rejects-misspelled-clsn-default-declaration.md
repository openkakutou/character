---
status: done
---
# Air Parse Rejects Misspelled Clsn Default Declaration

## Description
`air.Parse`'s `parseClsnDeclarationHeader` (in `air/parser.go`) only recognizes the exact keywords `Clsn1Default`, `Clsn2Default`, `Clsn1`, `Clsn2` (case-insensitively) as Clsn declaration headers. A real `.air` file can contain an author typo in the keyword itself — e.g. `Clsn2deault: 1` instead of `Clsn2Default: 1` — which falls through to frame-line parsing instead and is rejected with `air: line N: malformed frame line %q: expected at least 5 comma-separated fields`. Found in 1 of 717 real characters in the local corpus (`~/workspace/ikemen-quick-versus/chars`), surfaced while verifying item 045's fix against `Donovan.def` (a second fixture that item's Notes expected to also parse, but which fails here for this unrelated reason instead).

## Acceptance Criteria
- [ ] `~/workspace/ikemen-quick-versus/chars/Darkstalkers/Donovan/Donovan.air` parses successfully after the fix (fails today at its own line 4413, `Clsn2deault: 1`)
- [ ] A genuinely unrelated bracket-free content line is still handled exactly as before — this change only recognizes this specific real-world typo, not an open-ended fuzzy match on every Clsn-like line
- [ ] Existing `Animation`/`Frame` data for files using the correctly spelled keywords is unchanged

## Notes
Decide the fix shape when implementing: either tolerate this specific typo (`Clsn[12]deault`) alongside the correct spelling, mirroring how `def.Parse`/`cns.Parse` tolerate other specific real-world authoring mistakes rather than attempting general fuzzy matching, or confirm no other misspelling variant appears elsewhere in the corpus before deciding how narrow the tolerance should be.
