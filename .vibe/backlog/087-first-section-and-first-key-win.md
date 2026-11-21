---
status: todo
---
# First section and first key win, as Ikemen GO reads `.def` and `.cmd`

## Description
Found by the audit of 2026-11-21 (medium).
- `def.Parse` merges every `[Info]` / `[Files]` section, last key wins (`def/parser.go:71`, and `def/edit.go:172`). Ikemen GO reads only the first of each section (char.go:4237-4241, 4273-4277; compiler.go:8580-8603) and keeps a key's first occurrence (common.go:1085-1088): `anim = old.air` then `anim = kfm.air` loads another file than Ikemen GO.
- `cmd` reads only a section named exactly `command` (`cmd/parser.go:171`); Ikemen GO takes any name starting with `command` (compiler.go:8711), reads only the first `[Remap]` and `[Defaults]` (8657, 8696), and keeps a repeated key's first value. `[Command 1]` sections are lost.

🔒 MUST: Ikemen GO behaviour (roadmap decision `038`).

## Acceptance Criteria
- [ ] Repeated sections after the first and repeated keys within a section are ignored, in `.def` and `.cmd`; `stN` / `palN` deduplicated by key
- [ ] `[Command 1]` is read as a command
- [ ] The edit paths patch the line that won
