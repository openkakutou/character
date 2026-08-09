---
status: blocked
---
# Decode v2 RLE5 Pixel Format

## Description
`DecodeV2Sprite` (`sff/v2_decoder.go`) currently returns a descriptive "unsupported format" error for `V2FormatRLE5`, same as it does for any unrecognized format. A real, authoritative reference implementation exists to port from — `Rle5Decode` in the actual game engine, `ikemen-engine/Ikemen-GO` (`src/image.go`) — so this is technically portable the same way items 024/025 (RLE8/LZ5) were.

**Do not implement this until a real RLE5-encoded sprite is found and can be used as a fixture.** A scan of a large local real-character corpus (~562 real `.sff` v2 files, see `.vibe/fixture-sources.md`) found RLE8/LZ5/PNG8/24/32 all used thousands of times each, and RLE5 (format code 3) **zero** times. `ikemen-launcher/sff-extractor` itself has no RLE5 decoder either. The format appears to be effectively unused by real modern characters, so shipping a decoder for it now would be exercised only by a hand-built synthetic fixture — untestable against anything real, for a code path that may never actually run against real character data. This item stays open as a placeholder recording that decision, not as work to schedule.

## Acceptance Criteria
- [ ] Not implemented until a real RLE5-encoded sprite (a genuine character file using format code 3) is found — treat finding one as the trigger to start this item, not a pre-requisite for closing it
- [ ] If/when implemented: `DecodeV2Sprite(V2FormatRLE5, ...)` decodes RLE5-encoded data into the correct row-major index buffer, ported from `Rle5Decode` (`Ikemen-GO/src/image.go`), verified against the real fixture that justified doing the work
- [ ] If/when implemented: malformed/truncated RLE5 data returns a descriptive error instead of panicking or reading out of bounds, per this package's existing decoder convention
- [ ] Until then: `V2FormatRLE5` keeps returning its current "unsupported format" error, and that behavior stays covered by an explicit test (already the case via the shared "unrecognized format" test in `v2_decoder_test.go`)

## Notes
Source (for whenever this is picked up): `https://github.com/ikemen-engine/Ikemen-GO/blob/develop/src/image.go` (`Rle5Decode`, plus `Rle8Decode`/`Lz5Decode` as cross-validation that the file's algorithms match `sff-extractor`'s, and `readV2`'s format dispatch confirming format codes 2/3/4 and the 4-byte length-prefix convention).

## Blocked
2026-08-09: Ran `/vibe:feature 030 --auto`. Re-scanned the local real-character corpus (`~/workspace/ikemen-quick-versus/chars/`, 562 `.sff` files) for any v2 sprite entry with `Format == 3` (RLE5) using the existing `sff.ParseV2` reader — still **0** RLE5 sprites found, same result as when this item was written. The trigger condition ("a real RLE5-encoded sprite is found") has not occurred, so per this item's own acceptance criteria, implementation was not started. No code changes made. Re-run this item once a genuine RLE5-encoded character file turns up.
