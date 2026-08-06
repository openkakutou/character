---
date: 2026-08-06
status: accepted
---
# Frame-to-Sprite resolution lives in the air package, as a SpriteResolver type

**Context:** `.air` `Frame` values reference sprites only by `(Group, Image)`; the actual `sff.Sprite` data lives in a separately loaded `[]sff.SpriteGroup`. Something needs to turn a `Frame` reference into the `sff.Sprite` it names, failing descriptively when no such sprite exists, and doing so the same way regardless of whether the sprites came from a `.sff` v1 or v2 file.

**Decision:** Add the resolution logic to the `air` package (a new file, e.g. `air/resolve.go`) as a `SpriteResolver` type: built once from a `[]sff.SpriteGroup` via a constructor that indexes sprites by `(Group, Image)`, exposing a `Resolve(frame Frame) (sff.Sprite, error)` method. The `air` package therefore gains an import of `sff`.

**Reason:**
- CLAUDE.md's own architecture diagram already documents `air/` as depending on `sff/` ("animation parsing (text, depends on sff)"), so this import direction was anticipated, not a new coupling.
- `sff` stays free of any dependency on `air`, so a consumer that only needs sprite data (e.g. a future sprite-only tool) never pulls in animation types — matching CLAUDE.md's "read path stays minimal" and no-circular-dependency constraint.
- The version-agnostic `Sprite`/`SpriteGroup` model (populated identically by both v1 and v2 parsers) already satisfies the "no version-specific branching required by the caller" acceptance criterion without any extra work here — the resolver only ever sees the shared shape.
- A dedicated `SpriteResolver` (built once, queried many times) avoids repeated O(n) linear scans over every frame of an animation against every sprite group, and gives a natural place to build the `(Group, Image)` index once.
- Returning `(sff.Sprite, error)` — not a `(sff.Sprite, bool)` "found" flag — matches the backlog's explicit requirement that a missing reference "fails explicitly rather than silently rendering blank": a caller cannot easily forget to check an `error` return the way a `bool` return invites being ignored.

**Rejected alternatives:**
- *Put resolution in the root `character` package instead of `air`.* Rejected for now: the root package's job (per CLAUDE.md) is assembling already-resolved data into `Character`, which is backlog item 014, not 013. Placing the lower-level `(group, image) → Sprite` lookup itself in `air` keeps it reusable by the root package (014) without forcing 014's assembly concerns into 013.
- *Put resolution in the `sff` package instead.* Rejected: `sff` would then need to import `air`'s `Frame` type, reversing the dependency CLAUDE.md's architecture diagram already establishes and creating a package that knows about animations just to resolve sprites.
- *Expose a plain lookup function `Resolve(frame Frame, groups []sff.SpriteGroup) (sff.Sprite, error)` instead of a resolver type.* Rejected: re-scans and re-indexes `groups` on every call, which is wasteful when resolving every frame of every animation of a character (the intended usage in 014).
