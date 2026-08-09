---
date: 2026-08-10
status: accepted
---
# Blank-frame sentinel: any negative Group/Image, not just exactly -1

**Context:** `.vibe/decisions/013-blank-frame-sentinel-representation.md` let `air.Parse` accept `-1` on `Group`/`Image` as the "no sprite shown" sentinel, but deliberately kept rejecting anything more negative (e.g. `-2`) as malformed. A real community character (`Bardock`, reported via `character-viewer-web`) failed to load entirely on this: its `.air` file uses a whole sequence of blank-frame markers with a *varying* Image value alongside `Group = -1` — `-1,0` / `-1,-1` / `-1,-2` / `-1,-3` / `-1,-4` / `-1,-5` across consecutive frames of the same action. Real MUGEN/Ikemen engines treat any negative `Group` as "show no sprite for this frame" unconditionally, regardless of what `Image` holds — the `-2`/`-3`/... values are not corrupt data, they're just whatever the character's author happened to leave in that field. `Frame.IsBlank()` already encoded this broader rule (`Group < 0 || Image < 0`), but `parseFrameLine`'s own range check contradicted it by rejecting anything past `-1` before `IsBlank()` ever got a chance to apply.

**Decision:** Drop the `< -1` range check on both `Group` and `Image` in `parseFrameLine` entirely. Any negative integer on either field parses successfully, exactly mirroring `Frame.IsBlank()`'s existing `Group < 0 || Image < 0` rule — one source of truth for "blank" instead of two disagreeing ones. A non-numeric or missing field is still a parse error; only the *range* restriction is removed.

**Reason:**
- Matches real-world `.air` files and real engine behavior: the on-disk convention is "any negative `Group` means don't draw", not "the sentinel is exactly `-1,-1`".
- Removes a validation rule that was stricter than the data model's own contract (`IsBlank()`), which is a worse inconsistency than having no extra check at all — `013`'s attempt to catch "genuinely negative, therefore bad" data was based on an incomplete read of the convention.
- No new field or type: `IsBlank()` was already general enough; only the parser needed to stop second-guessing it.

**Rejected alternatives:**
- *Widen the accepted range to some other fixed floor (e.g. `>= -10`)* — rejected: no such floor exists in the MUGEN/Ikemen spec or in practice; an arbitrary cutoff would just move the bug to a different negative value.
- *Keep `013`'s strict range and special-case `Group == -1` only, ignoring `Image` range-checking when `Group == -1`* — rejected: more special-casing for the same outcome `IsBlank()` already expresses uniformly across both fields; the Bardock file also uses the same varying pattern conceptually symmetric on either field, per `013`'s own "Group or Image" framing.
