---
status: todo
---
# Character Load Gives Confusing Error For Def Without Files Section

## Description
`character.Load` (`load.go`) returns a confusing, low-level error (e.g. `opening animation file "...": read ...: is a directory`) when pointed at a `.def` file that has no `[Files]` section at all — most often an Ikemen GO "storyboard"/intro/ending screen definition (a `[SceneDef]`/`[Scene N]`-based file, a different sub-format entirely) that happens to sit inside a character's own folder alongside the real character `.def`. `def.Parse` correctly returns a zero-value `CharacterInfo` for these, per its own documented "skip unrecognized sections" behavior — but `AnimationFile`/`SpriteFile`/`ConstantsFile` then end up empty, so `filepath.Join(dir, "")` resolves to the containing directory itself, and the subsequent `os.Open` fails opaquely instead of naming the real problem. Found in 48 of 717 real corpus `.def` files exercised in the local corpus (`~/workspace/ikemen-quick-versus/chars`) — a real, common shape of input, even though it's a diagnostics/clarity issue rather than a parsing-correctness gap.

## Acceptance Criteria
- [ ] `character.Load` detects an empty `AnimationFile`, `SpriteFile`, or `ConstantsFile` before attempting to open it, and returns a clear error naming the `.def` path and stating it doesn't look like a character definition (e.g. missing `[Files]` entries), instead of a raw filesystem error
- [ ] A `.def` with a fully populated `[Files]` section continues to load exactly as today (no behavior change on valid input)
- [ ] `~/workspace/ikemen-quick-versus/chars/Marvel/Captain Marvel/Ending.def` produces the new, clear diagnostic instead of today's `is a directory` error

## Notes
Lower priority than items 043–050 in this batch: this doesn't unlock any currently-unreadable *character*, it only makes the failure mode clear when a caller (e.g. a folder-scanning UI in `character-editor`/`character-viewer-web`) happens to point `Load` at one of these auxiliary screen `.def` files instead of a real character `.def`. Second real fixture: `~/workspace/ikemen-quick-versus/chars/Misc/Popeye/ending.def`.
