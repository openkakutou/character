---
date: 2026-08-06
status: accepted
---
# Frame Clsn boxes are pre-resolved, not default+override

**Context:** Defining the pure-data `Animation`/`Frame`/`ClsnBox` model for the `air` package (backlog item 001). The `.air` text format lets a `Clsn1Default`/`Clsn2Default` declaration apply collision boxes to every frame of an animation unless a specific frame overrides them with its own indexed `Clsn[i]` lines.

**Decision:** `Frame` stores fully-resolved `Clsn1`/`Clsn2` box slices — the boxes actually active on that frame — rather than modeling the file's default/override authoring mechanism in the struct.

**Reason:** Per CLAUDE.md's design constraint, the read path must be "the surface a future game engine would consume", designed as if the engine were already a real client. An engine only needs "what boxes are active on this frame", not the `.air` file's default/override authoring shortcut. Resolving defaults is a parsing concern (item 002), which keeps the write/parse-only logic out of the pure-data type the read-only API exposes, per the read/write separation constraint.

**Rejected alternatives:** Storing a separate `Defaults` field on `Animation` with per-frame override flags — rejected because it leaks a parsing/authoring concern into the pure-data read API and would force every consumer to re-implement the resolution logic themselves.
