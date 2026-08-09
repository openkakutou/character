---
status: todo
---
# Def Parse Rejects Non-Key-Value Lines Inside Sections

## Description
`def.Parse`'s `Parse` (in `def/parser.go`) hard-errors with `def: line N: malformed key=value line %q` on any content line inside a recognized `[Info]`/`[Files]` section that has no `=` character — via `parseKeyValueLine` returning `ok == false`, mirroring the exact same limitation as item 043's `cns.Parse` issue. Real `.def` files sometimes contain a decorative separator line or a truncated key left over from editing. Found in 4 of 717 real characters in the local corpus (`~/workspace/ikemen-quick-versus/chars`).

## Acceptance Criteria
- [ ] A content line inside a recognized `[Info]`/`[Files]` section that has no `=` is ignored by `def.Parse` instead of returning an error, the same way an unrecognized key already is
- [ ] A malformed section header (`[...]` with no closing `]`) still returns a descriptive, line-numbered error exactly as today — this change only relaxes handling of non-header content lines within a known section
- [ ] Existing `CharacterInfo` data produced from a file with only well-formed key=value lines is unchanged
- [ ] `~/workspace/ikemen-quick-versus/chars/Disgaea/Etna/Etna.def` parses successfully after the fix (fails today at its own line 13, a bare `------------------------------------------------` separator line with no `=` right under `[Files]`)

## Notes
Second real fixture: `~/workspace/ikemen-quick-versus/chars/King of Fighters/K (XIII)/K.def`, failing today at its own line 4 (a bare, truncated `MugenVersion` key with no `=` under `[Info]`); same pattern also seen in `~/workspace/ikemen-quick-versus/chars/Samurai Shodown/Suija/Suija.def`. See item 043 for the identical issue in `cns.Parse` — same root cause, different package, likely the same fix shape (extend `Parse`'s handling of `parseKeyValueLine`'s `ok == false` case to skip rather than error, for a non-header line).
