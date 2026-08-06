---
date: 2026-08-06
status: accepted
---
# .def parsing ignores unknown sections and orders palettes by number, not file order

**Context:** Implementing the `.def` INI-style text parser (backlog item 016) that populates `CharacterInfo` (item 015). Real-world `.def` files contain sections beyond `[Info]`/`[Files]` (e.g. `[Arcade]`, `[Palette Keymap]`, `[Quotes]`), and `CharacterInfo` has no field to hold their content.

**Decision:** `Parse` recognizes only the `[Info]` and `[Files]` sections (case-insensitively); any other section's lines are skipped entirely without validation, and parsing continues into the next known section rather than aborting. Within `[Files]`, `StateFiles` are collected in file-appearance order (matching the "st", "st1", "st2", ..." keys), while `Palettes` are collected and then sorted by their numeric suffix ("pal1", "pal2", ...) regardless of the order those keys appeared in the file, matching the field's documented contract ("in palette number order").

**Reason:** Mirrors the existing `air` package precedent (decision 002: comments are stripped, not preserved, by `Parse`) — `CharacterInfo` is the read-path, pure-data surface, and content that surface cannot represent is dropped rather than causing a hard failure. This satisfies backlog item 016's acceptance criteria ("unknown sections are preserved or reported without aborting the parse of known sections") via the simpler of the two allowed behaviors; format-preserving round-trip (comments, unknown sections, original ordering) is explicitly deferred to the write path (item 017), consistent with the project's read/write separation constraint.

**Rejected alternatives:**
- Erroring out on an unrecognized section — would make the parser unusable on any real-world `.def` file, which always carries sections beyond `[Info]`/`[Files]`.
- Capturing unknown section content on `CharacterInfo` itself — would leak write-only/format-preservation concerns into the read-path model, violating CLAUDE.md's read/write separation constraint (that belongs in item 017's write-path type, following the `air` package's separate-Document-type precedent, decision 003).
