---
status: todo
---
# `.air` actions compiled as Ikemen GO compiles them

## Description
Found by the audit of 2026-11-21 (medium). The frame-line leniency is half done, and two action rules differ from Ikemen GO:
- `parseFrameLine` drops a frame whose field has no leading digits (`air/parser.go:196-202`); Ikemen GO's `Atoi` reads 0 and keeps it (common.go:191-221, anim.go:58-70), so later frame indexes shift. A Clsn count like `Clsn2: 2 boxes` is not seen as a declaration (`parser.go:312`; Ikemen GO reads 2, anim.go:411). Five error messages are built and discarded; the `Parse` doc (56-61) and decision 024 still promise an error.
- A `Loopstart` after the last frame leaves `LoopStart == len(Frames)` (`parser.go:162`); Ikemen GO restores the loop start in effect at the last frame (anim.go:383, 440-442).
- An action with no valid frame and no copy comes back empty (`parser.go:134`); Ikemen GO's `compileAction` gives it the next action's frames (anim.go:1341-1354).

🔒 MUST: Ikemen GO behaviour (roadmap decision `038`).

## Acceptance Criteria
- [ ] Numeric fields and Clsn counts read as Ikemen GO's `Atoi`; only a line Ikemen GO skips is skipped
- [ ] `LoopStart` is never past the last frame, and loops where Ikemen GO loops
- [ ] An empty action takes the next action's frames
- [ ] The `Parse` doc and a decision superseding 024 describe the rules; the serializer's "LoopStart past the last frame" case is revisited
