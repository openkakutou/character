---
status: done
---
# `cns.Parse` Errors On An Unclosed Bracket Line That Isn't A Header Attempt

## Description
`cns.Parse` currently treats any line starting with `[` and missing a closing `]` as a malformed header attempt, and returns a "malformed section header" error unless the line also matches `[Statedef ...]`/`[State ...]`'s own attempt patterns (in which case it's recovered instead, per backlog item 042). A line that starts with `[` but is neither a valid header nor an attempt at one of *those two specific* headers (e.g. a bare `[` used as a decorative banner character, with no other bracket content at all) still hits the error path — inconsistent with how a *closed* unrecognized bracket line (e.g. `[foo]`) is already skipped silently as out-of-scope content.

Discovered via a 520-file real-`.cmd`-file corpus scan while implementing backlog item 036 (this repo's `cmd` package runs `cns.Parse` against `.cmd` file text to parse the shared `[Statedef -1]`/`[State ...]` block — see `.vibe/decisions/025`): King of Fighters' "Shun Ei" character's `Command.cmd` has a lone `[` on its own line, used purely as a section-banner decoration (surrounded by `;===...` comment lines), with no closing `]` anywhere nearby. Real MUGEN/Ikemen engines evidently tolerate this; `cns.Parse` currently does not.

## Acceptance Criteria
- [ ] A bracket line missing its closing `]` that is not a "[Statedef ...]"/"[State ...]" attempt (per the existing `statedefAttemptPattern`/`stateAttemptPattern`) is skipped without erroring, the same way a closed unrecognized section already is
- [ ] The Shun Ei `Command.cmd` real-world case (or an equivalent trimmed fixture) parses without error
- [ ] Existing "malformed section header" error coverage for a *genuine* typo in one of this package's own two known headers is unchanged

## Notes
Out of scope for item 036 itself — that item's own `cmd` package has no dependency on this fix (it already tolerates the case at its own layer for `.cmd`-specific headers; this gap is purely inside `cns.Parse`, exercised incidentally because `cmd.Parse` delegates to it). Root cause differs from item 042/043's own bracket-handling fixes, so filed separately rather than folded into either.
