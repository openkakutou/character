# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Defined the animation and frame data model (`Animation`, `Frame`, `ClsnBox`) that will represent MUGEN/Ikemen GO character animations once `.air` file reading is implemented
- `.air` animation files can now be read into animation data: actions, frame sequences, collision boxes, and loop points are all parsed from the file's text format
- `.air` reading now handles malformed and unusual input gracefully: comment lines are ignored, malformed action headers and frame lines report a clear error naming the offending line, negative sprite group/image indices are rejected, and an empty file returns an empty result instead of an error
