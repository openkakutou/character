---
status: todo
depends_on: [069]
---
# Sample 3D character fixture: an improved Quaternius character

## Description
No 3D character exists yet to validate item 070 and the editor/viewer/mode work of roadmap decision `028`. Build one from a **Quaternius** base (CC0, glTF): a rigged character from the Universal Animation Library 1/2 and the clips it ships (locomotion, melee combos — verify each clip used is in the free tier, via the Quaternius animation viewer), **improved** with the fighting-game moves Quaternius lacks, authored in Blender on the same rig. Mixamo is excluded: its terms forbid redistributing raw files, which a public fixture would do.

Moves to add or rework (non-exhaustive, adjust after checking what Quaternius already provides):
- fighting stance idle (instead of a neutral idle), crouch, crouch idle;
- at least one kick and one crouching attack;
- the standard get-hit / fall / lie-down / get-up actions `common1.cns` needs (5000+ range), so custom states of sprite characters can drive this one.

The character reuses `common1.cns` plus a few states of its own. Its `.air` uses 3D frames for the body; Clsn boxes are hand-written for now (`character-editor` item 028 will bake them later).

Reproducibility: the Blender source (`.blend`) and a headless export script (`blender -b … --python …`) are committed with the fixture, so the `.glb` can be regenerated. The `OPENKAKUTOU_character` extension and `KHR_materials_variants` palettes are added by that script or a post-export step, whichever the Blender glTF exporter allows. Blender is not currently installed in the dev environment.

The whole fixture (Quaternius base + our additions) is released **CC0**, so it stays single-licensed and `character-editor`, `character-viewer-web` and `mode-quick-versus` can vendor it freely for their own tests (pinned to a `character` version, as for WASM per roadmap decisions `016`/`025`).

## Acceptance Criteria
- [ ] Fixture committed under a `testdata/` folder, with `CREDITS` naming each Quaternius pack and its CC0 license, and stating our additions are CC0
- [ ] `.blend` source and export script committed; running the script regenerates the `.glb`
- [ ] `.glb` declares `OPENKAKUTOU_character` and at least two `KHR_materials_variants` (palettes)
- [ ] Every action listed above exists, including the `common1.cns` get-hit set
- [ ] Parses with item 070's parser and is used by its tests
- [ ] Fixture size documented and kept small (unused clips stripped)
