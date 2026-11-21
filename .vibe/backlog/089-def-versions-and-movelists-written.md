---
status: todo
---
# `.def` versions and move lists are written

## Description
Found by the audit of 2026-11-21 (high). `MugenVersion` / `IkemenVersion` are parsed (`def/parser.go:93-96`) but `writeInfoSection` writes only name and author (`def/serializer.go:35-53`), and `scalarFields` / `indexSource` omit them (`def/edit.go:130-139, 83`): a new character saved by the editor has no `ikemenversion`, and an edited version is dropped. Ikemen GO reads `ikemenversion` and forces `mugenversion` 1.1 when it is set (compiler.go:8583-8588), so the character silently runs with legacy behaviours. `cloneInfo` (`def/edit.go:116-120`) does not copy `Movelists`, and `SerializeEdited` treats move lists as read-only while `Serialize` writes them.

🔒 MUST: Ikemen GO behaviour (roadmap decision `038`).

## Acceptance Criteria
- [ ] Both versions are written by `Serialize` and patched by `SerializeEdited`, with round-trip tests
- [ ] Move lists are cloned and written the same way on both paths
