---
status: todo
---
# Validate .zss against a broader Ikemen GO corpus

## Description
zss was validated on one character's 7 files (390 blocks). Ikemen GO ships many more constructs (nested functions, loops, helpers). Extend corpus validation to a wider set of real Ikemen GO characters and record unsupported constructs.

## Acceptance Criteria
- [ ] Corpus test covers at least 10 different real Ikemen GO characters
- [ ] Every block round-trips; unparsable blocks are listed in a documented known-gaps file
- [ ] Gaps that block engine execution are filed as engine backlog items
