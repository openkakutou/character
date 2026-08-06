# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
