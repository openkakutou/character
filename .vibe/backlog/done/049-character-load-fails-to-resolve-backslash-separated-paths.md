---
status: done
---
# Character Load Fails To Resolve Backslash-Separated Paths

## Description
`character.Load` (`load.go`) fails to resolve a `.def` `[Files]` entry that uses a Windows-style backslash path separator (e.g. `states\constants.cns`, `graphics\yoshi.air`), because `filepath.Join` treats the whole backslash-containing string as one literal filename component on a forward-slash filesystem (Linux, and the browser/WASM target this library must also support) — the referenced subdirectory is never traversed, and `os.Open` returns a plain `no such file or directory`. This is by far the single largest cause of file-not-found failures across the local real-character corpus (`~/workspace/ikemen-quick-versus/chars`): 54 of the 84 "no such file" failures across 717 real characters (~7.5% of the whole corpus).

## Acceptance Criteria
- [x] A `.def`-referenced file path containing backslash separators (e.g. `"states\constants.cns"`) resolves to the correct nested file on a forward-slash filesystem
- [x] A `.def`-referenced file path using forward slashes, or no separator at all, resolves exactly as today (no regression)
- [x] `~/workspace/ikemen-quick-versus/chars/Dragon Ball/Bardock/Bardock.def` (references `bardock_unotag\bardock.cns`) loads successfully after the fix
- [x] `~/workspace/ikemen-quick-versus/chars/Nintendo/Yoshi/Yoshi.def` (references `graphics\yoshi.air`) loads successfully after the fix
- [x] `~/workspace/ikemen-quick-versus/chars/DC Comics/Black Manta/Black Manta.def` (references `BLACK MANTA_unotag\BLACK MANTA.cns`) loads successfully after the fix

## Notes
Natural fix site: `load.go`'s `loadAnimations`/`loadSprites`/`loadStateDefs` (or a shared helper they all call), normalizing backslashes to `/` (or to `filepath.Separator`) on the referenced path string before joining it against the `.def`'s own directory — keeping `def.CharacterInfo`'s fields themselves as plain, unmodified strings (per CLAUDE.md's read-path "pure-data API" constraint) and doing the normalization only at the point file paths get resolved to the real filesystem. See also item 050 (case-insensitive resolution) — same area of `load.go`, likely worth implementing together since both affect the same file-resolution code path.
