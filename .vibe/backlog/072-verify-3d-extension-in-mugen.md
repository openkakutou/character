---
status: todo
depends_on: [071]
---
# Verify that MUGEN 1.1 ignores the 3D character extension

## Description
The 3D character spec (`docs/3d-character-extension.md`, item 069) proves from Ikemen GO's source that a `Pose3D:` line in `.air` and a `model` key in `.def` `[Files]` are skipped without error. For MUGEN 1.0/1.1 this is unverified: its source is closed and its docs say nothing about it. Load the 071 sample, plus a sprite character with one `Pose3D:` line added, in MUGEN 1.1, and record the result in the spec and in ADR 034.

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO compatibility, the extension strictly additive.

## Acceptance Criteria
- [ ] MUGEN 1.1's reaction to a `Pose3D:` line and to a `model` key is observed and written into the spec's "Why this syntax" section and section 1, replacing "not verified"
- [ ] If MUGEN rejects or misreads either one, a new syntax is proposed in a superseding ADR
