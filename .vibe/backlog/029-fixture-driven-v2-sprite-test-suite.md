---
status: todo
depends_on: [023, 024, 025, 026, 027]
---
# Fixture-Driven v2 Sprite Test Suite

## Description
Using the real fixtures vendored in item 023, port each scenario from `extract-v2.test.mjs` (14 tests) as a Go test, comparing decoded pixel buffers against the corresponding expected `sprites/v2-sprite-0NN.png` (decoded via `image/png`, same pixel-buffer comparison approach as item 028). Covers: version string, RLE8 (item 024), LZ5 (item 025), PNG8/24/32, zero-length (copy first sprite) linkage, PNG8 forced-alpha, external palette (item 027), empty-palette-bank fallback (use first / use previous bank, item 026), and `loadMode = 1` ("on-demand" data section addressing).

## Acceptance Criteria
- [ ] All 14 ported v2 scenarios pass against the real fixtures and produce pixel-identical output to the reference PNGs
- [ ] The `loadMode = 1` (`kazuki-v2.sff`) scenario passes; if it requires reading `onDemandDataSizeTotal`-relative offsets that `ParseV2` (`sff/v2.go`) does not currently resolve, `ParseV2` is extended to support it
- [ ] Version-string extraction (`ParseV2` header) is asserted for at least one fixture

## Notes
Reference: `extractSpritesFromSFFV2.mjs` in `ikemen-launcher/sff-extractor` for the on-demand addressing mode.
