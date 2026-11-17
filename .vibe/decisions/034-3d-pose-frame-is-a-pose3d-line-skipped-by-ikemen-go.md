---
date: 2026-11-17
status: accepted
---
# A 3D `.air` frame is a `Pose3D:` keyword line, not a variant of the sprite frame line

**Context:** Roadmap decision `028` adds 3D pose frames `(clip, tick)` to `.air`, coexisting with sprite frames, and left the line syntax to this repo, constrained by what MUGEN/Ikemen GO parsers tolerate. 🔒 Existing MUGEN/Ikemen GO files must keep parsing unchanged.
**Decision:** A 3D frame is the line `Pose3D: <clip>, <tick>, <duration>` (specified in `docs/3d-character-extension.md`). It starts with a letter and avoids the keyword prefixes Ikemen GO matches (`copy action `, `loopstart`, `interpolate ...`, `clsn`). Ikemen GO's `ReadAnimFrame` only reads lines starting with a digit or `-` as frames, so it skips the line silently (`src/anim.go:45-48`, commit `54d4148a728153c2097a0fb899a06812b2d3e193`). MUGEN 1.0/1.1's reaction is unverified (closed source).
**Reason:** A keyword line cannot be misread as `Group,Image,X,Y,Time`, needs no change to any existing line, and is skipped rather than rejected by the one engine whose source was read.
**Rejected alternatives:** Reusing the frame line with a negative group and the clip in the image field (a sprite parser reads it as a blank frame and misreads the data); a new section in the `.air` (an unknown bracket line is skipped together with the content after it, which would lose any sprite frames mixed in the same action); a separate file (breaks the mixed sprite/3D action and the single source of truth for `Clsn`).
