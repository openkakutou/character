---
status: in_progress
---
# Add `.cmd` File Parsing (Commands & Triggers)

## Description
MUGEN/Ikemen GO characters define their input commands (motion + button sequences that map to special moves) in a `.cmd` file, separate from `.air`/`.cns`. This repo doesn't parse it at all today. Add a `cmd` package (parse + serialize, following the same read/write and format-preservation principles as `def`/`air`/`cns`) that models command definitions (name, input sequence, time window) and their state-triggering links. Needed by `character-viewer-web`'s special-move list feature and by `engine`'s future input matching.

## Acceptance Criteria
- [ ] `.cmd` files parse into a structured `Command` model (name, input sequence, buffer time)
- [ ] Round-trip serialize preserves original formatting for unmodified content, same guarantee as other formats
- [ ] Both MUGEN-style and Ikemen GO-style `.cmd` syntax variants parse correctly (fixture-driven against real files, same practice as `.cns`)
- [ ] A malformed `.cmd` file returns a descriptive error instead of crashing

## Notes
Compatibility target is MUGEN 1.0/1.1 and Ikemen GO — see `roadmap`'s `.vibe/decisions/009` compatibility note.
