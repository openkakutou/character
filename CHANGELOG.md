# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `.sff` v2 sprites with no pixel data of their own (a common "reuse the character's first sprite" shorthand real files use) can now have their image resolved instead of reporting an error

### Fixed

- `.sff` v2 sprites using the 32-bit PNG pixel format (PNG32) now show their true colors again for any pixel that isn't fully opaque: a previous fix (0.2.0) had them alpha-premultiplied, darkening semi-transparent areas, but that "fix" was itself validated against a broken comparison that silently hid the very regression it introduced
- `.sff` v2 sprites using the 8-bit indexed PNG pixel format (PNG8) now always show fully opaque colors outside their transparent index, instead of occasionally showing an unintended partial transparency some real files' own embedded color data carries

## [0.2.0] - 2026-08-08

### Added

- A sprite's actual image can now be resolved directly in a web browser via the WASM module — not just its dimensions and metadata as before — including recoloring it with an external palette file; multiple sprites can be resolved in a single call so a whole sprite sheet doesn't require one round trip per sprite

### Fixed

- `.sff` v2 sprites using the PNG-encoded pixel formats (PNG8/24/32) can now be decoded at all: their pixel data was never actually valid PNG bytes as read, because a 4-byte length header real files store ahead of the PNG data (already correctly skipped for the RLE8/LZ5 formats) was never skipped for the PNG formats
- `.sff` v2 sprites using the 32-bit PNG pixel format (PNG32) now resolve their correct colors when partially transparent: their color data is alpha-premultiplied on disk, like the 24-bit format already was, but was previously read as if it were not, distorting colors wherever a pixel wasn't fully opaque or fully transparent

## [0.1.0] - 2026-08-08

### Added

