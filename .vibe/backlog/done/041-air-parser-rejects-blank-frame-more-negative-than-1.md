---
status: done
---
# Air Parser Rejects Blank-Frame Values More Negative Than -1

## Description
`air.Parse`'s `parseFrameLine` (in `air/parser.go`) hard-errors on any `Group`/`Image` value more negative than `-1` (e.g. `-2`), per the deliberate choice recorded in `.vibe/decisions/013-blank-frame-sentinel-representation.md`. In practice, real MUGEN/Ikemen engines treat any negative `Group` as "show no sprite for this frame" regardless of `Image`'s value — matching `Frame.IsBlank()`'s own broader `Group < 0 || Image < 0` check, which the parser's stricter validation contradicts. Reported via `character-viewer-web`: a real user-supplied `.air` file failed to load on the legitimate frame line `-1,-2, 0,0, 1`, with the error `image index must not be negative (except the -1 "no sprite" sentinel), got -2`.

## Acceptance Criteria
- [x] `air.Parse` accepts a frame line where `Group` and/or `Image` is any negative value (not just exactly `-1`) as the blank-frame sentinel, matching `Frame.IsBlank()`
- [x] A frame parsed from such a line resolves via `SpriteResolver`/`character.ResolveSprite` the same way an existing `-1,-1` blank frame does (no sprite lookup attempted)
- [x] Existing behavior for a genuinely malformed frame line (non-numeric fields, too few fields) is unchanged — still a descriptive, line-numbered error
- [x] `.vibe/decisions/013-blank-frame-sentinel-representation.md` is superseded by a new ADR recording the corrected rule (its `status` updated to `superseded by NNN`, per the ADR append-only convention)

## Notes
Root cause: `parseFrameLine` rejects `group < -1` / `image < -1` outright before `Frame.IsBlank()` ever gets a chance to recognize the value as a blank sentinel. The fix is presumably to relax those two range checks to accept any negative value, letting `IsBlank()` (already `< 0`) be the single source of truth for "blank" — no new field or type needed.
