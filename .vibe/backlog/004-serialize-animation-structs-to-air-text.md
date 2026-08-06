---
status: todo
depends_on: [001]
---
# Serialize Animation Structs to .air Text

## Description
Implement the write path for the `air` package: serialize `Animation`/`Frame` structs back into valid `.air` text syntax. This is a first pass without a byte-exact round-trip guarantee — it establishes correct, readable output from a freshly built `Animation`, not necessarily one parsed from an existing file.

## Acceptance Criteria
- [ ] Serializing a multi-action `Animation` produces valid `.air` text that re-parses (via item 002's parser) into an equivalent structure
- [ ] Clsn boxes and `Loopstart` are correctly emitted
- [ ] Output uses consistent, valid MUGEN/Ikemen `.air` syntax (verified by re-parsing it)

## Notes
Write-path logic must stay isolated from the read-path types, per CLAUDE.md's design constraints.
