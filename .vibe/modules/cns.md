# Module: cns

**Role:** MUGEN/Ikemen GO combat logic (`.cns`) files — a character's states and the state controllers that run within them. Has the pure-data model (`StateDef`, `Controller`) and the `StateType`/`MoveType`/`PhysicsType` enums for a Statedef header, plus a read-path text parser (`Parse`). No serializer or wiring into the root `Character` struct exists yet (backlog items 021–022).

**Files:** `cns/statedef.go`, `cns/parser.go`

**Exports:**
- `StateDef` (struct: `Number`, `Type`, `MoveType`, `Physics`, `Anim`, `Ctrl`, `PowerAdd`, `Juggle`, `FaceP2`, `HitDefPersist`, `MoveHitPersist`, `HitCountPersist`, `SprPriority`, `Controllers []Controller`) — a `.cns [Statedef N]` block's header parameters plus its state controllers, in file order
- `Controller` (struct: `Type string`, `Triggers []string`, `Parameters map[string]string`) — a `.cns [State N]` block, stored as unevaluated data rather than resolved or type-checked; a nil/empty `Triggers` means the controller runs unconditionally, not "never runs" (see `.vibe/decisions/011-cns-controller-parameters-are-untyped-key-value-data.md`)
- `StateType`, `MoveType`, `PhysicsType` (string-based enums matching `.cns` header tokens, e.g. `StateTypeStanding = "S"`)
- `Parse(r io.Reader) ([]StateDef, error)` — reads `.cns` text into `StateDef`s, in file order; a `trigger`-prefixed key appends to a controller's `Triggers`, `type` sets `Controller.Type`, every other key becomes a lowercase-normalized `Parameters` entry; bracket sections other than `[Statedef N]`/`[State N]` are skipped without validation, but a header that looks like an attempted Statedef/State header yet fails to parse returns a line-numbered error, as does a `[State N]` block with no enclosing Statedef (see `.vibe/decisions/012-cns-parse-header-detection-strategy.md`)

**Depends on:** nothing
