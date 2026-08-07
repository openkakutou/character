---
date: 2026-08-07
status: accepted
---
# Blank-frame sentinel is represented as -1 on Frame.Group/Image, not a separate field

**Context:** Real MUGEN/Ikemen `.air` files widely use `-1,-1` (or `-1` in only the group or only the image position) on a frame line to mean "show no sprite for this frame". `air.Parse` previously rejected any negative group/image index outright. Something needs to let `Frame` carry this "no sprite" state so `air.SpriteResolver`/`character.ResolveSprite` can recognize it directly instead of attempting a lookup that would fail.

**Decision:** Keep `Frame.Group`/`Frame.Image` as the sole carriers of this state: `-1` on either field is the recognized sentinel, matching the on-disk convention verbatim. Add a `Frame.IsBlank() bool` method (`Group < 0 || Image < 0`) to the read-path model as the single recognition point. `air.Parse` accepts `-1` as a valid value on either field (still rejecting anything more negative than `-1`, e.g. `-2`, as malformed). `SpriteResolver.Resolve` checks `frame.IsBlank()` first and returns a zero `sff.Sprite` with a `nil` error — a normal, intentional outcome, not a failure — without touching the sprite index.

**Reason:**
- Mirrors the file format's own convention directly: no translation step between "what the file says" and "what the struct holds", keeping `Parse` simple and the model easy to reason about against real `.air` files.
- A blank frame is semantically "there is deliberately no sprite here", not "the sprite reference is missing" — returning `(zero Sprite, nil)` from `Resolve` keeps that distinct from `.vibe/decisions/008-air-sprite-resolution-lives-in-air-package.md`'s existing contract that a genuinely missing `(Group, Image)` reference fails with a descriptive error. `IsBlank()` gives callers (and `Resolve` itself) one explicit place to tell the two apart, rather than inferring "blank" from a lookup miss.
- Extends the existing negative-value carve-out already present in `parseFrameLine` (X/Y/Time are already accepted negative) by exactly one more case, rather than introducing a new field that would duplicate information already expressible via Group/Image.

**Rejected alternatives:**
- *Add a dedicated `Frame.Blank bool` field, leaving Group/Image at some placeholder (e.g. 0).* Rejected: loses the on-disk distinction between "-1 in the group position" and "-1 in the image position" for no benefit, and creates a state where two fields (`Blank` and `Group`/`Image`) could disagree if constructed by hand outside `Parse`.
- *Have `Resolve` return a sentinel error (e.g. `ErrBlankFrame`) instead of `(zero Sprite, nil)`.* Rejected: a blank frame is not an error condition — forcing every caller to use `errors.Is` to reach the common "draw nothing" path is more ceremony than the case warrants, and `IsBlank()` already lets a caller check upfront if it wants to skip resolution entirely.
