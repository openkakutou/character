# Module: def

**Role:** Pure-data model for MUGEN/Ikemen GO character definition (`.def`) files — the entry point that references a character's other files (sprite, animation, sound, commands, combat logic, additional states, palettes). Parsing `.def` INI-style text into this model is not implemented yet (backlog item 016), nor is wiring it into the root `Character` struct (backlog item 018).

**Files:** `def/character_info.go`

**Exports:** `CharacterInfo` (struct: `Name`, `Author`, `SpriteFile`, `AnimationFile`, `SoundFile`, `CommandFile`, `ConstantsFile`, `StateFiles []string`, `Palettes []string`)

**Depends on:** nothing (no dependency on `air`/`sff`, matching backlog item 015's note; item 018 will wire it to them)
