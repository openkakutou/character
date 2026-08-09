---
status: done
---
# Decode v2 RLE5 Pixel Format

## Description
`DecodeV2Sprite` (`sff/v2_decoder.go`) currently returns a descriptive "unsupported format" error for `V2FormatRLE5`, same as it does for any unrecognized format. A real, authoritative reference implementation exists to port from — `Rle5Decode` in the actual game engine, `ikemen-engine/Ikemen-GO` (`src/image.go`) — so this is technically portable the same way items 024/025 (RLE8/LZ5) were.

**Do not implement this until a real RLE5-encoded sprite is found and can be used as a fixture.** A scan of a large local real-character corpus (~562 real `.sff` v2 files, see `.vibe/fixture-sources.md`) found RLE8/LZ5/PNG8/24/32 all used thousands of times each, and RLE5 (format code 3) **zero** times. `ikemen-launcher/sff-extractor` itself has no RLE5 decoder either. The format appears to be effectively unused by real modern characters, so shipping a decoder for it now would be exercised only by a hand-built synthetic fixture — untestable against anything real, for a code path that may never actually run against real character data. This item stays open as a placeholder recording that decision, not as work to schedule.

## Acceptance Criteria
- [ ] ~~Not implemented until a real RLE5-encoded sprite (a genuine character file using format code 3) is found~~ — superseded: the Product Owner explicitly asked to override this and proceed anyway (see Resolved)
- [x] `DecodeV2Sprite(V2FormatRLE5, ...)` decodes RLE5-encoded data into the correct row-major index buffer, ported from `Rle5Decode` (`Ikemen-GO/src/image.go`) — implemented in the external `github.com/openkakutou/sff` module (`sff`'s own backlog item 006), not in this repo, since this repo now fully delegates `.sff` handling to that dependency (item 035)
- [x] Malformed/truncated RLE5 data returns a descriptive error instead of panicking or reading out of bounds — same delegation, inherited from `sff`
- [x] `character` picks it up automatically via the `github.com/openkakutou/sff` dependency bump to v0.2.0 — no `character`-specific code needed, no `character`-specific test added (nothing character-specific to cover; `sff`'s own test suite is the coverage)

## Notes
Source (for whenever this is picked up): `https://github.com/ikemen-engine/Ikemen-GO/blob/develop/src/image.go` (`Rle5Decode`, plus `Rle8Decode`/`Lz5Decode` as cross-validation that the file's algorithms match `sff-extractor`'s, and `readV2`'s format dispatch confirming format codes 2/3/4 and the 4-byte length-prefix convention).

## Resolved
2026-08-09: The Product Owner explicitly asked to unblock this item and implement RLE5 anyway, without waiting for the real fixture this item was deferring on. Since this repo migrated its sprite handling to the external `github.com/openkakutou/sff` module (item 035), the actual implementation happened there — `sff`'s own backlog item 006, with the same trade-off explicitly accepted and recorded in `sff`'s decision `014-v2-rle5-decode-and-encode-implemented-without-a-real-fixture.md`. That work was committed, released as `sff` v0.2.0, and this repo's `go.mod` bumped to depend on it (`go get github.com/openkakutou/sff@v0.2.0 && go mod tidy`) — `go build`/`go test ./...`/`gofmt -l .`/`go vet ./...` all clean afterward. No code changes were needed in this repo itself; `DecodeV2Sprite`/`ResolveSpritePixels` calls now resolve RLE5 sprites the same as every other format, purely through the updated dependency.

Same validation gap as upstream: RLE5 has never been validated against a real on-disk `.sff` file in either repo — see `sff`'s decision `014` for the accepted trade-off. If a genuine RLE5-encoded character file ever turns up, it should be added as a fixture in the `sff` repo (not here), per that decision's own follow-up note.

Previous blocked-status history:
2026-08-09: Ran `/vibe:feature 030 --auto`. Re-scanned the local real-character corpus (`~/workspace/ikemen-quick-versus/chars/`, 562 `.sff` files) for any v2 sprite entry with `Format == 3` (RLE5) using the existing `sff.ParseV2` reader — still **0** RLE5 sprites found, same result as when this item was written. The trigger condition ("a real RLE5-encoded sprite is found") has not occurred, so per this item's own acceptance criteria, implementation was not started. No code changes made.
