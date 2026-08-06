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
