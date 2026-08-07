---
date: 2026-08-08
status: accepted
---
# .sff v2 LZ5 pixel decoding: strict bounds/overrun errors instead of the reference decoder's silent clamping

**Context:** Item 025 ports LZ5 decoding (`decodeLZ5.mjs` from `ikemen-launcher/sff-extractor`, cross-validated against `Lz5Decode` in `ikemen-engine/Ikemen-GO`'s `src/image.go` — both implement the identical algorithm) into `DecodeV2Sprite`. Both reference implementations read past the end of malformed/truncated input by clamping their read cursor to the last valid byte instead of erroring (`if i < len(rle)-1 { i++ }`, otherwise re-reading the same last byte forever), and let a literal/back-reference run that would overrun the declared image silently stop writing once the output buffer is full, without signaling that anything was wrong.

**Decision:** The Go port (`decodeV2LZ5`) does not replicate this clamping. Any read past the end of the control stream, or any literal/back-reference run whose declared length would overrun the declared image size, returns a descriptive error instead of silently reusing stale bytes or truncating output. This matches the error-handling behavior `decodeV2RLE8` (item 024) already established for the sibling RLE8 format — the two formats now share the same "malformed input is a reported error, not best-effort output" contract, per the project's existing `.air`/`.cns` parsing convention of returning descriptive, position-aware errors on malformed input rather than guessing.

**Reason:** A game engine has good reason to be forgiving of a slightly malformed asset at runtime (better a corrupted-looking sprite than a crash mid-match), but this is a library whose read path is meant to be a trustworthy, stable data source for other tools (the editor, the future engine) — silently returning a partially-wrong pixel buffer for truncated/corrupt input would be a worse failure mode than a caller-visible error, and matches the RLE8 precedent that already made the same call for the same class of problem (compressed .sff v2 pixel data).

**Rejected alternatives:**
- Porting the clamping behavior verbatim for maximum fidelity to the reference decoders — rejected: on well-formed data the two behave identically (clamping only ever triggers on data that is already malformed), so there is no fidelity cost to real files, only a difference in how already-broken input is handled; and it would leave `DecodeV2Sprite`'s two compressed formats (RLE8, LZ5) with inconsistent error-handling contracts for no benefit.
