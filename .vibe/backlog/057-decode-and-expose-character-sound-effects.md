---
status: blocked
---
# Decode And Expose Character Sound Effects

## Description
`Character.SoundFile` (and `CharacterInfo.SoundFile`) is currently a referenced path string only — the `.snd` file itself is never opened, decoded, or exposed. Once the org's new shared `snd` library exists, wire it in the same way `sff` already is: resolve `SoundFile` against the character's folder, decode it via `snd`, and expose the result through both the native Go API and the WASM `load`/`LoadBytes` JSON contract.

## Acceptance Criteria
- [ ] `Load`/`LoadBytes` resolve and decode a character's `SoundFile` via the `snd` library, the same way sprites are already resolved via `sff`
- [ ] A missing or unreadable sound file surfaces a descriptive error naming the file and step, matching this library's existing error convention for missing/unreadable referenced files
- [ ] Decoded sound groups/samples are exposed on `Character` and through the WASM JSON contract, keyed the same way `PlaySnd`'s `(group, sample)` parameters address them
- [ ] A character with no `SoundFile` (or an unreadable ambient `.snd`) still loads successfully for every other file kind, matching this library's existing "one missing optional piece doesn't fail the whole load" behavior where applicable

## Notes
Cross-repo: blocked on `snd` publishing a decode API. `snd` was created 2026-09-22 (roadmap `.vibe/decisions/026`) and now has its own backlog (`001`/`002` — v1/v2 read + PCM decode) but neither is implemented yet; this item only needs the native Go API those provide, not `snd`'s own WASM build (`003`). Re-check once `snd#001`/`002` ship. Feeds `character-editor#017` (sound browser) and, indirectly through `engine#020`'s triggered events, `mode-quick-versus#013` (match audio playback).
