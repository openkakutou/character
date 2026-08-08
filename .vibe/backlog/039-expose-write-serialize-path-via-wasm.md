---
status: todo
depends_on: [035]
---
# Expose Write/Serialize Path Via WASM (Def/Air/Cns/Cmd/Zss Round-Trip Save)

## Description
The WASM bridge currently only exposes read operations (`load`, `resolveSprites`). `character-editor` needs to save edited characters back to `.def`/`.air`/`.cns`/`.cmd`/`.zss`, which requires the existing Go `Serialize*` functions to be exposed through `cmd/wasm/main.go` the same way the read path already is.

## Acceptance Criteria
- [ ] WASM exports a save/serialize entrypoint accepting an edited in-memory character representation and returning serialized file bytes for each format
- [ ] Round-trip (load → no edits → save) produces byte-identical output to the original, matching the existing Go-level guarantee
- [ ] An invalid/incomplete in-memory character passed to the save entrypoint returns a descriptive error instead of a WASM panic
- [ ] Verified by a Node-based smoke test, same pattern as the existing `cmd/wasm/smoke.mjs`

## Notes
Depends on item 035 (sff migration) for sprite round-trip, since sprite pixel encode now lives in the external `sff` repo. Critical cross-repo dependency for `character-editor`'s "Save/Export Character Files" item.
