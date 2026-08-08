# Data models

## CharacterInfo
| Field | Type | Notes |
|---|---|---|
| Name | string | Character's display name (.def `[Info]` "name") |
| Author | string | Character's author/creator (.def `[Info]` "author") |
| SpriteFile | string | Path to the .sff sprite sheet (.def `[Files]` "sprite") |
| AnimationFile | string | Path to the .air animation file (.def `[Files]` "anim") |
| SoundFile | string | Path to the .snd sound file (.def `[Files]` "sound") |
| CommandFile | string | Path to the .cmd command input file (.def `[Files]` "cmd") |
| ConstantsFile | string | Path to the main .cns combat logic file (.def `[Files]` "cns") |
| StateFiles | []string | Additional .st files beyond ConstantsFile, in .def file order |
| Palettes | []string | .act palette files, in palette number order (.def `[Files]` "pal1", "pal2", ...) |

Paths are stored exactly as written in the .def file; resolving them against a filesystem is left to the package that eventually loads a .def file. Populated by `Parse(r io.Reader) (CharacterInfo, error)`, which reads `[Info]`/`[Files]` `.def` text into this shape, skipping unrecognized sections/keys. Written back out by `Serialize(w io.Writer, info CharacterInfo) error`.

Defined in: `def/character_info.go`, `def/parser.go`, `def/serializer.go`

## Document (def)
| Field | Type | Notes |
|---|---|---|
| Info | CharacterInfo | Decoded the same way `Parse`'s return value is |
| source | []byte | Unexported; the exact bytes `ParseDocument` read, replayed verbatim by `Serialize` |

Write-path type: `ParseDocument`/`Document.Serialize` round-trip an unmodified `.def` file byte-for-byte, including comments, section ordering, and unrecognized sections. Mutating `Info` does not change what `Serialize` writes. Mirrors `air`'s own `Document` (see below).

Defined in: `def/document.go`

## Character
| Field | Type | Notes |
|---|---|---|
| Name | string | Populated from `CharacterInfo.Name` when built via `Load` |
| Animations | []air.Animation | Exposed through `air`'s read-path type only |
| Sprites | []sff.SpriteGroup | Exposed through `sff`'s read-path type only |
| StateDefs | []cns.StateDef | Exposed through `cns`'s read-path type only |

Method: `(*Character) ResolveSprite(frame air.Frame) (sff.Sprite, error)` — resolves `frame`'s `(Group, Image)` reference against `Sprites`, by delegating to `air.NewSpriteResolver(c.Sprites)`; returns the same descriptive error `SpriteResolver.Resolve` does when no match exists, including when `Sprites` is empty (e.g. a zero-value `Character`).

Package function: `Load(path string) (*Character, error)` — the top-level entry point: opens the `.def` file at `path`, parses it with `def.Parse`, resolves its `.air`/`.sff`/`.cns` references against `path`'s own directory, reads them with `air.Parse`/`sff.Load`/`cns.Parse`, and returns the assembled `Character`. A missing or unreadable `.def`/`.air`/`.sff`/`.cns` file returns a descriptive error rather than panicking.

Defined in: `character.go`, `load.go`

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

## Document (air)
| Field | Type | Notes |
|---|---|---|
| Animations | []Animation | Decoded the same way `Parse`'s return value is |
| source | []byte | Unexported; the exact bytes `ParseDocument` read, replayed verbatim by `Serialize` |

Write-path type: `ParseDocument`/`Document.Serialize` round-trip an unmodified `.air` file byte-for-byte, including comments. Mutating `Animations` does not change what `Serialize` writes.

Defined in: `air/document.go`

## SpriteResolver
| Field | Type | Notes |
|---|---|---|
| sprites | map[spriteKey]sff.Sprite | Unexported; indexes the sprite groups passed to `NewSpriteResolver` by (Group, Image) |

Built by `NewSpriteResolver(groups []sff.SpriteGroup) *SpriteResolver`; queried via `(*SpriteResolver) Resolve(frame Frame) (sff.Sprite, error)`, which returns a descriptive error rather than a zero `Sprite` for a frame referencing a (Group, Image) pair not present in the indexed groups.

Defined in: `air/resolve.go`

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

## V1WriteSprite
| Field | Type | Notes |
|---|---|---|
| Group | int | Sprite group index |
| Image | int | Image index within Group |
| AxisX | int | Horizontal offset from top-left corner to the axis (pivot) point |
| AxisY | int | Vertical offset from top-left corner to the axis (pivot) point |
| SharedPalette | bool | True when this sprite reuses the previous sprite's palette |
| PixelData | []byte | This sprite's PCX-encoded pixel data (e.g. from `EncodePCX`); empty to write a linked sprite instead |
| LinkedIndex | int | Index, within the `SerializeV1` call's sprite slice, this sprite links to; only meaningful when PixelData is empty |

Write-only counterpart to `V1SpriteEntry`, passed to `SerializeV1`; it has no `Offset`/`Length` fields since those are computed by the writer, not supplied.

