---
date: 2026-08-06
status: accepted
---
# .air comment-preserving round trip lives in a separate Document type, not in Parse/Serialize

**Context:** Backlog item 005 requires a round-trip test suite proving that parsing a real `.air` file and re-serializing it reproduces the original content — including comments — with no loss of frames, Clsn boxes, or `Loopstart` markers. ADR 002 already decided that `Parse` strips comments and the pure-data `Animation`/`Frame` model never records them, because the read path must stay the minimal, stable vocabulary a future game engine consumes. Those two requirements are in tension: comments cannot be preserved on write if nothing captured them on read.

**Decision:** Add a new `Document` type (`air/document.go`), separate from `Animation`/`Frame`, that `ParseDocument` builds by keeping the original source bytes alongside the same decoded `[]Animation` that `Parse` returns. `Document.Serialize` writes the retained source back out verbatim, so a `ParseDocument` → `Serialize` round trip on unmodified content reproduces the original file byte-for-byte, comments and all. `Parse`/`Serialize`/`Animation`/`Frame` are unchanged — ADR 002 still holds for the pure-data read path. This scoped `Document` does not attempt to regenerate output from an *edited* `Animations` slice while preserving unrelated comments/ordering — mutating `Document.Animations` after `ParseDocument` has no effect on what `Serialize` writes. That heavier per-line reconciliation (needed so an editor's partial edits produce a small diff instead of a full rewrite) is deliberately deferred to a future item once there is a concrete editing use case to design against.

**Reason:** Keeps the read-only model exactly as ADR 002 intended (no write-only concerns leak into `Animation`/`Frame`), while still giving the write path a genuine, testable round-trip guarantee for the common "load, don't touch, save" case — the actual claim backlog item 005 asks to be proven. Building full mutation-aware regeneration now would be speculative: no consumer yet edits a parsed `.air` file through this library, so the reconciliation strategy (how to keep a comment attached to a frame that moved, was deleted, etc.) has no real requirements to design against yet.

**Rejected alternatives:**
- Making `Parse` retain comments on `Animation`/`Frame` — rejected, reopens ADR 002 and leaks a write-only concern into the read-only API.
- A line-element model (`Document` decomposed into typed header/frame/comment elements) that regenerates each structural line from live `Animations` while replaying stored comments/blank lines around it — rejected for now as premature: it would speculatively solve an editing-preservation problem this item does not require, before any real editor workflow defines what "preserve on partial edit" should mean. Revisit if/when a concrete editing use case needs it.
