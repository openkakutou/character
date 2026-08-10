---
status: in_progress
---
# Cns Parse Rejects Section Header Missing Closing Bracket

## Description
`cns.Parse` hard-errors on any line starting with `[` that does not also end with `]` (`cns/parser.go`: `strings.HasPrefix(line, "[")` followed by an unconditional `!strings.HasSuffix(line, "]")` check), with no tolerance at all — unlike the two real-world compatibility gaps already closed by backlog item 032 (missing state number, expression-valued numeric fields). Discovered while verifying backlog item 041's fix against a real community character (Bardock, provided by the user via `character-viewer-web`): its `.cns` file has `[State 110, 1` (missing the closing `]`) instead of `[State 110, 1]`, and the whole file fails to load because of it. Real MUGEN/Ikemen engines are known to tolerate this kind of typo in community-authored files.

## Acceptance Criteria
- [ ] A `.cns` file containing a `[State ...`/`[Statedef ...` header line missing its closing `]` parses successfully instead of erroring, at least for the common case where the line is otherwise a recognizable Statedef/State header
- [ ] The resulting `StateDef`/`Controller` data is equivalent to what a correctly-bracketed line would have produced
- [ ] A bracket line that is not recoverable as a Statedef/State header attempt (genuinely unrelated malformed content) still returns a descriptive, line-numbered error, consistent with `.vibe/decisions/022-cns-parse-state-header-accepts-any-label.md`'s existing "genuinely unrelated vs. a typo in the two header shapes this parser is responsible for" distinction

## Notes
Only one real-world example is in hand so far (the Bardock file's `[State 110, 1`); worth a quick corpus check (similar to item 032's) to see how common this specific typo is before deciding how far the tolerance should extend (e.g. also missing opening `[`?). Not part of item 041's scope — that item only concerned `.air` blank-frame sentinels.

**Update (full corpus scan, 2026-08-10):** confirmed as the single most common `character.Load` failure across the whole local corpus (`~/workspace/ikemen-quick-versus/chars`) — 109 of 717 real `.def` files (~15%) fail on exactly this pattern, all with the identical message shape `cns: line N: malformed section header "..."`. Additional real fixture: `~/workspace/ikemen-quick-versus/chars/Capcom/Commando (CVS)/Commando.def` fails on `cvscommando.cns` with `[State -2, Blocking-12` (missing `]`) — same root cause as the Bardock case, confirms the fix should not be scoped narrowly to one character's file.