Defined in: `sff/v1_serializer.go`

## V2Header
| Field | Type | Notes |
|---|---|---|
| Version | [4]byte | Raw version bytes as stored in the file (verlo3, verlo2, verlo1, verhi); Version[3] is 2 for a v2 file |
| SpriteCount | int | Total number of sprites declared in the header |
| PaletteCount | int | Total number of palette banks declared in the header |

Defined in: `sff/v2.go`

## V2SpriteEntry
| Field | Type | Notes |
|---|---|---|
| Group | int | Sprite group index |
| Image | int | Image index within Group |
| Width | int | Pixel width, as declared in the table |
| Height | int | Pixel height, as declared in the table |
| AxisX | int | Horizontal offset from top-left corner to the axis (pivot) point |
| AxisY | int | Vertical offset from top-left corner to the axis (pivot) point |
| Offset | int64 | Absolute file offset of this sprite's encoded pixel data, already resolved against the file's literal or translated data section |
| Length | int | Encoded pixel data length in bytes; 0 means this sprite links to another sprite's data (see LinkedIndex) |
| LinkedIndex | int | Index, within the owning V2SpriteTable's Sprites, of the sprite this one shares pixel data with; only meaningful when Length is 0 |
| Format | int | Pixel-data encoding code (see the V2Format* constants); only meaningful when Length is non-zero |
| ColorDepth | int | Bit depth of the encoded pixel data |
| PaletteIndex | int | Index, within the owning V2SpriteTable's Palettes, of the palette bank this sprite is drawn with |

Defined in: `sff/v2.go`

## V2PaletteEntry
| Field | Type | Notes |
|---|---|---|
| Group | int | Palette bank's group index |
| Number | int | Palette bank's index within Group |
| ColorCount | int | Number of colors declared for this palette bank |
| Offset | int64 | Absolute file offset of this palette bank's RGBA color data |
| Length | int | Color data length in bytes; 0 means this bank links to another bank's data (see LinkedIndex) |
| LinkedIndex | int | Index, within the owning V2SpriteTable's Palettes, of the palette bank this one shares color data with; only meaningful when Length is 0 |

Defined in: `sff/v2.go`

## V2SpriteTable
| Field | Type | Notes |
|---|---|---|
| Header | V2Header | The parsed file header |
| Sprites | []V2SpriteEntry | Sprite index table entries, in file order |
| Palettes | []V2PaletteEntry | Palette bank table entries, in file order |

Read helpers: `Offset(group, image int) (int64, bool)` resolves a `(group, image)` pair to its pixel data's file offset; `PaletteOffset(group, number int) (int64, bool)` resolves a `(group, number)` pair to its palette bank's color data file offset.

Defined in: `sff/v2.go`

## V2Image
| Field | Type | Notes |
|---|---|---|
| Width | int | Pixel width |
| Height | int | Pixel height |
| BytesPerPixel | int | 1 for indexed data (raw, PNG8), 3 for RGB (PNG24), 4 for RGBA (PNG32) |
| Pixels | []byte | Row-major buffer, length Width*Height*BytesPerPixel; indexed data holds palette index values (no RGB/palette resolution performed), direct-color data holds actual color channels |

Defined in: `sff/v2_decoder.go`

## V2WriteSprite
| Field | Type | Notes |
|---|---|---|
| Group | int | Sprite group index |
| Image | int | Image index within Group |
| Width | int | Sprite width in pixels |
| Height | int | Sprite height in pixels |
| AxisX | int | Horizontal offset from top-left corner to the axis (pivot) point |
| AxisY | int | Vertical offset from top-left corner to the axis (pivot) point |
| Format | int | How PixelData is encoded (see the V2Format* constants); only meaningful when PixelData is non-empty |
| ColorDepth | int | Bit depth of PixelData |
| PaletteIndex | int | Index, within the `SerializeV2` call's palette slice, of the palette bank this sprite is drawn with |
| PixelData | []byte | This sprite's already-encoded pixel data (e.g. from `EncodeV2Sprite`); empty to write a linked sprite instead |
| LinkedIndex | int | Index, within the `SerializeV2` call's sprite slice, this sprite links to; only meaningful when PixelData is empty |

Write-only counterpart to `V2SpriteEntry`, passed to `SerializeV2`; it has no `Offset`/`Length` fields since those are computed by the writer, not supplied.

Defined in: `sff/v2_serializer.go`

## StateDef
| Field | Type | Notes |
|---|---|---|
| Number | int | State number this block defines (the N in `[Statedef N]`) |
| Type | StateType | State classification (`.cns` "type" parameter) |
| MoveType | MoveType | Move classification (`.cns` "movetype" parameter) |
| Physics | PhysicsType | Built-in physics applied while active (`.cns` "physics" parameter) |
| Anim | int | Animation number played on entering this state; 0 means "not set" (defaults to Number) |
| Ctrl | bool | Whether the player has control while this state is active |
| PowerAdd | int | Power meter gain applied on entering this state |
| Juggle | int | Juggle points this state costs against an airborne opponent |
| FaceP2 | bool | Whether the character turns to face the opponent on entering this state |
| HitDefPersist | bool | Whether an active hit definition survives into this state |
| MoveHitPersist | bool | Whether "MoveHit"-triggered conditions survive into this state |
| HitCountPersist | bool | Whether the hit counter survives into this state |
| SprPriority | int | Sprite drawing (layering) priority for this state |
| Controllers | []Controller | State controllers that run while this state is active, in file order |

