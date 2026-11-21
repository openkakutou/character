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
- [ ] Clsn counts and indexes from the file cannot drive allocation: `Clsn1: 2000000000` with `Clsn1[1999999999]` (a 60-byte `.air`, `air/parser.go:332-369`) errors or is clamped (audit 2026-11-21, high)
- [ ] An unclosed `[` header in `.zss` does not concatenate the rest of the file quadratically (`zss/parser.go:142-152`, `zss/edit.go` indexBlocks)
- [ ] `trigger65536 = 1` does not pre-fill 65536 groups (`cns/parser.go:233-254`)
- [ ] `Copy Action` resolution has a bound on copied frames, or shares the frame slice (`air/parser.go:230-272`)
- [ ] The LCS alignment of the edit paths falls back to positional pairing above a size threshold (`cns/edit.go:245`, `air/edit.go:250`, `cmd/edit.go:253`, `zss/edit.go:105, 164`)
