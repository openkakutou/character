---
date: 2026-08-06
status: accepted
---
# .air comments are stripped by Parse, not preserved in the read-path model

**Context:** Hardening `.air` parsing (backlog item 003) against malformed/unusual input, including comment lines (`;`). The `.air` text format allows `;` to start a comment, either as a whole line or trailing after real content on the same line.

**Decision:** `Parse` strips comments (from the first `;` to end of line, whole-line or trailing) before interpreting a line, and does not record their text or position anywhere in the `Animation`/`Frame` model.

**Reason:** Per CLAUDE.md's design constraint, the read path exposes a minimal, stable, pure-data API — comment text has no meaning to a consumer like a future game engine. Preserving comments (and their exact position) for a low-diff round-trip is explicitly a write-path concern, scoped to later backlog items (004/005) that implement `.air` serialization. Keeping `Parse` comment-stripping-only avoids leaking format-preservation logic into the read-only model, per the read/write separation constraint.

**Rejected alternatives:** Attaching comment text to the nearest `Animation`/`Frame` now, anticipating round-trip needs — rejected because it would put write-only concerns into the pure-data type the read-only API exposes, before the write path's actual preservation strategy is even designed.
