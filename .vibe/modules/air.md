# Module: air

**Role:** Read path for MUGEN/Ikemen GO animation (`.air`) files: the pure-data vocabulary a library consumer (editor, engine) works with (`animation.go`), plus a parser that turns `.air` text into that model (`parser.go`). The parser covers the format's happy path and hardens against malformed/unusual input — comment lines (`;`, whole-line or trailing) are ignored, an empty file yields an explicit empty result, and a malformed action header, a frame line with missing/non-numeric/negative group-or-image fields, or a reader failure returns a descriptive, line-numbered error rather than panicking or producing incorrect data. `.air` serialization (write path) is a separate, not-yet-implemented backlog item.
**Files:** `air/animation.go`, `air/parser.go`
**Exports:** `Animation` (struct), `Frame` (struct), `ClsnBox` (struct), `Flip` (string type) with constants `FlipNone`, `FlipH`, `FlipV`, `FlipHV`, `BlendMode` (string type), `Parse(r io.Reader) ([]Animation, error)`
**Depends on:** none