Populated by `Parse(r io.Reader) ([]StateDef, error)`, which reads `[Statedef N]`/`[State N]` `.cns` text into this shape, skipping unrecognized sections. Written back out by `Serialize(w io.Writer, states []StateDef) error` (a semantic round trip through `Parse`, not byte-exact preservation of an original file's formatting/comments — see `Document (cns)` below for that case). Not yet wired into the root `Character` struct (backlog item 022). See `.vibe/decisions/011-cns-controller-parameters-are-untyped-key-value-data.md` and `.vibe/decisions/012-cns-parse-header-detection-strategy.md`.

Defined in: `cns/statedef.go`, `cns/parser.go`

## Controller
| Field | Type | Notes |
|---|---|---|
| Type | string | Controller type (`.cns [State N]` "type" parameter, e.g. "ChangeState", "VelSet") |
| Triggers | []string | Trigger condition expressions, verbatim and unevaluated, in file order; nil/empty means the controller runs unconditionally |
| Parameters | map[string]string | Remaining key/value parameters, verbatim and unevaluated, keyed by lowercase parameter name |

A controller's effect (e.g. which state a "ChangeState" controller transitions to) is just another `Parameters` entry, not a dedicated field.

Defined in: `cns/statedef.go`, `cns/parser.go`

## Document (cns)
| Field | Type | Notes |
|---|---|---|
| StateDefs | []StateDef | Decoded the same way `Parse`'s return value is, for convenient structured access |
| source | []byte | Unexported; the exact bytes `ParseDocument` read, replayed verbatim by `Serialize` |

Write-path type: `ParseDocument`/`Document.Serialize` round-trip an unmodified `.cns` file byte-for-byte, including comments, block ordering, and unrecognized sections. Mutating `StateDefs` does not change what `Serialize` writes. Mirrors `air`'s and `def`'s own `Document` (see above).

Defined in: `cns/document.go`

## StateType / MoveType / PhysicsType
String-based enums matching `.cns [Statedef N]` header tokens.
| Type | Values | Notes |
|---|---|---|
| StateType | `StateTypeStanding` ("S"), `StateTypeCrouching` ("C"), `StateTypeAir` ("A"), `StateTypeLiedown` ("L"), `StateTypeUnchanged` ("U") | "type" parameter |
| MoveType | `MoveTypeAttack` ("A"), `MoveTypeIdle` ("I"), `MoveTypeHit` ("H"), `MoveTypeUnchanged` ("U") | "movetype" parameter |
| PhysicsType | `PhysicsStanding` ("S"), `PhysicsCrouching` ("C"), `PhysicsAir` ("A"), `PhysicsNone` ("N"), `PhysicsUnchanged` ("U") | "physics" parameter |

Defined in: `cns/statedef.go`

## V2WritePalette
| Field | Type | Notes |
|---|---|---|
| Group | int | Palette bank's group index |
| Number | int | Palette bank's index within Group |
| ColorCount | int | Number of colors in this palette bank |
| ColorData | []byte | This bank's own RGBA color data; empty to write a linked bank instead |
| LinkedIndex | int | Index, within the `SerializeV2` call's palette slice, this bank links to; only meaningful when ColorData is empty |

Write-only counterpart to `V2PaletteEntry`, passed to `SerializeV2`.

Defined in: `sff/v2_serializer.go`

## Palette
`[256]color.RGBA` — a resolved color table, indexed by a decoded sprite's palette index bytes (`PCXImage.Pixels` / `V2Image.Pixels` with `BytesPerPixel: 1`). Kept separate from `PCXImage`/`V2Image`/`Sprite`; produced by `DecodeV1Palette`/`ResolveV1Palette`, `DecodeV2Palette`/`ResolveV2Palette`, or `DecodeExternalPalette` (an external `.act` file, used as an optional `override` argument to `ResolveV1Palette`/`ResolveV2Palette` in place of a sprite's own); consumed by `ResolvePixels`.

Defined in: `sff/palette.go`

## AlphaRule
`int`-based enum selecting how `ResolvePixels` determines a resolved pixel's alpha at palette index 0.
| Value | Notes |
|---|---|
| AlphaForceTransparentAtIndexZero | Forces index 0 to `(0,0,0,0)` regardless of the palette's own stored value; used for PCX (v1) and PNG8 (v2) decoded pixel data |
| AlphaLiteral | Uses the palette's own stored alpha unmodified, including at index 0; used for RLE8/LZ5 (v2) decoded pixel data |

Defined in: `sff/palette.go`
