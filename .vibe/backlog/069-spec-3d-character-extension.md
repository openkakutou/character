---
status: todo
---
# Specify the OpenKakutou-exclusive 3D character extension

## Description
Roadmap decision `028` (OpenKakutou-exclusive 3D characters) adds a 3D presentation layer to the existing character format. `character` is the reference for the character format, so it owns the written specification that `character-editor`, `character-viewer-web` and `game` implement against. Specification only: no glTF reading in this repo, the no-rendering-dependency constraint holds.

Cover:
- the `.def` `[Files]` `model` key (`.glb` path resolution, same rules as `sprite`/`anim`);
- the `.air` 3D frame syntax: a frame referencing `(animation clip, time)` instead of `(group, image)`, coexisting with sprite frames in the same action; choose a syntax that a MUGEN/Ikemen GO `.air` parser rejects cleanly or skips, never one it would misread as a valid sprite frame;
- the `OPENKAKUTOU_character` glTF vendor extension schema (model scale, ground origin relative to the character axis, root bone);
- palettes as `KHR_materials_variants` (palette slot → variant name mapping);
- facing = 180° rotation around the vertical axis, and why Clsn mirroring stays exact (orthographic projection on the fight plane);
- the standard action numbers a 3D character must provide as 3D poses (get-hit 5000+ range etc.) to interoperate with sprite characters' custom states.

🔒 MUST: every existing MUGEN 1.0/1.1 and Ikemen GO file keeps parsing and behaving exactly as before; the extension is strictly additive.

## Acceptance Criteria
- [ ] A spec document under `docs/` covers all points above, with one complete example (`.def` + `.air` excerpt + glTF extension JSON)
- [ ] It links roadmap decision `028`
- [ ] The `.air` syntax choice records how MUGEN and Ikemen GO parsers react to it, grounded in a source (engine source or doc)
