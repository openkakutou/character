# Module: cns

**Role:** MUGEN/Ikemen GO combat logic (`.cns`) files — a character's states and the state controllers that run within them. Currently scaffolding only: the pure-data model (`StateDef`, `Controller`) and the `StateType`/`MoveType`/`PhysicsType` enums for a Statedef header. No `.cns` text parser, serializer, or wiring into the root `Character` struct exists yet (backlog items 020–022).

**Files:** `cns/statedef.go`

**Exports:**
- `StateDef` (struct: `Number`, `Type`, `MoveType`, `Physics`, `Anim`, `Ctrl`, `PowerAdd`, `Juggle`, `FaceP2`, `HitDefPersist`, `MoveHitPersist`, `HitCountPersist`, `SprPriority`, `Controllers []Controller`) — a `.cns [Statedef N]` block's header parameters plus its state controllers, in file order
- `Controller` (struct: `Type string`, `Triggers []string`, `Parameters map[string]string`) — a `.cns [State N]` block, stored as unevaluated data rather than resolved or type-checked; a nil/empty `Triggers` means the controller runs unconditionally, not "never runs" (see `.vibe/decisions/011-cns-controller-parameters-are-untyped-key-value-data.md`)
- `StateType`, `MoveType`, `PhysicsType` (string-based enums matching `.cns` header tokens, e.g. `StateTypeStanding = "S"`)

**Depends on:** nothing
