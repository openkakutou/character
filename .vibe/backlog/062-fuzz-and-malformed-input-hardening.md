---
status: todo
---
# Fuzz parsers and harden against malformed input

## Description
Parsers face untrusted community files and the WASM build runs in browsers. Add Go native fuzz targets for def, air, cns, cmd and zss parsers to find panics, hangs and unbounded memory use; fix whatever is found with regression tests.

## Acceptance Criteria
- [ ] Fuzz targets exist for all five parsers and run in CI for a short budget
- [ ] Any crash/hang found is fixed with a committed regression seed
- [ ] Pathological inputs (huge repeats, deep nesting, giant numbers) return an error within a bounded time and memory
