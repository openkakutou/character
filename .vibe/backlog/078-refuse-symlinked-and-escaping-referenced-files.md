---
status: todo
---
# Refuse symlinked and escaping files referenced from a `.def`

## Description
`Load` opens the files a character's `.def` references (`sprite`, `anim`, `cns`, `st*`, `sound`) through `openReferencedFile` (`load.go`). It already accepts `\` and falls back to a case-insensitive match, but it follows symbolic links and accepts a path that leaves the character's folder (`..\..\x.sff`) or an absolute one. Roadmap decision `033` §4 says every reader refuses symbolic links and keeps paths relative to the game folder. `game` enforces this for the paths it builds itself (`game` ADR 017, `native/internal/game/paths.go` `ResolvePath`), but a character's own files are opened here, so a community pack can still make the game read a file outside its folder.

Apply the same rules as `game`: resolve `..` textually before touching the disk, refuse a result that leaves the folder, refuse absolute paths, and refuse any symbolic-link component. Return a descriptive error, never a panic. Keep the existing case-insensitive and `\` handling.

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO content still loads as is: a `..` that stays inside the character's folder keeps working.

## Acceptance Criteria
- [ ] A `.def` referencing a symbolic-link sprite, anim, cns or sound file fails to load with a descriptive error
- [ ] A `.def` referencing a file outside its folder (`..` escape, including one behind a missing folder, or an absolute path) fails to load with a descriptive error
- [ ] Mixed-case and `\` references inside the folder still load

## Notes
Raised by `game#037` (2026-11-18). Once released, `game` bumps its `character` dependency.
