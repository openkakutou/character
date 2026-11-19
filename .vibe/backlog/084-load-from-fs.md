---
status: todo
---
# Load a character from an `fs.FS`

## Description
Roadmap decision `045` (`game` item 062): `game`'s shared core must load content without touching the file system, so the native build passes the game folder and the web build passes in-memory files. Add `LoadFS(fsys fs.FS, name string)` (slash-separated name of the `.def` inside `fsys`) doing exactly what `Load` does today: every referenced file (`.air`, `.sff`, `.snd`, `.cmd`, `.cns`, `.zss`, palettes) resolved relative to the `.def`, Windows backslashes accepted, the case-insensitive fallback of backlog 050 kept. `Load(path)` keeps its behavior and is built on `LoadFS`, so there is one implementation.

🔒 MUST: no change for any character `Load` reads today (real-file corpus included).

## Acceptance Criteria
- [ ] `LoadFS` over an in-memory `fstest.MapFS` loads a real character identically to `Load` on disk
- [ ] Case-insensitive and backslash references resolve inside an `fs.FS` as on disk
- [ ] A reference that leaves the root of `fsys` is reported as a missing file, never a panic
- [ ] `Load` is implemented on top of `LoadFS`; existing tests pass unchanged
