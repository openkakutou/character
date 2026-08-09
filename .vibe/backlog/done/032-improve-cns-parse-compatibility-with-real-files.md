---
status: done
---
# Improve .cns Parse Compatibility With Real Files

## Description
`cns.Parse` currently rejects the majority of real-world `.cns` files. Discovered while confronting backlog item 022's `Character`/`.cns` wiring against a large, real character corpus: spot-checking `cns.Parse` directly against 604 real `.cns` files referenced by real characters' `.def` files, only 213 (35%) parsed without error. Two distinct root causes were found:

1. **State headers without a usable state number.** Real files sometimes write `[State <label>]` with no number at all (e.g. `[State removeexplod]`, `[State PlaySnd]`, `[State var]`) instead of `[State N]`/`[State N, label]`. Real MUGEN/Ikemen engines tolerate this; `cns.Parse`'s current header-detection strategy (see `.vibe/decisions/012-cns-parse-header-detection-strategy.md`) does not — it returns a line-numbered error.
2. **Statedef header fields holding expressions instead of literal integers.** Numeric header fields such as `anim` or `poweradd` sometimes hold MUGEN trigger/expression syntax (e.g. `anim = IfElse(ceil(lifemax/2) < life ,181,182)`, `poweradd = ifelse(PrevStateNo = 9000, 0, 20)`) rather than a plain integer. `cns.Parse` currently requires `strconv.Atoi` to succeed on these fields and errors otherwise.

## Acceptance Criteria
- [x] A design decision is made (and recorded as an ADR) on how far `cns.Controller`'s existing "unevaluated data" philosophy should extend to `StateDef`'s typed header fields, to resolve root cause 2 — this needs a real design conversation, not a quick patch: it directly affects the shape of `StateDef`'s public fields — see `.vibe/decisions/023-statedef-numeric-header-fields-unevaluated-expression-escape-hatch.md` (new `StateDef.HeaderExprs map[string]string` escape hatch)
- [x] A decision is made on `cns.Parse`'s header-detection strategy for a `[State <label>]` header with no number, to resolve root cause 1, consistent with (or superseding) `.vibe/decisions/012-cns-parse-header-detection-strategy.md` — see `.vibe/decisions/022-cns-parse-state-header-accepts-any-label.md` (supersedes ADR 012 for `[State ...]` only, not `[Statedef ...]`)
- [x] Once decided: real-world `.cns` files exhibiting either pattern parse successfully
- [x] A file that is genuinely malformed in some other way still returns a descriptive, line-numbered error

## Notes
No need to vendor real fixtures for this — the corpus scan already establishes this is common; synthetic snippets reproducing each pattern are enough to write tests against. Flag both root causes for the design conversation up front; don't presuppose the fix for either. See the real-corpus findings recorded in backlog item 022 (now in `.vibe/backlog/done/`) for the original scan results.
