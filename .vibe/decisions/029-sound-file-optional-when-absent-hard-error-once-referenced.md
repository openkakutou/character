---
date: 2026-09-25
status: accepted
---
# SoundFile Is Optional When Absent, a Hard Error Once Referenced

**Context:** Backlog item 057 wires `Character.SoundFile` to the new `github.com/openkakutou/snd` library, the same way `SpriteFile`/`AnimationFile`/`ConstantsFile` are already wired to `sff`/`air`/`cns`. Those three are always required once a `.def` has a `[Files]` section at all (`validateFilesSection` never includes `SoundFile` in that check) — a missing or unreadable one already hard-fails `Load`/`LoadBytes`. The backlog item's acceptance criteria pull in two directions for `SoundFile`: one criterion asks for the same descriptive-error convention on a missing/unreadable file, another asks for "one missing optional piece doesn't fail the whole load" behavior.

**Decision:** `SoundFile` is treated as optional only in the sense it already was in `validateFilesSection`: an empty `SoundFile` (the common case — many real characters have no dedicated `.snd`, or share one that isn't always present in every distribution) is not an error at all — `Character.Sounds` is simply empty, and `Load`/`LoadBytes` proceed normally. Once `SoundFile` is non-empty, it is required exactly like the other three referenced files: a missing or undecodable `.snd` at that path is a hard error from `Load`/`LoadBytes`, in the same `"character: <step> %q: %w"` shape `loadAnimations`/`loadSprites`/`loadStateDefs` already use.

`LoadBytes` gains a fifth parameter, `sndBytes []byte` — nil or empty means "no sound data supplied", mirroring an empty `SoundFile`, not an error. The WASM `load` global's new fifth argument is optional at the JS call boundary (4 arguments still works, meaning "no sound data") rather than bumping the required count to 5, so existing WASM callers keep working unchanged until they choose to pass sound bytes too.

**Reason:** This keeps the new behavior consistent with the one place this repo already drew this exact line (`validateFilesSection`'s existing three-required-fields check already excludes `SoundFile`), rather than inventing a new "silently swallow a real error" contract that would conflict with this library's existing descriptive-error convention and its robustness review agent. It also avoids a breaking change to the WASM contract's required argument count for a genuinely additive feature.

**Rejected alternatives:** Making any sound-decode failure (including a present-but-broken file) non-fatal, surfaced instead through a new `Character.SoundLoadError`-style field — rejected as unnecessary new API surface, and inconsistent with how every other referenced file already fails loudly once declared. Requiring `LoadBytes`/WASM `load` callers to always pass sound bytes (bumping the required argument count) — rejected as an avoidable breaking change when the empty case is a legitimate, common input.
