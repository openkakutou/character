---
status: done
---
# Air Parse Clsn Box Line Regex Rejects Whitespace Before Bracket

## Description
`air.Parse`'s `clsnBoxLinePattern` regex (in `air/parser.go`) requires `Clsn1`/`Clsn2` to be followed immediately by `[index]` with no whitespace in between (`^Clsn[12]\[\d+\]\s*=\s*...`). Real-world `.air` files sometimes have a space before the bracket (e.g. `Clsn2 [0] = -17, -97, 18, 2`), which real MUGEN/Ikemen engines tolerate but this parser rejects as `air: line N: malformed Clsn box line %q`. Found in 3 of 717 real characters in the local corpus (`~/workspace/ikemen-quick-versus/chars`).

## Acceptance Criteria
- [ ] `clsnBoxLinePattern` accepts optional whitespace between the `Clsn1`/`Clsn2` keyword and the `[index]` bracket
- [ ] A `Clsn1[0] = ...`/`Clsn2[0] = ...` line with no whitespace before the bracket still parses exactly as today
- [ ] A genuinely malformed Clsn box line (wrong coordinate count, non-numeric coordinate) still returns the existing descriptive, line-numbered error
- [ ] `~/workspace/ikemen-quick-versus/chars/King of Fighters/Mai (98)/Mai.def` (referencing `98Mai.air`, failing today at line 1631 on ` Clsn2 [0] = -17, -97, 18, 2`) parses successfully after the fix

## Notes
Small, low-risk regex fix — the Statedef/State header patterns in `cns/parser.go` already tolerate similar flexible internal whitespace (`\s*`), so this brings `air.Parse`'s Clsn box pattern in line with that established convention.
