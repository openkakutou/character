---
date: 2026-08-06
status: accepted
---
# .sff v1 header/sprite table parsing uses a separate low-level type, not the public Sprite/SpriteGroup model

**Context:** Item 007 parses the MUGEN `.sff` v1 file header and its (group, image) sprite index table, resolving each entry to the file offset of its pixel data — but explicitly does not decode pixel data yet (item 008). The public read-path model (`Sprite`/`SpriteGroup`, defined in item 006) includes `Width`/`Height`, which for v1 are only knowable by reading the embedded PCX pixel data's own header — out of scope here — and has no field to carry a file offset/length, which the not-yet-written pixel decoder (item 008) needs to locate each sprite's data.

**Decision:** Introduce a version-specific, package-internal-to-the-format-detail (but exported, since `sff` stays a single flat package per the existing architecture) type pair, `V1Header` and `V1SpriteTable`/`V1SpriteEntry`, produced by `ParseV1`. These carry the header metadata (version, group/image counts, palette-sharing flag) and, per sprite, its (group, image) key, axis point, palette-sharing flag, linked-sprite index, and file offset/length of its pixel data. They are not `Sprite`/`SpriteGroup` and are not assembled into one in this item.

**Reason:** Keeps the pure-data read-path model (`Sprite`/`SpriteGroup`) free of file-offset bookkeeping and of a partially-populated (zero `Width`/`Height`) state that would be misleading to a consumer. A later item (the v1 pixel decoder, 008, or the cross-format assembly, 014) is the natural place to fold a decoded `V1SpriteEntry` plus its decoded pixel dimensions into a `Sprite`.

**Rejected alternatives:**
- Populating `Sprite`/`SpriteGroup` directly from the header/table, leaving `Width`/`Height` at zero and dropping the file offset/length needed by item 008 — rejected because it produces a `Sprite` that silently lies about its own dimensions and loses information the very next backlog item needs.
- Adding `Offset`/`Length` fields to the public `Sprite` struct — rejected because those are v1-specific (and even there, an implementation detail of file layout, not sprite metadata), which would violate the model's stated version-agnostic design constraint.
