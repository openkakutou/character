# Data models

## Character
| Field | Type | Notes |
|---|---|---|
| Name | string | Placeholder field only — sprites, animations, and hitboxes will be added once the `def`/`sff`/`air`/`cns` sub-packages are implemented |

Defined in: `character.go`

## Animation
| Field | Type | Notes |
|---|---|---|
| Number | int | Corresponds to the `.air` file's `[Begin Action N]` action number |
| Frames | []Frame | Ordered sequence of displayed frames |
| LoopStart | int | Index into Frames the animation loops back to; zero value (0) matches `.air`'s own default of looping to the first frame when no `Loopstart` marker is present |

Defined in: `air/animation.go`

## Frame
| Field | Type | Notes |
|---|---|---|
| Group | int | Sprite group index |
| Image | int | Sprite image index within the group |
| X | int | Horizontal offset |
| Y | int | Vertical offset |
| Time | int | Display duration in game ticks |
| Flip | Flip | Mirroring axis (none, horizontal, vertical, both) |
| Blend | BlendMode | Blending mode token (e.g. additive, subtractive); zero value means normal blending |
| Clsn1 | []ClsnBox | Attack collision boxes active on this frame, already resolved from any `Clsn1Default` |
| Clsn2 | []ClsnBox | Vulnerability collision boxes active on this frame, already resolved from any `Clsn2Default` |

Defined in: `air/animation.go`

## Document
| Field | Type | Notes |
|---|---|---|
| Animations | []Animation | Decoded the same way `Parse`'s return value is |
| source | []byte | Unexported; the exact bytes `ParseDocument` read, replayed verbatim by `Serialize` |

Write-path type: `ParseDocument`/`Document.Serialize` round-trip an unmodified `.air` file byte-for-byte, including comments. Mutating `Animations` does not change what `Serialize` writes.

Defined in: `air/document.go`

## ClsnBox
| Field | Type | Notes |
|---|---|---|
| Left | int | |
| Top | int | |
| Right | int | |
| Bottom | int | |

Defined in: `air/animation.go`

## Sprite
| Field | Type | Notes |
|---|---|---|
| Group | int | Sprite group index |
| Image | int | Image index within Group |
| Width | int | Pixel width |
| Height | int | Pixel height |
| AxisX | int | Horizontal offset from top-left corner to the sprite's axis (pivot) point |
| AxisY | int | Vertical offset from top-left corner to the sprite's axis (pivot) point |
| Palette | int | Palette reference; exact meaning defined by the `.sff` version that populates it |

Defined in: `sff/sprite.go`

## SpriteGroup
| Field | Type | Notes |
|---|---|---|
| Index | int | Group index shared by every Sprite in Sprites |
| Sprites | []Sprite | Ordered collection of sprites belonging to this group; not itself validated against each Sprite's own Group field |

Defined in: `sff/sprite.go`

## V1Header
| Field | Type | Notes |
|---|---|---|
| Version | [4]byte | Raw version bytes as stored in the file (verhi, verlo1, verlo2, verlo3) |
| GroupCount | int | Number of sprite groups declared in the header |
| ImageCount | int | Total number of sprites declared in the header |
| SharedPalette | bool | True when sprites share one palette table (SPRPALTYPE_SHARED); false when each sprite carries its own (SPRPALTYPE_INDIV) |

Defined in: `sff/v1.go`

## V1SpriteEntry
| Field | Type | Notes |
|---|---|---|
| Group | int | Sprite group index |
| Image | int | Image index within Group |
| AxisX | int | Horizontal offset from top-left corner to the axis (pivot) point |
| AxisY | int | Vertical offset from top-left corner to the axis (pivot) point |
| Offset | int64 | Absolute file offset of this sprite's pixel data, immediately after its subheader |
| Length | int | Pixel data length in bytes; 0 means this sprite links to another sprite's data (see LinkedIndex) |
| LinkedIndex | int | Index, within the owning V1SpriteTable's Sprites, of the sprite this one shares pixel data with; only meaningful when Length is 0 |
| SharedPalette | bool | True when this sprite reuses the previous sprite's palette |

Defined in: `sff/v1.go`

## PCXImage
| Field | Type | Notes |
|---|---|---|
| Width | int | Pixel width, recovered from the PCX data's own embedded header |
| Height | int | Pixel height, recovered from the PCX data's own embedded header |
| Pixels | []byte | Row-major buffer of palette index values, length Width*Height; no RGB/palette resolution performed |

Defined in: `sff/pcx.go`

## V1SpriteTable
| Field | Type | Notes |
|---|---|---|
| Header | V1Header | The parsed file header |
| Sprites | []V1SpriteEntry | Sprite index table entries, in file order |

Write/read helper: `Offset(group, image int) (int64, bool)` resolves a `(group, image)` pair to its pixel data's file offset.

Defined in: `sff/v1.go`
