---
status: todo
---
# Cns Parse Boolean Header Fields Reject Trigger Expressions

## Description
`cns.Parse`'s `applyStatedefField`/`parseBoolField` (in `cns/parser.go`) hard-error via `strconv.ParseBool` when a Statedef header boolean field (`ctrl`, `facep2`, `hitdefpersist`, `movehitpersist`, `hitcountpersist`) is given a real MUGEN/Ikemen trigger expression instead of a literal `true`/`false`/`0`/`1`. This same file's *numeric* header fields (`anim`, `poweradd`, `juggle`, `sprpriority`) already tolerate an unevaluated expression by falling back to `StateDef.HeaderExprs` instead of erroring, per `.vibe/decisions/023-statedef-numeric-header-fields-unevaluated-expression-escape-hatch.md` — but that escape hatch was never extended to the boolean fields, which real files rely on just as often. Found in 21 of 717 real characters in the local corpus (`~/workspace/ikemen-quick-versus/chars`): 13 on `ctrl`, 8 on `facep2`.

## Acceptance Criteria
- [ ] A trigger-expression value on `ctrl`, `facep2`, `hitdefpersist`, `movehitpersist`, or `hitcountpersist` (anything that fails `strconv.ParseBool`) is stored verbatim via the same `HeaderExprs` escape hatch `applyIntOrExprField` already uses for numeric fields, instead of returning an error
- [ ] A literal boolean value (`0`, `1`, `true`, `false`, etc.) on any of these five fields still sets the corresponding `StateDef` field exactly as today
- [ ] `~/workspace/ikemen-quick-versus/chars/BlazBlue/Mai/Mai.def` (referencing `Mai.cns`, failing today at line 10878 with `ctrl = "0&(var(0):=Cond(parent,var(51)>0,parent,var(51),58))"`) parses successfully after the fix
- [ ] `~/workspace/ikemen-quick-versus/chars/Jojo's Bizarre Adventure/Avdol/Avdol.def` (referencing `avdul.cns`, failing today at line 448 with `facep2 = "1-(prevstateno=[100,119])"`) parses successfully after the fix

## Notes
Matches `applyIntOrExprField`'s existing pattern exactly (same file) — the fix is likely a `applyBoolOrExprField` helper mirroring it: try `strconv.ParseBool`, and on failure store the raw value in `HeaderExprs[name]` instead of returning an error. `HeaderExprs` is already a `map[string]string` keyed by field name, so no data-model change should be needed beyond reusing it for these five keys too.
