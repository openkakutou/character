---
date: 2026-08-09
status: accepted
supersedes: 012
---
# cns.Parse accepts any [State ...] header content, not just a numeric state number

**Context:** ADR 012 established that a bracket line starting with the "state" keyword must match `[State N]` (a numeric index, optional trailing ", label" comment) or else it is rejected as a malformed header. Item 032's real-file corpus scan found this is too strict: real MUGEN/Ikemen `.cns` files widely write `[State <label>]` with no number at all (e.g. `[State removeexplod]`, `[State PlaySnd]`, `[State var]`), which real engines accept without complaint — this pattern was one of two root causes behind only 35% of a 604-file real corpus parsing successfully.

**Decision:** `[State ...]` headers no longer require (or capture) a numeric index at all: any bracket line starting with the `state` keyword followed by more content (or immediately closing) is accepted as a controller header, and its content is discarded entirely — the controller that follows attaches to whichever Statedef is currently open. This is not a new loss: `cns.Controller` already never stored the `[State N]` header's own number (see ADR 011/012's own note that it is "not stored on Controller") — the header's content was already write-only decoration `Parse` threw away, so relaxing what shape it must be in doesn't remove anything the model was keeping. `[Statedef N]`'s own header is unaffected and still requires a numeric index: `StateDef.Number` is a real, load-bearing field with no discard-precedent to lean on.

**Reason:** MUGEN/Ikemen engines effectively treat `[State ...]`'s bracket content as a comment/label, using only the fact that a `[State` block started to know a new controller begins under the current Statedef — this parser already modeled that by discarding the number, so requiring the header text to still look numeric was an artificial restriction this model never actually relied on, not a real correctness boundary. Making the parser this permissive also removes the "malformed State header" error path entirely (it becomes unreachable, since anything that used to reach that check now already matches the broadened valid-header pattern) — the dedicated `stateAttemptPattern` regex is removed as dead code along with it.

**Rejected alternatives:**
- Special-case only the "no number, just a label" shape (`[State removeexplod]`) while still requiring a valid `[State N]`/`[State N, label]` shape otherwise — rejected as more code for less coverage: real files legitimately mix both shapes, and there is no downstream consumer of the `[State]` header's own text to justify parsing its shape at all once the number is confirmed unused.
- Keep requiring a numeric prefix and reject non-numeric ones as an error, asking authors to fix their files — rejected: these are real, unmodified files widely produced and accepted by real MUGEN/Ikemen tooling; this library's job is to read them as they are, not as a stricter spec says they should be.
