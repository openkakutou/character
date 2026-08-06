# Module: root

**Role:** Entry point of the library. Assembles the already-implemented `air`/`sff` sub-packages into a single `Character` struct — the unit a library consumer actually works with — and exposes frame-to-sprite resolution directly on it. `def`/`cns` are not yet implemented, so `Character` has no fields for them yet.
**Files:** `character.go`
**Exports:** `Character` (struct: `Name string`, `Animations []air.Animation`, `Sprites []sff.SpriteGroup`), `(*Character) ResolveSprite(frame air.Frame) (sff.Sprite, error)`
**Depends on:** `modules/air.md` (for `air.Animation`/`air.Frame` and to delegate resolution to `air.NewSpriteResolver`), `modules/sff.md` (for `sff.SpriteGroup`/`sff.Sprite`)

**Design decision:** `ResolveSprite` does not reimplement frame-to-sprite lookup — it builds an `air.SpriteResolver` from `c.Sprites` and delegates to it, keeping that logic owned by `character/air` (per `.vibe/decisions/008-air-sprite-resolution-lives-in-air-package.md`). Only `air`/`sff` read-path types are reachable from `Character`; no write-only (format-preservation) type is exposed.
