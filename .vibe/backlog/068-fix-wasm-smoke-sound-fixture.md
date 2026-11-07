---
status: todo
---
# Fix the WASM smoke test's broken sound fixture

## Description
`node cmd/wasm/smoke.mjs` fails on "sound group index is 2 (got: 589826)" and "sound group has 1 sound with sample 9", already before item 060. The JS `buildV1Snd` writes the subheader group/sample as 16-bit fields at offsets 8/10, while the Go helper `buildV1SndFile` (sound_test.go) and the decoder use 32-bit fields at 8/12. The decoder is right; the fixture is wrong. The smoke test is never run in CI, which is why it went unnoticed.

## Acceptance Criteria
- [ ] `buildV1Snd` writes 32-bit group/sample at offsets 8/12; smoke exits 0
- [ ] Smoke test runs in CI after the WASM build
