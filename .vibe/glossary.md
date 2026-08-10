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
A Frame using the `.air` format's "no sprite" sentinel — any negative value (not just `-1`) in place of a real sprite group/image reference, meaning "show no sprite on this frame". A legitimate, engine-recognized authoring convention widely used by real characters — not a malformed frame — distinct from a Frame whose sprite reference is simply absent from the loaded sprites, which is a genuine error.
**Do not confuse with:** Frame, which normally references a real Sprite; a Blank frame explicitly does not.
_Sources: `air/animation.go`, `air/parser.go`, `air/resolve.go`, `.vibe/decisions/024-blank-frame-sentinel-accepts-any-negative-value.md`_

## Collision box
An axis-aligned box attached to a Frame that defines a region used for hit detection: an attack box (offense, "Clsn1") or a vulnerability box (can be hit, "Clsn2"). Represented per frame as the boxes actually active on that frame, already resolved from any file-level default.
_Sources: `air/animation.go`, `.vibe/decisions/001-frame-clsn-boxes-pre-resolved.md`_

## Sprite
A single image belonging to a character, identified by its group and image index, with a pixel width/height, an axis (pivot) point offset used when positioning it, and a reference to the palette it is drawn with. A Frame's `Group`/`Image` fields identify the Sprite it displays. Defined and decoded by the external `github.com/openkakutou/sff` module (item 035); this repo consumes it as `sff.Sprite`.
**Do not confuse with:** Frame, which is a step of an Animation that references a Sprite to display, not the sprite itself.
_Sources: `character.go`, `load.go`, `load_bytes.go`, `air/resolve.go`_

## Sprite group
A collection of Sprites that share the same group index — e.g. the frames of a single stance or attack, addressed by their image index within the group. Defined by the external `github.com/openkakutou/sff` module (item 035); this repo consumes it as `sff.SpriteGroup`.
_Sources: `character.go`, `load.go`, `load_bytes.go`, `air/resolve.go`_

## Palette
The resolved set of 256 colors a Sprite's pixel indices are drawn with — the final on-screen color for each possible index byte. An External palette, once decoded, can be supplied in place of a Sprite's own Palette when resolving its colors. Resolution logic (per-`.sff`-version palette lookup, alpha rules) is defined by the external `github.com/openkakutou/sff` module (item 035); this repo consumes it as `sff.Palette`, notably in `cmd/wasm`'s palette-override support.
**Do not confuse with:** External palette, which is a standalone recolor file rather than data owned by any particular Sprite.
_Sources: `cmd/wasm/main.go`, `.vibe/decisions/014-palette-resolution-api-shape.md`_

## External palette
A standalone `.act` palette file (256 RGB colors) used to recolor a character without touching its sprites — e.g. for alternate costumes/skins. Unlike a Sprite's own embedded Palette, it belongs to no particular Sprite or `.sff` file; its 256 colors are stored in reverse index order on disk, and only its resulting index 0 is treated as transparent. Decoded by the external `github.com/openkakutou/sff` module's `DecodeExternalPalette` (item 035); this repo calls it from `cmd/wasm` to support a palette-override request from a browser caller.
**Do not confuse with:** Palette, which is a Sprite's own resolved colors — an External palette is a separate file a caller may substitute in its place.
_Sources: `cmd/wasm/main.go`, `.vibe/decisions/016-external-palette-override-api-shape.md`_

## State
A named mode of a character's behavior (e.g. standing, an attack, a hit reaction), defined by a `.cns` `[Statedef N]` block: a state number, its type/move-type/physics classification, and the State controllers that run while it is active. A character may instead define its states via `.zss`'s own `[Statedef N; ...]` block shape (never both formats at once) — there, the same header fields are kept as unevaluated `key: value` data rather than typed classification fields, and the controllers that run are Lua-like script statements kept as one raw, unevaluated body rather than individually parsed.
**Do not confuse with:** Animation, which is the visual sequence of Frames a state typically plays but is a separate `.air` concept referenced by number, not part of the state itself; Function, `.zss`'s other block kind, which is a reusable named script routine rather than a mode of the character's own behavior.
_Sources: `cns/statedef.go`, `cns/parser.go`, `zss/script.go`, `zss/parser.go`_

## State controller
A single behavior a State can perform (e.g. changing velocity, dealing a hit, transitioning to another State), defined by a `.cns` `[State N]` block. Currently modeled as unevaluated data — its Triggers and Parameters are stored verbatim, not resolved against MUGEN/Ikemen's expression language.
**Do not confuse with:** State, which owns an ordered list of State controllers rather than being one itself.
_Sources: `cns/statedef.go`, `cns/parser.go`, `.vibe/decisions/011-cns-controller-parameters-are-untyped-key-value-data.md`_

## Trigger
A condition expression attached to a State controller (e.g. `"Time = 0"`) that determines whether the controller runs. A State controller with no triggers runs unconditionally whenever its State is active.
**Do not confuse with:** State controller, which owns triggers rather than being one.
_Sources: `cns/statedef.go`, `cns/parser.go`_

## Command
A named input sequence (a motion and/or button combination, e.g. `"~D, DF, F, a"`) a player can perform, defined by a `.cmd [Command]` section. A Command's name is referenced by a Trigger on a linked State controller (e.g. `command = "holdback"`) to make performing it change the character's State — the actual sequence text itself is stored verbatim, not evaluated against MUGEN/Ikemen's input-matching rules.
**Do not confuse with:** Trigger, which is the condition expression that reads a Command's name to decide whether a State controller runs; a Command is the input sequence being named, not the condition that reacts to it.
_Sources: `cmd/command.go`, `cmd/parser.go`_

## Function
A reusable, named block of Lua-like script logic, defined by a `.zss [Function Name(params) ret]` block — Ikemen GO's `.zss` state-script format's equivalent of a subroutine, callable from a State's own script body or from another Function. Its parameter names and declared return variable name(s) are parsed; its actual script logic is kept as raw, unevaluated text, the same way a State's script body is.
**Do not confuse with:** State, `.zss`'s other block kind, which defines a mode of the character's own behavior rather than a reusable routine.
_Sources: `zss/script.go`, `zss/parser.go`_
