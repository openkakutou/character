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

## ClsnBox
| Field | Type | Notes |
|---|---|---|
| Left | int | |
| Top | int | |
| Right | int | |
| Bottom | int | |

Defined in: `air/animation.go`
