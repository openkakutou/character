---
status: todo
---
# `.air` edits patch the lines `Parse` read

## Description
Found by the audit of 2026-11-21 (high, reproduced). `indexSource` (`air/edit.go:115-147`) does not classify lines as `Parse` does:
- it counts a `Copy Action N` line as a frame and keeps a span for an action `resolveCopyActions` dropped: with an unresolvable copy, every edited save fails "document index out of sync" (line 231); with a one-frame copy target, the edit patches `Copy Action 6` into `Copy Action 6, , 10`;
- it counts every non-trivia line as a frame and matches `loopstart` exactly (Ikemen GO matches the prefix, anim.go:395): the counts mismatch, so `patchAction` regenerates the whole action and drops its comments and `Interpolate` lines, which change behaviour in Ikemen GO;
- the in-place Clsn rewrite assumes box i is on line `decl+1+i` (`air/edit.go:562`), while `readClsnBoxes` places boxes by their `[n]` index: with out-of-order or duplicate indices, the wrong box is rewritten.

🔒 MUST: byte-exact round trip; MUGEN 1.0/1.1 and Ikemen GO content (roadmap decision `038`).

## Acceptance Criteria
- [ ] `Parse` and `indexSource` share one line classifier
- [ ] A file with a valid, a dropped and a one-frame `Copy Action` saves unchanged byte-exact, and an edit elsewhere leaves those blocks verbatim
- [ ] Editing one frame of an action with `Loopstart:`, `Interpolate` and comments keeps them
- [ ] Editing a box of a Clsn block with out-of-order indices changes that box only
