---
status: in_progress
---
# Add `.zss` File Parsing (Ikemen GO Lua-Like State Scripts)

## Description
Ikemen GO allows character state logic to be written in `.zss`, a Lua-like scripting format, as an alternative to `.cns`. Add a `zss` package that parses `.zss` files into a structured document model, keeping script bodies and expressions as unevaluated raw text (same principle as `cns`'s triggers) — this repo parses, it does not execute. Execution is `engine`'s responsibility.

## Acceptance Criteria
- [ ] `.zss` files parse into a structured model preserving per-state script bodies as raw text blocks
- [ ] Round-trip serialize preserves original formatting for unmodified content
- [ ] Fixture-driven tests against real Ikemen GO `.zss` files from community characters
- [ ] A malformed `.zss` file returns a descriptive error instead of crashing

## Notes
A character uses either `.cns` or `.zss`, never both — this is purely additive, not a replacement for `cns`. See `roadmap`'s `.vibe/decisions/012`.
