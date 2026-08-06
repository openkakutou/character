---
status: done
---
# Define Animation and Frame Data Model

## Description
Define the pure-data model for the `air` package: `Animation` (a sequence of `Frame`s plus loop metadata), `Frame` (sprite group, image index, x, y, duration, flip, blend mode), and `ClsnBox` (Clsn1/Clsn2 collision box coordinates). No parsing logic yet — this is the stable read-path surface a future `engine` consumer would import. Establishes the vocabulary the parser (item 002) will populate.

## Acceptance Criteria
- [x] `Animation`, `Frame`, and `ClsnBox` structs are exported from the `air` package with documented fields matching the `.air` format (group, image, x, y, time, flip, blend mode; Clsn box coordinates)
- [x] Zero-value `Animation{}` and `Frame{}` compile and behave predictably (no nil-pointer panics), covered by a zero-value test following the `character_test.go` pattern
- [x] No parsing, file I/O, or write-only logic present in this package yet

## Notes
Read path only, per CLAUDE.md's read/write separation constraint. This item has no dependency.
