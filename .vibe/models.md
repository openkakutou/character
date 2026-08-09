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

Every field carries a `json:"..."` tag (`name`, `animations`, `sprites`, `stateDefs`).

Method: `(*Character) ResolveSprite(frame air.Frame) (sff.Sprite, error)` — resolves `frame`'s `(Group, Image)` reference against `Sprites`, by delegating to `air.NewSpriteResolver(c.Sprites)`; returns the same descriptive error `SpriteResolver.Resolve` does when no match exists, including when `Sprites` is empty (e.g. a zero-value `Character`).

Package function: `Load(path string) (*Character, error)` — the top-level entry point: opens the `.def` file at `path`, parses it with `def.Parse`, resolves its `.air`/`.sff`/`.cns` references against `path`'s own directory, reads them with `air.Parse`/`sff.Load`/`cns.Parse`, and returns the assembled `Character`. A missing or unreadable `.def`/`.air`/`.sff`/`.cns` file returns a descriptive error rather than panicking.

Package function: `LoadBytes(defBytes, airBytes, sffBytes, cnsBytes []byte) (*Character, error)` — filesystem-independent counterpart of `Load`, for a caller (chiefly `cmd/wasm`) that already holds each file's bytes in memory: parses each buffer directly, never resolving referenced paths itself. Every slice/map reachable from the returned `Character` is normalized to non-`nil` before returning, so its JSON marshaling never surprises a caller with `null` where an empty, iterable collection was expected. See `.vibe/decisions/019-wasm-entrypoint-byte-buffer-loading-and-json-contract.md`.

Defined in: `character.go`, `load.go`, `load_bytes.go`

## Animation
| Field | Type | Notes |
|---|---|---|
| Number | int | Corresponds to the `.air` file's `[Begin Action N]` action number |
| Frames | []Frame | Ordered sequence of displayed frames |
| LoopStart | int | Index into Frames the animation loops back to; zero value (0) matches `.air`'s own default of looping to the first frame when no `Loopstart` marker is present |

`json:"..."`-tagged (`number`, `frames`, `loopStart`).

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

`json:"..."`-tagged (`group`, `image`, `x`, `y`, `time`, `flip`, `blend`, `clsn1`, `clsn2`).

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

`json:"..."`-tagged (`left`, `top`, `right`, `bottom`).

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

`json:"..."`-tagged (`group`, `image`, `width`, `height`, `axisX`, `axisY`, `palette`).

Defined by the external `github.com/openkakutou/sff` module (item 035); consumed in: `character.go`, `load.go`, `load_bytes.go`, `air/resolve.go`

## SpriteGroup
| Field | Type | Notes |
|---|---|---|
| Index | int | Group index shared by every Sprite in Sprites |
| Sprites | []Sprite | Ordered collection of sprites belonging to this group; not itself validated against each Sprite's own Group field |

`json:"..."`-tagged (`index`, `sprites`).

Defined by the external `github.com/openkakutou/sff` module (item 035); consumed in: `character.go`, `load.go`, `load_bytes.go`, `air/resolve.go`

`.sff`-specific low-level types (`V1Header`/`V1SpriteEntry`/`V1SpriteTable`/`V1WriteSprite`/`PCXImage`,
`V2Header`/`V2SpriteEntry`/`V2PaletteEntry`/`V2SpriteTable`/`V2Image`/`V2WriteSprite`/`V2WritePalette`,
`AlphaRule`) moved out of this repo along with the `sff` package itself (item 035) — they now live in
`github.com/openkakutou/sff`'s own model docs; this repo no longer references them by name.

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
| HeaderExprs | map[string]string | Raw, unevaluated source text (keyed by lowercase field name) of `anim`/`poweradd`/`juggle`/`sprpriority` whenever the source held a MUGEN trigger expression rather than a literal integer; the corresponding typed field stays at its zero value in that case |
| Controllers | []Controller | State controllers that run while this state is active, in file order |

`json:"..."`-tagged (`number`, `type`, `moveType`, `physics`, `anim`, `ctrl`, `powerAdd`, `juggle`, `faceP2`, `hitDefPersist`, `moveHitPersist`, `hitCountPersist`, `sprPriority`, `headerExprs`, `controllers`).

Populated by `Parse(r io.Reader) ([]StateDef, error)`, which reads `[Statedef N]`/`[State ...]` `.cns` text into this shape, skipping unrecognized sections. A `[State ...]` header's own content (numbered or a bare label) is unconstrained and discarded. Written back out by `Serialize(w io.Writer, states []StateDef) error` (a semantic round trip through `Parse`, not byte-exact preservation of an original file's formatting/comments — see `Document (cns)` below for that case); a field with a `HeaderExprs` entry is written using that raw text instead of its typed value. Wired into the root `Character` struct via `character.Load`/`character.LoadBytes`. See `.vibe/decisions/011-cns-controller-parameters-are-untyped-key-value-data.md`, `.vibe/decisions/012-cns-parse-header-detection-strategy.md`, `.vibe/decisions/022-cns-parse-state-header-accepts-any-label.md`, and `.vibe/decisions/023-statedef-numeric-header-fields-unevaluated-expression-escape-hatch.md`.

Defined in: `cns/statedef.go`, `cns/parser.go`

## Controller
| Field | Type | Notes |
|---|---|---|
| Type | string | Controller type (`.cns [State N]` "type" parameter, e.g. "ChangeState", "VelSet") |
| Triggers | []string | Trigger condition expressions, verbatim and unevaluated, in file order; nil/empty means the controller runs unconditionally |
| Parameters | map[string]string | Remaining key/value parameters, verbatim and unevaluated, keyed by lowercase parameter name |

A controller's effect (e.g. which state a "ChangeState" controller transitions to) is just another `Parameters` entry, not a dedicated field.

`json:"..."`-tagged (`type`, `triggers`, `parameters`).

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

## Palette
`[256]color.RGBA` — a resolved color table, indexed by a decoded sprite's palette index bytes. Produced by `DecodeV1Palette`/`ResolveV1Palette`, `DecodeV2Palette`/`ResolveV2Palette`, or `DecodeExternalPalette` (an external `.act` file, used as an optional `override` argument in place of a sprite's own).

Defined by the external `github.com/openkakutou/sff` module (item 035); consumed in: `cmd/wasm/main.go` (palette-override support)
