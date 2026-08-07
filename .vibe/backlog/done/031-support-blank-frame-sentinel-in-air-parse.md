---
status: done
---
# Support Blank Frame Sentinel in .air Parsing

## Description
`air.Parse` currently rejects any frame line with a negative sprite group index as malformed. Real MUGEN/Ikemen `.air` files widely use `-1,-1` as the sprite group/image reference on a frame line to mean "show no sprite for this frame" (a legitimate, engine-recognized convention), not a data error. Discovered while confronting `character.Load` against a large real-character corpus for backlog item 022: roughly 40% of real characters' `.air` files use this sentinel, and all of them currently fail to load through this library even though the file is well-formed per real MUGEN/Ikemen engines.

## Acceptance Criteria
- [ ] A frame line using `-1,-1` (or `-1` in either the group or image position) as the sprite reference is parsed successfully instead of returning an error
- [ ] The resulting `Frame` represents "no sprite shown" in a way `ResolveSprite`/`SpriteResolver` can recognize directly, rather than attempting a lookup against `Sprites` that would fail
- [ ] A genuinely malformed frame line (e.g. a non-numeric field, or a negative value other than the `-1` sentinel) still returns the existing descriptive, line-numbered error
- [ ] A test fixture (synthetic `.air` snippet reproducing `-1,-1, 0,0, N`) covers the sentinel explicitly, alongside a case exercising `ResolveSprite` on such a frame

## Notes
No need to vendor real files — a synthetic fixture reproducing the sentinel pattern is enough. Touches `character/air`'s parser and, depending on how the "no sprite" case is represented, potentially `air.Frame`'s data model and `air.SpriteResolver`/`character.ResolveSprite`. See the real-corpus findings recorded in backlog item 022 (now in `.vibe/backlog/done/`) for more examples of characters exhibiting this pattern.
