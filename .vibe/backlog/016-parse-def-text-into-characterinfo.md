---
status: in_progress
depends_on: [015]
---
# Parse .def Text Into CharacterInfo

## Description
Implement a parser for the MUGEN/Ikemen `.def` INI-style text format (`[Section]` blocks with `key = value` pairs) into the `CharacterInfo` model (item 015).

## Acceptance Criteria
- [ ] Parsing a valid multi-section `.def` sample produces a `CharacterInfo` with correct name, author, and file references per section
- [ ] Unknown or unrecognized sections are preserved or reported without aborting the parse of known sections
- [ ] Malformed `key=value` lines return a descriptive error identifying the offending line

## Notes
Depends on 015 for the target model.