- Defined the animation and frame data model (`Animation`, `Frame`, `ClsnBox`) that will represent MUGEN/Ikemen GO character animations once `.air` file reading is implemented
- `.air` animation files can now be read into animation data: actions, frame sequences, collision boxes, and loop points are all parsed from the file's text format
- `.air` reading now handles malformed and unusual input gracefully: comment lines are ignored, malformed action headers and frame lines report a clear error naming the offending line, negative sprite group/image indices are rejected, and an empty file returns an empty result instead of an error
- Animations can now be exported back into `.air` text — actions, frame sequences, collision boxes, and loop points are all written out in valid, re-readable MUGEN/Ikemen syntax
- `.air` files can now be round-tripped without losing comments: loading a file and saving it back unchanged reproduces the original text exactly, including comment lines, so re-saving an untouched file no longer produces a noisy diff
- Defined the sprite and sprite group data model (`Sprite`, `SpriteGroup`) that will represent a MUGEN/Ikemen GO character's sprites once `.sff` file reading is implemented
- `.sff` v1 sprite sheet files can now have their header and sprite index table read, resolving each sprite's (group, image) key to where its image data lives in the file; malformed or truncated files report a descriptive error instead of crashing
- `.sff` v1 sprite image data (PCX-encoded pixels) can now be decoded into a plain pixel buffer with its width and height; corrupted or truncated pixel data reports a descriptive error instead of crashing
- Sprites can now be saved back out to a valid `.sff` v1 file, including their PCX-encoded pixel data, sprite-linking (shared pixel data), and palette-sharing settings; saved files load back correctly with no image data lost
- `.sff` v2 sprite sheet files (the Ikemen-compatible format) can now have their header and sprite/palette index tables read, resolving each sprite's (group, image) key and each palette bank's (group, number) key to where their data lives in the file, including shared/linked sprites and palette banks; malformed or truncated files report a descriptive error instead of crashing
- `.sff` v2 sprite image data can now be decoded into a plain pixel buffer, supporting both raw (uncompressed) and PNG-encoded sprites (indexed and true-color); an unrecognized or not-yet-supported encoding reports a descriptive error instead of crashing
- Sprites can now be saved back out to a valid `.sff` v2 file, including raw and PNG-encoded pixel data, sprite-linking (shared pixel data), and palette bank data with palette-linking; saved files load back correctly with no image data lost and all palette bank references preserved
- Animation frames can now be matched to their actual sprite image from a loaded sprite collection (from either `.sff` version), with a clear error reported if a frame points to a sprite that doesn't exist instead of silently showing nothing
- A character's animations and sprites can now be assembled into a single `Character`, matching each animation frame to its actual sprite image directly from that character
- Defined the character definition data model (`CharacterInfo`) that will represent a MUGEN/Ikemen GO character's name, author, and the sprite, animation, sound, command, combat logic, state, and palette files it references, once `.def` file reading is implemented
- `.def` character definition files can now be read: name, author, and every referenced file (sprite, animation, sound, commands, combat logic, extra states, palettes) are parsed from the file's text format, with unrecognized sections skipped rather than aborting the read and malformed lines reporting a clear error naming the offending line
- Character definitions can now be saved back out to `.def` text, and re-saving an untouched file reproduces it exactly — comments, section ordering, and unrecognized sections included — so re-saving without changes no longer produces a noisy diff
- A full character can now be loaded from a single `.def` file: its name, animations, and sprites (either `.sff` version) are all read and assembled automatically from the files it references, ready to use; a missing or unreadable referenced file reports a clear error instead of crashing
- Defined the combat logic data model (`StateDef`, `Controller`) that will represent a MUGEN/Ikemen GO character's states and their behavior once `.cns` file reading is implemented
- `.cns` combat logic files can now be read: every state (`Statedef`) and the behaviors it runs (state controllers) are parsed from the file's text format, with their trigger conditions and parameters kept as raw data rather than evaluated; unrelated sections are skipped rather than aborting the read, and a malformed state or behavior header reports a clear error naming the offending line
- Combat logic can now be saved back out to `.cns` text, and re-saving an untouched file reproduces it exactly — comments, block ordering, and unrelated sections included — so re-saving without changes no longer produces a noisy diff
- A full character can now be loaded from a single `.def` file with its combat logic (`.cns`) included alongside its animations and sprites, completing the full character load; a missing or unreadable combat logic file reports a clear error instead of crashing
- Animation frames using the `-1,-1` "no sprite shown" convention widely found in real MUGEN/Ikemen GO characters are now read successfully instead of being rejected as invalid, and resolve to no sprite instead of a failed lookup
- Decoded sprite pixel data can now be resolved into its actual on-screen colors using the sprite's palette, including reading a `.sff` v1 sprite's own embedded palette and following `.sff` v2 palette bank sharing/linking, with the correct transparency rule applied depending on how the sprite was encoded
- `.sff` v2 sprites using the run-length-compressed "RLE8" pixel format can now be decoded, alongside the previously supported raw and PNG formats; malformed or truncated compressed data reports a descriptive error instead of crashing
- `.sff` v2 sprites using the dictionary-compressed "LZ5" pixel format can now be decoded, alongside the previously supported raw, RLE8, and PNG formats; malformed or truncated compressed data reports a descriptive error instead of crashing
- A character's colors can now be recolored using an external `.act` palette file instead of a sprite's own — resolving a sprite's colors accepts an optional override palette, and a `.act` file's colors are read correctly regardless of its reversed on-disk color order; a wrong-sized `.act` file reports a descriptive error instead of crashing
- `.sff` v1 sprites can now have their actual decoded pixel data resolved through the same public API as their palette, so a caller can get a sprite's real image, not just its dimensions; validated against real, unmodified MUGEN/Ikemen character files rather than only hand-built test data, which surfaced and fixed two long-standing decoding inaccuracies for real files — a sprite whose stored "linked sprite" reference points at itself or a later sprite now correctly falls back to its own image instead of misreading unrelated data, and a sprite's own color palette is now located correctly for every real file layout, not just the common case; a sprite with a corrupted or nonsensical declared size now falls back to a blank placeholder image instead of risking a crash
- A character (name, animations, sprites, combat logic) can now be loaded directly in a web browser, no local Go installation required: tagging a new release automatically publishes a downloadable WebAssembly build of the library alongside the small glue file it needs to run

[Unreleased]: https://github.com/openkakutou/character/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/openkakutou/character/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/openkakutou/character/releases/tag/v0.1.0
