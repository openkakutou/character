---
date: 2026-10-06
status: accepted
---
# Per-sound decode errors do not fail the load
**Context:** The sound browser in `character-editor` must show working sounds and flag broken ones; `Load`/`LoadBytes` used to abort on the first undecodable `.snd` entry.
**Decision:** An entry that fails to decode stays in its group with an `Error` string and no PCM; unparsable files (bad signature/table) remain hard errors.
**Reason:** An editor must let the user see and fix broken entries rather than lose the whole sound set.
**Rejected alternatives:** Returning a separate list of per-entry errors (splits one sound across two structures); dropping bad entries silently (hides the problem).
