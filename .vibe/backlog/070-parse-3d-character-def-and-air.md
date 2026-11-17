---
status: todo
depends_on: [069]
---
# Parse and write the 3D character `.def` key and `.air` 3D frames

## Description
Implement, as pure data, the `.def` `model` key and the `.air` 3D frames specified by item 069 (roadmap decision `028`). A 3D frame carries clip name, time, offset, duration, flip/blend fields and its Clsn boxes like a sprite frame. No glTF reading, no rendering.

🔒 MUST: existing MUGEN/Ikemen GO characters parse and save exactly as before (byte-exact unmodified save).

## Acceptance Criteria
- [ ] `model` is exposed and round-trips through save
- [ ] Sprite and 3D frames coexist in one action; both expose duration and Clsn the same way to consumers (`engine` needs no change)
- [ ] Unmodified save of a 3D character is byte-exact
- [ ] WASM exposes the new fields
- [ ] The existing real-world fixture suite still passes unchanged
