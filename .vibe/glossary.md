# Ubiquitous Language

## Character
The in-memory representation of a MUGEN/Ikemen GO fighting-game character, combining its definition, sprites, animations, and combat logic. It is the top-level unit a library consumer (editor, engine) works with, rather than raw per-format structs (`.def`/`.sff`/`.air`/`.cns`).
_Sources: `character.go`_

## Animation
An ordered sequence of Frames (a `.air` file's `[Begin Action N]` block) plus the point at which playback loops back once it has played through once.
**Do not confuse with:** Frame, which is a single step of an Animation.
_Sources: `air/animation.go`_

## Frame
A single displayed image within an Animation: which sprite to show, where to show it, how long to hold it, how to mirror/blend it, and the collision boxes active while it is displayed.
_Sources: `air/animation.go`_

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

## Palette bank
A named (group, number) collection of colors a `.sff` v2 sprite can be drawn with, stored in the file's own palette table separately from the sprite table. Like a Linked sprite, a palette bank can link to (reuse) another bank's already-stored color data instead of storing its own, identified by an index. A Sprite's Palette reference identifies which palette bank it uses.
**Do not confuse with:** Sprite, which is drawn using a palette bank's colors but is not itself one.
_Sources: `sff/v2.go`_
