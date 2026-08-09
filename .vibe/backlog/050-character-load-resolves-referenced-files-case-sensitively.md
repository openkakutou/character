---
status: todo
---
# Character Load Resolves Referenced Files Case-Sensitively

## Description
`character.Load` (`load.go`) resolves `.def` `[Files]` entries via a direct `os.Open` on the literal referenced string, which is case-sensitive on Linux and the browser/WASM target. Many real community `.def` files reference a file under different letter casing than the actual filename on disk — a legacy of MUGEN/Ikemen historically running on case-insensitive Windows filesystems, which authors relied on without noticing the mismatch. Found in 12 of 717 real characters in the local corpus (`~/workspace/ikemen-quick-versus/chars`) — the second-largest cause of file-not-found failures after item 049's backslash-path issue.

## Acceptance Criteria
- [ ] When the exact-case referenced path doesn't exist, `character.Load`'s file resolution falls back to a case-insensitive match against the actual directory entries before giving up
- [ ] A referenced path that already matches on-disk casing exactly still resolves via the existing direct-open path (no behavior change, no unnecessary directory listing)
- [ ] A referenced file that genuinely doesn't exist under any casing still returns the existing descriptive "no such file or directory"-style error
- [ ] `~/workspace/ikemen-quick-versus/chars/King of Fighters/Kyo (XIII)/Kyo-XIII.def` (references `Kyoxiii.air`; the actual file is `KyoXIII.air`) loads successfully after the fix
- [ ] `~/workspace/ikemen-quick-versus/chars/One Piece/Luffy/Luffy.def` (references `Luffy.air`; the actual file is `LUFFY.AIR` — case differs on both the name and the extension) loads successfully after the fix

## Notes
Third real fixture: `~/workspace/ikemen-quick-versus/chars/Arcana Heart/Catherine/Catherine.def` (references `Catherine.air`; the actual file is `catherine.air`). Same area of `load.go` as item 049 (backslash-separator resolution) — both change how `loadAnimations`/`loadSprites`/`loadStateDefs` resolve a referenced path to a real file, likely worth implementing together in one pass over that code.
