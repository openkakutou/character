---
status: in_progress
depends_on: [002]
---
# Handle .air Parsing Edge Cases and Errors

## Description
Harden the `.air` parser (item 002) against malformed or unusual input: malformed action headers, missing or non-numeric frame values, empty files, comment lines (`;`), and negative group/image indices. The parser must return descriptive errors rather than panicking or silently producing incorrect data.

## Acceptance Criteria
- [ ] Parsing an empty file returns a clear error (or an explicit empty result, per the chosen contract) rather than panicking
- [ ] A malformed action header or frame line returns an error identifying the offending line
- [ ] Comment lines (`;`) are ignored and do not break parsing of surrounding valid lines
- [ ] Negative group/image indices are rejected with a descriptive error

## Notes
Completes the read-path error contract for `.air` parsing before write-path work begins.
