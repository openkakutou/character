---
status: done
---
# Define CharacterInfo Data Model

## Description
Define the pure-data model for the `def` package: `CharacterInfo`, capturing character name, author, and the file paths it references (sprite, animation, sound, and other sections of a `.def` file). No parsing yet — establishes the read-path vocabulary for item 016.

## Acceptance Criteria
- [ ] `CharacterInfo` struct is exported with documented fields for name, author, and referenced file paths per `.def` section
- [ ] Zero-value `CharacterInfo{}` compiles and behaves predictably, covered by a zero-value test
- [ ] No parsing or file I/O present in this package yet

## Notes
No dependency on `air`/`sff` work — can proceed independently, though item 018 wires it to them.
