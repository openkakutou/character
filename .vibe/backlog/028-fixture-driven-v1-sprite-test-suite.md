---
status: todo
depends_on: [023, 026, 027]
---
# Fixture-Driven v1 Sprite Test Suite

## Description
Using the real fixtures vendored in item 023, port each scenario from `extract-v1.test.mjs` (11 tests) as a Go test: decode the sprite via `ParseV1`/`DecodePCX`/item 026's resolver, decode the corresponding expected `sprites/v1-sprite-0NN.png` via `image/png`, and compare decoded pixel buffers (not raw file bytes, since Go's PNG encoder won't byte-match the reference project's `pngjs` output). Covers: version string, basic sprite, last-group sprite, zero-length (copy previous) sprite, external-palette override, shared/own-palette consistency across multiple files, and an invalid-size sprite.

## Acceptance Criteria
- [ ] All 11 ported v1 scenarios pass against the real fixtures and produce pixel-identical output to the reference PNGs
- [ ] The "index == linked index" (`greenarrow-v1.sff`) and "linked index > current index" (`cvssakura-v1.sff`) scenarios are resolved by testing against the real fixtures directly, not guessed upfront (see Notes)
- [ ] Version-string extraction (`ParseV1` header) is asserted for at least one fixture

## Notes
Two scenarios exercise a v1 linking rule the reference project applies that this package's `loadV1`/`resolveV1Pixels` (`sff/load.go`) does not currently replicate: the reference project zeroes `linkedSpriteIndex` whenever it is `>= index` (self- or forward-references become "no link", and the sprite instead literally copies the *previous table entry*), whereas this package follows the declared `LinkedIndex` chain regardless of direction. Determine the correct behavior empirically against the downloaded fixtures; adjust the resolution helper added in item 026 rather than `ParseV1`'s existing offset/length parsing (which is correct either way). If a real behavior change to `loadV1` turns out to be needed, record it as a new ADR under `.vibe/decisions/`.
