---
status: todo
---
# Vendor Real .sff/.act/.png Fixtures From sff-extractor

## Description
Download a subset of the real character `.sff`/`.act` files and their JS-generated expected-output PNGs from the user's public reference project (`github.com/ikemen-launcher/sff-extractor`, `tests/files/` and `tests/sprites/`) into a new `sff/testdata/` directory in this repo. These real files are the ground truth every fixture-driven test added in later items (028/029) compares against — no synthetic/invented data.

## Acceptance Criteria
- [ ] `sff/testdata/files/` contains every `.sff`/`.act` fixture referenced by the ported v1/v2 test scenarios (see items 028/029)
- [ ] `sff/testdata/sprites/` contains every expected-output PNG those scenarios compare against
- [ ] `sff/testdata/README.md` records the source repository and commit/ref the files were pulled from
- [ ] No production code changes in this item — fixture acquisition only

## Notes
Source: `https://raw.githubusercontent.com/ikemen-launcher/sff-extractor/main/tests/`. These are real binary character-file assets the user's own public repo already vendors under MIT.
