# Ubiquitous Language

## Character
The in-memory representation of a MUGEN/Ikemen GO fighting-game character, combining its definition, sprites, animations, and combat logic. It is the top-level unit a library consumer (editor, engine) works with, rather than raw per-format structs (`.def`/`.sff`/`.air`/`.cns`).
_Sources: `character.go`_

## Animation
An ordered sequence of Frames (a `.air` file's `[Begin Action N]` block) plus the point at which playback loops back once it has played through once.
**Do not confuse with:** Frame, which is a single step of an Animation.
_Sources: `air/animation.go`_

## Frame
A single displayed image within an Animation: which sprite to show, where to show it, how long to hold it, how to mirror/blend it, and the collision boxes active while it is displayed. A Frame may instead be a Blank frame, in which case it deliberately shows no sprite at all.
_Sources: `air/animation.go`_

## Blank frame
A Frame using the `.air` format's `-1` sentinel (most commonly `-1,-1`) in place of a real sprite group/image reference, meaning "show no sprite on this frame". A legitimate, engine-recognized authoring convention widely used by real characters — not a malformed frame — distinct from a Frame whose sprite reference is simply absent from the loaded sprites, which is a genuine error.
**Do not confuse with:** Frame, which normally references a real Sprite; a Blank frame explicitly does not.
_Sources: `air/animation.go`, `air/parser.go`, `air/resolve.go`, `.vibe/decisions/013-blank-frame-sentinel-representation.md`_

## Collision box
An axis-aligned box attached to a Frame that defines a region used for hit detection: an attack box (offense, "Clsn1") or a vulnerability box (can be hit, "Clsn2"). Represented per frame as the boxes actually active on that frame, already resolved from any file-level default.
_Sources: `air/animation.go`, `.vibe/decisions/001-frame-clsn-boxes-pre-resolved.md`_

## Sprite
A single image belonging to a character, identified by its group and image index, with a pixel width/height, an axis (pivot) point offset used when positioning it, and a reference to the palette it is drawn with. A Frame's `Group`/`Image` fields identify the Sprite it displays.
**Do not confuse with:** Frame, which is a step of an Animation that references a Sprite to display, not the sprite itself.
_Sources: `sff/sprite.go`_

## Sprite group
A collection of Sprites that share the same group index — e.g. the frames of a single stance or attack, addressed by their image index within the group.
_Sources: `sff/sprite.go`_

## Sprite index table
The part of a `.sff` file that lists every sprite by its `(group, image)` key and points to where that sprite's pixel data lives in the file, without describing the pixel data itself.
_Sources: `sff/v1.go`_

## Linked sprite
A sprite whose pixel data is not stored separately in the `.sff` file: it reuses ("links to") a previous sprite's already-stored pixel data, identified by an index, to avoid duplicating identical image data on disk.
**Do not confuse with:** Sprite, which always has its own metadata entry even when it links to another sprite's pixel data.
_Sources: `sff/v1.go`, `sff/v2.go`_

## Palette
The resolved set of 256 colors a Sprite's pixel indices are drawn with — the final on-screen color for each possible index byte. Getting a Sprite's Palette differs by `.sff` version: v1 embeds each non-shared sprite's own palette bytes directly in the file, right after its pixel data; v2 organizes colors into Palette banks, referenced by a Sprite's Palette index. Resolving a decoded Sprite's raw pixel indices against its Palette also applies one of two rules for how transparent index 0 is, depending on how the sprite was originally encoded. An External palette, once decoded, can be supplied in place of a Sprite's own Palette when resolving its colors.
**Do not confuse with:** Palette bank, which is v2's specific, linkable on-disk storage unit for a Palette's colors — v1 has no equivalent, storing a palette directly per sprite instead. Also not to be confused with External palette, which is a standalone recolor file rather than data owned by any particular Sprite.
_Sources: `sff/palette.go`, `.vibe/decisions/014-palette-resolution-api-shape.md`_

## External palette
A standalone `.act` palette file (256 RGB colors) used to recolor a character without touching its sprites — e.g. for alternate costumes/skins. Unlike a Sprite's own embedded Palette or a v2 Palette bank, it belongs to no particular Sprite or `.sff` file; its 256 colors are stored in reverse index order on disk, and only its resulting index 0 is treated as transparent.
**Do not confuse with:** Palette, which is a Sprite's own resolved colors — an External palette is a separate file a caller may substitute in its place.
_Sources: `sff/palette.go`, `.vibe/decisions/016-external-palette-override-api-shape.md`_

## Palette bank
A named (group, number) collection of colors a `.sff` v2 sprite can be drawn with, stored in the file's own palette table separately from the sprite table. Like a Linked sprite, a palette bank can link to (reuse) another bank's already-stored color data instead of storing its own, identified by an index. A Sprite's Palette reference identifies which palette bank it uses.
**Do not confuse with:** Sprite, which is drawn using a palette bank's colors but is not itself one.
_Sources: `sff/v2.go`_

## State
A named mode of a character's behavior (e.g. standing, an attack, a hit reaction), defined by a `.cns` `[Statedef N]` block: a state number, its type/move-type/physics classification, and the State controllers that run while it is active.
**Do not confuse with:** Animation, which is the visual sequence of Frames a state typically plays but is a separate `.air` concept referenced by number, not part of the state itself.
_Sources: `cns/statedef.go`, `cns/parser.go`_

## State controller
A single behavior a State can perform (e.g. changing velocity, dealing a hit, transitioning to another State), defined by a `.cns` `[State N]` block. Currently modeled as unevaluated data — its Triggers and Parameters are stored verbatim, not resolved against MUGEN/Ikemen's expression language.
**Do not confuse with:** State, which owns an ordered list of State controllers rather than being one itself.
_Sources: `cns/statedef.go`, `cns/parser.go`, `.vibe/decisions/011-cns-controller-parameters-are-untyped-key-value-data.md`_

## Trigger
A condition expression attached to a State controller (e.g. `"Time = 0"`) that determines whether the controller runs. A State controller with no triggers runs unconditionally whenever its State is active.
**Do not confuse with:** State controller, which owns triggers rather than being one.
_Sources: `cns/statedef.go`, `cns/parser.go`_
