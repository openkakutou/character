---
date: 2026-08-06
status: accepted
---
# Controller triggers/parameters (including transition targets) are untyped key-value data, not dedicated struct fields

**Context:** Defining the minimal `cns` data model (`StateDef`/`Controller`) as scaffolding for a future `.cns` parser (item 020), without building a trigger-expression engine yet. `.cns` state controllers (e.g. `ChangeState`) can set a "transition target" (which state to switch to next), among many other controller-type-specific parameters.

**Decision:** `Controller` stores its trigger conditions and parameters as plain, unevaluated data (`Triggers []string`, `Parameters map[string]string`) rather than modeling controller-type-specific fields such as a dedicated `TransitionTarget` field. A `ChangeState` controller's target state is just another entry in `Parameters` (e.g. `"value"`), exactly like any other controller type's parameters.

**Reason:** `.cns` has dozens of controller types, each with its own parameter shape, and MUGEN/Ikemen's own trigger expressions are a small expression language, not fixed data — modeling either as typed fields would mean building the expression engine now, which item 019 explicitly scopes out ("no full expression engine", "not evaluated"). A generic key-value shape covers every controller type uniformly and defers typed/evaluated modeling to whichever future item actually needs it.

**Rejected alternatives:** A dedicated `TransitionTarget int` field on `Controller` for `ChangeState`-like controllers — rejected because it special-cases one controller type while leaving every other type's parameters generic, an inconsistent model for no real benefit at this scaffolding stage.
