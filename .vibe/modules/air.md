# Module: air

**Role:** Read path for MUGEN/Ikemen GO animation (`.air`) files: the pure-data vocabulary a library consumer (editor, engine) works with (`animation.go`), plus a parser that turns `.air` text into that model (`parser.go`). Covers the format's happy path only — malformed/unusual input handling and `.air` serialization (write path) are separate, not-yet-implemented backlog items.
**Files:** `air/animation.go`, `air/parser.go`
**Exports:** `Animation` (struct), `Frame` (struct), `ClsnBox` (struct), `Flip` (string type) with constants `FlipNone`, `FlipH`, `FlipV`, `FlipHV`, `BlendMode` (string type), `Parse(r io.Reader) ([]Animation, error)`
**Depends on:** none
