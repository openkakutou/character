---
date: 2026-08-10
status: accepted
---
# `cmd.Parse` synthesizes an implicit "[Statedef -1]" header before delegating to `cns.Parse`

**Context:** `cmd.Parse` delegates its `[Statedef -1]`/`[State ...]` block to `cns.Parse` (see `.vibe/decisions/025`). A 520-file real-character corpus scan surfaced two files (Marvel's "Jean Grey" and "Nova") whose `.cmd` never declares `[Statedef -1]` at all — their `[State -1, ...]` controllers sit directly after the last `[Command]` block. `cns.Parse` alone rejects this (`[State ...] block found outside of any Statedef`), since a bare `.cns` file always requires an explicit Statedef header.

**Decision:** Before calling `cns.Parse`, `cmd.Parse` scans the source for an explicit `[Statedef ...]` header; if none is found before the first `[State ...]` header, it prepends a synthetic `[Statedef -1]` line to the text passed to `cns.Parse`. A file that already declares its own `[Statedef ...]` header is passed through unmodified.

**Reason:** A `.cmd` file's "always" section can only ever be state `-1` — there is no other valid Statedef number in this context — so the header carries no information a reader doesn't already know, and real MUGEN/Ikemen engines evidently tolerate its omission. Synthesizing it at the `cmd` package boundary keeps the fix scoped to `.cmd`-specific semantics rather than weakening `cns.Parse`'s own stricter requirement, which is correct for genuine `.cns` files.

**Rejected alternatives:** Relaxing `cns.Parse` itself to allow a `[State ...]` block with no enclosing Statedef — rejected because a `.cns` file has no equivalent "only one possible number" convention; doing so there would silently misattribute an actually-missing Statedef in a real `.cns` file instead of surfacing it as an error.
