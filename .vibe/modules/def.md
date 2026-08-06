# Module: def

**Role:** MUGEN/Ikemen GO character definition (`.def`) files — the entry point that references a character's other files (sprite, animation, sound, commands, combat logic, additional states, palettes). Provides the pure-data model (`CharacterInfo`) and a text parser that reads `.def` INI-style text into it. Serializing back to `.def` text is not implemented yet (backlog item 017), nor is wiring `CharacterInfo` into the root `Character` struct (backlog item 018).

**Files:** `def/character_info.go`, `def/parser.go`

**Exports:**
- `CharacterInfo` (struct: `Name`, `Author`, `SpriteFile`, `AnimationFile`, `SoundFile`, `CommandFile`, `ConstantsFile`, `StateFiles []string`, `Palettes []string`)
- `Parse(r io.Reader) (CharacterInfo, error)` — reads `[Info]`/`[Files]` `.def` text into a `CharacterInfo`; unrecognized sections and keys are skipped rather than aborting the parse; `StateFiles` keeps file-appearance order, `Palettes` is sorted by numeric suffix; a malformed section header or `key=value` line returns a descriptive error naming the line (see `.vibe/decisions/009-def-parse-ignores-unknown-sections.md`)

**Depends on:** nothing (no dependency on `air`/`sff`, matching backlog item 015's note; item 018 will wire it to them)
