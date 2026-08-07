---
status: done
---
# Vendor Real .sff/.act/.png Fixtures From sff-extractor

## Description
Download the real character `.sff`/`.act` files and their JS-generated expected-output PNGs from the user's public reference project (`github.com/ikemen-launcher/sff-extractor`, `tests/files/` and `tests/sprites/`), then trim each `.sff`/`.act` down to a minimal file containing only the sprite(s)/palette(s) a ported test scenario actually needs, using real bytes copied verbatim from the source file (no invented pixel/palette data). The full downloaded files total ~329MB (real characters carry every sprite/animation frame); committing that verbatim would permanently bloat this repo's git history, so only the trimmed, minimal-but-real fixtures are vendored into `sff/testdata/`. Expected-output PNGs (small, ~KB-sized) are vendored as downloaded, unmodified.

## Acceptance Criteria
- [ ] A trimming tool (kept out of the public package API — e.g. a throwaway `go run` script, not a new exported `sff` function) reads each real source file with `ParseV1`/`ParseV2`, locates the exact (group, image) entry a test scenario needs (per the JS test's own comments), and writes a minimal valid `.sff` containing that entry's real subheader/pixel/palette bytes — renumbering table positions as needed but preserving each entry's real `LinkedIndex`/palette-linkage *relationship* (self-reference, forward-reference, zero-length copy) so the quirk under test survives trimming
- [ ] `sff/testdata/files/` contains the resulting trimmed `.sff`/`.act` fixtures (KB-scale, not MB-scale) plus a smoke test asserting each one still parses via `ParseV1`/`ParseV2` and exposes the documented sprite at the expected (group, image)
- [ ] `sff/testdata/sprites/` contains the expected-output PNGs, vendored unmodified
- [ ] `sff/testdata/README.md` records the source repository/commit the originals were pulled from and notes that `.sff`/`.act` fixtures are trimmed (not byte-identical to upstream), while `sprites/*.png` are unmodified
- [ ] The full untrimmed downloads are not committed anywhere in the repo or its history

## Notes
Source: `https://raw.githubusercontent.com/ikemen-launcher/sff-extractor/main/tests/`. Two v2 scenarios need extra care during trimming: the `loadMode = 1` ("on-demand" data section) fixture for item 029, which this package's `ParseV2`/`SerializeV2` don't yet address (may need the trimming tool to hand-write that section, or `ParseV2` extended first — see item 029); and the zero-length "copy" scenarios, which need two real entries (source + copying entry) kept together with their linkage intact.
