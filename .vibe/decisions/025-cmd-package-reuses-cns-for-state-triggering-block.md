---
date: 2026-08-10
status: accepted
---
# `cmd` package reuses `cns.Parse`/`cns.Serialize` for the `.cmd` state-triggering block

**Context:** A `.cmd` file has two structurally distinct parts: `[Remap]`/`[Defaults]`/`[Command]` sections unique to the format, and a `[Statedef -1]` block containing `[State ...]` controllers — byte-for-byte the same syntax `.cns` already parses, typically using an unevaluated `command = "name"` trigger to link a command to a state change.

**Decision:** The new `cmd` package parses/serializes only the `Remap`/`Defaults`/`Command` sections itself. For the `[Statedef -1]`/`[State ...]` portion, it calls `cns.Parse`/`cns.Serialize` directly (already tolerant of unrelated bracket sections, which it skips) rather than reimplementing state/controller parsing a second time.

**Reason:** The two formats share this block byte-for-byte; `cns.Parse` already skips sections it doesn't recognize, so running it against a `.cmd` file's raw text picks out just the state-triggering block for free. The command-to-state "link" itself needs no new modeling — it already flows through `cns.Controller`'s existing unevaluated `Triggers` strings (e.g. `command = "holdback"`), the same "read-path model can't hold everything yet" pattern this repo already applies elsewhere.

**Rejected alternatives:** Reimplementing a second `StateDef`/`Controller`-equivalent inside `cmd` — rejected as pure duplication of already-working, already-tested code for an identical syntax, with no format-specific reason to diverge.
