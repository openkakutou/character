---
status: done
---
# Cns Parse Rejects Non-Key-Value Lines Inside Blocks

## Description
`cns.Parse`'s `Parse` (in `cns/parser.go`) hard-errors with `cns: line N: malformed key=value line %q` on any non-header content line inside a `[Statedef ...]`/`[State ...]` block that has no `=` character — via `parseKeyValueLine` returning `ok == false`. Real MUGEN/Ikemen character files routinely contain such lines (a truncated key left over from editing, a decorative separator, or a comment that lost its leading `;`), and real engines evidently tolerate them, the same way `cns.Parse` already ignores an unrecognized *key*. This is the largest failure class in the local real-character corpus (`~/workspace/ikemen-quick-versus/chars`) after item 042: 49 of 717 real characters fail on it.

## Acceptance Criteria
- [ ] A non-header content line inside a `[Statedef ...]`/`[State ...]` block that has no `=` (and thus isn't a valid key=value pair) is ignored by `cns.Parse` instead of returning an error, the same way an unrecognized key already is
- [ ] A bracket-line section header (`[...]`) that is malformed still returns a descriptive, line-numbered error exactly as today — this change only relaxes handling of non-header content lines
- [ ] Existing `StateDef`/`Controller` data produced from a file with only well-formed key=value lines is unchanged
- [ ] `~/workspace/ikemen-quick-versus/chars/Arcana Heart/Aino 2/Aino.def` (which references `Soul_Aino.cns`, failing today at line 4622 on the bare word `getpower` with no `=`) parses successfully after the fix

## Notes
Second real fixture: `~/workspace/ikemen-quick-versus/chars/Arcana Heart/Saki Tsuzura/Saki Tsuzura.def` (references `AI_by_J-J/saki.cns`, failing today at line 1980 — a Shift-JIS-encoded comment block whose lines are missing their leading `;`). A third example produces the literal single-character line `:` (`~/workspace/ikemen-quick-versus/chars/Arcana Heart/Mei Fang/Mei Fang.def`, referencing `Mei-Fang_R_CNS/Mei-Fang_R.cns` line 1914). See also item 044, the identical issue in `def.Parse` — same root cause, different package.
