---
status: done
---
# Wire .cmd parsing into Character

## Description
README states that wiring .cmd/.zss into Character is not implemented. The cmd package parses and round-trips .cmd files (item 036) but Character does not load or expose them. Load the .cmd file referenced by the .def, expose it on Character, include it in JSON/WASM output and in the Document write path, and surface a clear error when the referenced file is missing or malformed.

## Acceptance Criteria
- [ ] Loading a character also loads its referenced .cmd into a typed field
- [ ] A missing or unparseable .cmd yields a descriptive, non-fatal error
- [ ] The WASM build returns the command data to JS callers
- [ ] Saving an unmodified character reproduces its .cmd byte-for-byte
