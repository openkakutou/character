---
date: 2026-08-06
status: accepted
---
# cns.Parse distinguishes malformed Statedef/State headers from unrecognized sections

**Context:** `cns.Parse` (backlog item 020) reads `.cns` text into `StateDef`/`Controller`. A `.cns` file can contain bracket sections other than `[Statedef N]`/`[State N]` (e.g. `[Data]`, `[Clsn1Default]` in a full character constants file) — `def.Parse` already established, per ADR 009, that unrecognized sections are skipped rather than rejected. But unlike `.def`'s fixed, small set of section names, `.cns`'s "other" sections are open-ended, while `[Statedef ...]`/`[State ...]` headers are the two shapes this parser must get right.

**Decision:** A bracket line is classified in two passes: first check whether it matches a valid `[Statedef N]` or `[State N]` header (numeric index, optional trailing `, label` comment tolerated). If not, only then check whether it *starts with* the `statedef`/`state` keyword — if so, it's treated as a malformed attempt at that header and rejected with a line-numbered error; otherwise it's an unrecognized section (skipped without validating its content), matching ADR 009's precedent. A `[State N]` header encountered with no enclosing `Statedef` is also rejected, since a state controller with nothing to attach to cannot be represented.

**Reason:** Keeps the same "skip what we can't/don't need to represent" tolerance `def.Parse` established for genuinely unrelated sections, while still surfacing a clear, actionable error for the two header shapes this parser is actually responsible for getting right — a typo like `[Statedef abc]` should not be silently swallowed as "just some other section."

**Rejected alternatives:**
- Treat every bracket line as either a valid Statedef/State header or an error (mirrors `air.Parse`, which supports only one header shape) — rejected because real `.cns` files legitimately contain unrelated sections this parser has no reason to understand yet.
- Skip anything that isn't a valid Statedef/State header, with no distinction (mirrors nothing in the codebase) — rejected because it would silently accept typos in the two header shapes this item is specifically scoped to parse, undermining the acceptance criterion that malformed headers report an error.
