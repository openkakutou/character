# Module: sff

**Role:** Read-path pure-data model for MUGEN/Ikemen GO sprite (`.sff`) files. `sprite.go` defines `Sprite` and `SpriteGroup` — the stable vocabulary a library consumer (editor, engine) works with — deliberately version-agnostic (no v1/v2-specific fields), since `.sff` has two on-disk versions with different header layouts and pixel encodings that both need to populate the same shape. No binary decoding, file I/O, or pixel data yet; that is left to the v1 (backlog item 007) and v2 (item 010) parsers this model exists to feed.
**Files:** `sff/sprite.go`
**Exports:** `Sprite` (struct: `Group`, `Image`, `Width`, `Height`, `AxisX`, `AxisY`, `Palette`), `SpriteGroup` (struct: `Index`, `Sprites []Sprite`)
**Depends on:** none
