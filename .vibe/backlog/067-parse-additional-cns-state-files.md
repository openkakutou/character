---
status: todo
---
# Parse the additional .cns state files of a character

## Description
`StateFiles` (`st`, `stN` entries of the .def) are listed but a `.cns`-format one is never read: only `cns` reaches `StateDefs`. Item 060 loads the `.zss` ones; load the `.cns` ones too, in Ikemen GO's order (st, stN, then stcommon when in the character folder), non-fatal per file like ADR 032.

## Acceptance Criteria
- [ ] Every `.cns` state file referenced by `st`/`stN` is parsed and exposed
- [ ] A missing or malformed one is reported per file without failing the load
- [ ] WASM exposes them; unmodified save byte-exact
