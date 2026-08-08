---
date: 2026-08-08
status: accepted
---
# v1 sprite pixel-linking and palette-inheritance rules corrected to match the reference decoder

**Context:** Item 028 ports the reference project's (`ikemen-launcher/sff-extractor`) v1 fixture-driven
test scenarios as Go tests, comparing decoded pixel output against known-good reference PNGs. Two of
those scenarios ("index == linked index" and "linked index > current index") exist specifically to
exercise an edge case in how a v1 sprite's `LinkedIndex` field is resolved, and three more ("same
palette" x3) exercise how a shared-palette sprite's actual palette is resolved. Porting the reference
project's own sprite-extraction source (`extractSpritesFromSFFV1.mjs`) line by line surfaced two
concrete differences from `resolveV1Pixels`/`ResolveV1Palette`'s prior behavior:

1. **Zero-length ("linked") sprite pixel resolution.** The prior implementation followed a sprite's own
   `LinkedIndex` field to find its pixel-data owner. The reference decoder never does this for a
   zero-length sprite: it unconditionally copies the immediately preceding table entry's *already-resolved*
   image (buffer, width, height) — `LinkedIndex` is stored but not consulted for this case.
2. **Non-zero-length sprite pixel resolution, and palette resolution for a shared-palette sprite.** The
   reference decoder clamps `LinkedIndex` to `0` whenever it is `>=` the sprite's own table index
   (self- or forward-references), rather than treating it as a genuine link — a raw uint16 value equal to
   or exceeding the current index is a "no real link" convention in this format, not a special case to
   preserve verbatim. Separately, palette resolution for a sprite marked "shares the previous palette" is
   not simply "nearest earlier sprite that owns a palette" (what a plain backward scan over the
   `SharedPalette` bit would produce): a sprite that is itself `(Group 0, Image 0)` and shares its palette
   instead inherits table index 0's own resolved palette specifically, even if a more recent owner exists
   in between. Table index 0 itself always resolves its own embedded palette regardless of its own
   `SharedPalette` bit, since no prior sprite exists for it to inherit from.

**Decision:** `resolveV1Pixels` (`sff/load.go`) and `ResolveV1Palette` (`sff/palette.go`) are corrected to
replicate both rules exactly, verified against the real fixtures in `sff/testdata` and the original,
untrimmed source character files the fixtures were trimmed from (see item 028's test suite). `loadV1`,
which already calls `resolveV1Pixels` internally for a sprite's width/height, inherits the corrected
behavior automatically — this is a bugfix, not a new code path, so no separate `Character`-level change was
needed. An oversized/corrupt PCX dimension guard (falling back to a synthetic 1x1 image instead of
attempting a pixel buffer allocation sized from untrusted header bytes) was added alongside these fixes,
matching the reference decoder's own guard for the "invalid sprite size" scenario.

**Reason:** This package's stated goal for pixel/color decoding is byte-for-byte behavioral parity with the
reference decoders it tracks (see ADR 014, ADR 015, ADR 016) — not just "a plausible reading of the raw
format spec". Real MUGEN/Ikemen character files exercise both quirks, and the fixture-driven tests exist
specifically to catch a decoder that only handles the common case.

**Rejected alternatives:**
- *Leave `resolveV1Pixels` following `LinkedIndex` for zero-length sprites, and only add the `>=` clamp for
  non-zero-length sprites.* Rejected — item 028's own real fixture (`v1-zero-length-copy.sff`) demonstrated
  the current chain-following approach does not reliably reach the same real pixel data as "copy the
  previous table entry" in a real file, since the two rules disagree whenever a sprite's `LinkedIndex`
  points somewhere other than `index - 1`.
- *Replicate palette inheritance via a plain backward scan for "nearest sprite with `SharedPalette ==
  false`", without the `(Group 0, Image 0)` special case.* Rejected — it is observably not what the
  reference decoder does, and the "same palette" scenarios in item 028 are exactly the ones designed to
  catch this difference.
