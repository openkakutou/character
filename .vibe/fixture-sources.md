# Fixture sources for `.sff`/`.air`/`.cns` test data

Reference material and real-world files used to source or validate test
fixtures for this repo's format packages. Kept separate from the codebase
index proper since none of this is referenced by path from Go code — some
of it is a local, machine-specific resource unavailable in CI or on other
machines.

## `ikemen-launcher/sff-extractor` (JS reference project)

`github.com/ikemen-launcher/sff-extractor` — the user's own JS project,
extracting/decoding `.sff` sprite data. Used as the primary behavioral
reference for `sff`'s v1/v2 pixel decoding and palette-resolution logic
(RLE8, LZ5, PNG8's forced index-0 transparency, external `.act` palette
handling including its reversed-byte-order quirk). Its own `tests/files/`
and `tests/sprites/` are the source for the trimmed real fixtures vendored
into `sff/testdata/` (see `.vibe/backlog/done/023-vendor-real-sff-test-fixtures.md`).

Has no RLE5 decoder (`decodeSpriteBuffer.mjs` throws `TODO RLE5`) — not a
reference for that format, see below.

## `ikemen-engine/Ikemen-GO` (the real game engine)

`github.com/ikemen-engine/Ikemen-GO`, `src/image.go` — the actual game
engine `.sff` files are built to run in, written in Go. Used to
cross-validate `sff-extractor`'s decode logic (`Rle8Decode`/`Lz5Decode`
match `decodeRLE8.mjs`/`decodeLZ5.mjs` algorithmically) and as the primary
reference where `sff-extractor` has no implementation to port from:
- `Rle5Decode` — the RLE5 pixel decoder (item 030); no equivalent exists in
  `sff-extractor`.
- `readActPalette` — confirms the external `.act` palette's reversed byte
  order and forced index-0 transparency (item 027) independently of
  `convertExternalPaletteToRGBA.mjs`.

## Local real-character corpus (not referenced from code)

`~/workspace/ikemen-quick-versus/chars/` on the machine this repo is
usually developed on: a real Ikemen GO frontend install with **562 real
character `.sff` files (~15GB)** across ~57 game franchises. Available
interactively for:
- Finding real fixtures for scenarios `sff-extractor`'s own bundled test
  files don't cover.
- Statistically validating format-code assumptions — e.g. a full scan
  found **zero** RLE5-encoded (format 3) sprites across all 562 files
  (55806 RLE8, 6163 LZ5, 52682+149+3154 PNG8/24/32), meaning RLE5 is
  apparently unused by real modern characters. See item 030's notes.

**This path must never be hardcoded into Go source, tests, or committed
config** — it only exists on this machine, not in CI or for other
contributors. Use it to *find and verify* candidate real fixtures, then
trim/vendor the result into `sff/testdata/` the same way item 023 did for
`sff-extractor`'s own files, exactly as if it had been sourced from any
other one-off, non-reproducible location.
