---
date: 2026-11-08
status: accepted
---
# Format-preserving edits: per-format element alignment inside each Document

**Amends:** decision 028 — its edited-model fallback to a fresh `Serialize` and its rejected alternative "regenerate Document output around an edited model" are superseded; per-format exports and JSON normalization stay.

**Context:** Item 061 asks that saving a modified `.def`/`.air`/`.cns`/`.cmd`/`.zss` file rewrite only what changed, keeping untouched comments, blank lines and order, with added/removed sections at a documented, deterministic position. Until now `SerializeX` (decision 028) chose between the original bytes verbatim and a fresh `Serialize` of the whole model.

**Decision:** Each package's `Document` keeps, beside its source, an index from model elements to their source line ranges (never on the read-path model types). A new `(*Document).SerializeEdited(w, model)` aligns the parsed baseline with the edited model by identity key (longest common subsequence), pairs the unmatched elements between two anchors positionally as modifications (so a renumbered animation or state keeps its comments), and treats the surplus as insertions or deletions. Rules: comments and blank lines directly above a header belong to that element and are removed with it; an inserted element goes right after its nearest surviving predecessor in model order, or, when it has none, at the format's first-element position documented at the top of that package's `edit.go`; inside a modified element, changed `key = value` lines are patched in place (original key casing, trailing `;` comment and line ending kept), removed keys' lines are deleted, new keys go after the last parameter line, sorted; positional lines (frames, Clsn boxes, script bodies) are aligned by line content; an element whose patch is ambiguous (the same key twice) is regenerated alone with the format's `Serialize`. An appended `.def` `st`/`pal` key is numbered from the highest number already in the file plus one, so it never collides with sparse numbering. Generated lines use the file's dominant line ending. `SerializeX` keeps its signature and calls `SerializeEdited` whenever an original exists.

**Reason:** Fits the per-package write-path structure already in place and keeps each format's tolerance rules (decisions 012, 022, 027) untouched: the parsers are unchanged, and each `Document` builds its line index with its own scan of the source.

**Rejected alternatives:**
- **Shared lossless line layer for all formats:** one trivia rule written once, but couples five formats whose lines differ in kind (key/value, positional frame lines, raw script bodies).
- **Lossless syntax tree per format:** handles unforeseen Ikemen GO extensions for free, but means rewriting every parser and re-proving their real-file tolerances.
- **Stable element ids in the JSON contract:** would keep comments across any reordering, but changes the WASM contract (decision 019) for a gain positional pairing already covers in the common case.
