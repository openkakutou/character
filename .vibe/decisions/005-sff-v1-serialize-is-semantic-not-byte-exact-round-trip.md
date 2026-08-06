---
date: 2026-08-06
status: accepted
---
# .sff v1 serialization targets semantic round-trip, not byte-exact reproduction

**Context:** Item 009 adds the write path for `.sff` v1: turning sprite metadata plus PCX-encoded pixel data back into a valid binary file, and proving round-trip fidelity. Unlike `.air`/`.cns` (item 002's decision, `003-air-round-trip-via-separate-document-type.md`), CLAUDE.md's format-preservation constraint ("preserve as much of the original file structure as possible ... otherwise every save produces a huge, unreadable Git diff") is explicitly scoped to text formats read/edited by humans in version control. `.sff` is a binary sprite sheet with no meaningful textual diff, and nothing in this item's brief asks for byte-identical output — only that a serialized file re-parses into an equivalent structure with no pixel data loss and preserved shared-palette linkage.

**Decision:** `SerializeV1` takes a new write-only input shape (`V1WriteSprite`, mirroring `V1SpriteEntry`'s fields minus the file-offset bookkeeping `ParseV1` computes on read) and produces a fresh, valid `.sff` v1 layout: it always recomputes subfile offsets and header counts rather than trying to reproduce an original file's exact byte layout. Similarly, `EncodePCX` always RLE-encodes every scanline (never emits a literal byte where a run-encoded byte would also decode correctly) rather than trying to match whatever encoding heuristic produced some original PCX data. Round-trip fidelity is verified semantically: serialize, re-parse, re-decode, and compare the resulting metadata and pixel buffers, not the raw bytes.

**Reason:** A binary sprite format has no "diff-friendliness" concern the way `.air`/`.cns` text does — preserving an original file's exact subfile ordering or RLE encoding choices would add real complexity (tracking unused original offsets, matching an unknown original encoder's run/literal heuristic) for no benefit any actual consumer (editor, engine) needs; both only care that the sprites and pixels they get back are correct.

**Rejected alternatives:**
- Byte-exact round-trip via a `Document`-style write-only wrapper (the `.air` pattern) — rejected: no consumer need, and no diff-friendliness benefit for a binary format that isn't hand-edited or reviewed as text in version control.
- Reusing the read-only `V1SpriteEntry` as `SerializeV1`'s input — rejected: it carries `Offset`/`Length`, computed facts about a file being read, not inputs a writer should be trusted to supply; a caller passing a stale `Offset` would silently produce a corrupt file.
