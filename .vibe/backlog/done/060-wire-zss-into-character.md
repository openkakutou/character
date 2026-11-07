---
status: done
---
# Wire .zss scripts into Character

## Description
Same gap as the .cmd wiring, for Ikemen GO state scripts: the zss package parses .zss (item 037) but Character neither loads nor exposes it. Load every .zss file the .def references (alongside or instead of .cns), expose them, and keep the raw bodies available for engine.

## Acceptance Criteria
- [ ] Loading a character loads the .zss files referenced by its .def
- [ ] A character using .zss instead of .cns loads without error
- [ ] WASM exposes the zss blocks
- [ ] Unmodified save is byte-exact for .zss
