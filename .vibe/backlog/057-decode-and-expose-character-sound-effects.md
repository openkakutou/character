---
status: todo
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
Cross-repo: was blocked on `snd` publishing a decode API; this item only needs the native Go API, not `snd`'s own WASM build (`003`). Feeds `character-editor#017` (sound browser) and, indirectly through `engine#020`'s triggered events, `mode-quick-versus#013` (match audio playback).

## Unblocked
2026-09-25: `snd#001`/`002` (v1/v2 read + PCM decode) both shipped and published — `snd` v0.2.0 (2026-09-24) covers the v1/v2 read+decode API this item needs; v0.3.0 (2026-09-24) additionally adds the WASM build (`snd#003`), not required here. Back to `status: todo`.
