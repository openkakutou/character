---
status: done
depends_on: [023, 024, 025, 026, 027]
---
# Fixture-Driven v2 Sprite Test Suite

## Description
Using the real fixtures vendored in item 023, port each scenario from `extract-v2.test.mjs` (14 tests) as a Go test, comparing decoded pixel buffers against the corresponding expected `sprites/v2-sprite-0NN.png` (decoded via `image/png`, same pixel-buffer comparison approach as item 028). Covers: version string, RLE8 (item 024), LZ5 (item 025), PNG8/24/32, zero-length (copy first sprite) linkage, PNG8 forced-alpha, external palette (item 027), empty-palette-bank fallback (use first / use previous bank, item 026), and `loadMode = 1` ("on-demand" data section addressing).

## Acceptance Criteria
- [x] All 14 ported v2 scenarios pass against the real fixtures and produce pixel-identical output to the reference PNGs
- [x] The `loadMode = 1` (`kazuki-v2.sff`) scenario passes; if it requires reading `onDemandDataSizeTotal`-relative offsets that `ParseV2` (`sff/v2.go`) does not currently resolve, `ParseV2` is extended to support it
- [x] Version-string extraction (`ParseV2` header) is asserted for at least one fixture

## Notes
Reference: `extractSpritesFromSFFV2.mjs` in `ikemen-launcher/sff-extractor` for the on-demand addressing mode.

## Resolution
- Ported all 15 tests from the reference project's `extract-v2.test.mjs` (commit `2d4af64d26441bf4d692bb479275d64b11869678`, cloned and run directly to derive the authoritative scenario-to-fixture mapping — the file's own "14 tests" description undercounts by one) as `sff/v2_fixtures_test.go`.
- `ParseV2` needed no change for `loadMode = 1`: its existing offset-flag branch already computes the same absolute offset the reference project's `loadMode === 1` branch does, confirmed by decoding the real, untrimmed `kazuki-v2.sff` end to end with zero pixel mismatches.
- Real-file validation caught and fixed three pre-existing bugs, none anticipated by the original acceptance criteria: `ResolveSpritePixels` now resolves a v2 sprite with `Length == 0`/`LinkedIndex == 0` (copies the sprite table's own first entry, matching the reference project exactly — a `LinkedIndex > 0` sprite remains unsupported); `V2FormatPNG32` decode/encode reverted to straight (non-premultiplied) alpha, correcting a previous fix that had gone the wrong way, itself masked by a test-helper bug (`.At(x,y).RGBA()` always alpha-premultiplies, silently destroying color data at transparent pixels on both sides of the comparison); `ResolvePixels` gained a third `AlphaRule` (`AlphaOpaqueExceptIndexZero`) for PNG8, previously reusing PCX's rule incorrectly.
- See `.vibe/decisions/021-v2-zero-length-sprite-always-copies-table-index-zero.md` for the full design record.
