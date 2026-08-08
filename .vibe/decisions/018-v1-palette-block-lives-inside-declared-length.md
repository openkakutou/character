---
date: 2026-08-08
status: accepted
---
# A v1 sprite's declared Length includes its own trailing palette block, not just its pixel data

**Context:** While building item 028's fixture-driven pixel-comparison tests, comparing decoded output
against real, unmodified `.sff` v1 character files (not the trimmed fixtures already vendored in
`sff/testdata/files`) surfaced that `ResolveV1Palette`'s resolved colors did not match the reference
decoder's output even for a sprite that plainly owns its own palette (`SharedPalette == false`), with no
linking or sharing involved at all.

Direct inspection of a real file (`cvsryu-v1.sff`, sprite table index 2: `Offset=16749`, `Length=4574`,
`SharedPalette=false`) showed `Offset + Length` (21323) lands *exactly* on the next sprite's subheader —
there is no 768-byte gap between them for `ResolveV1Palette`'s prior assumption (palette starts right after
`Offset+Length`) to read from. The sprite's real palette bytes are actually the *last* 768 bytes of its own
`[Offset, Offset+Length)` span (verified byte-for-byte: index 251 of that span decodes to the RGB triplet
the reference PNG's corresponding pixel actually uses). In other words: **when a v1 sprite owns its palette,
`Length` is pixel-data-length-plus-768, not pixel-data-length alone** — `V1SpriteEntry.Length`'s prior doc
comment ("the pixel data length") was incomplete for this case.

ADR 014 (`014-palette-resolution-api-shape.md`) stated this same relationship as already "empirically
confirmed against the real fixtures in `sff/testdata/files`" — but that check only ran against the
*trimmed* fixtures item 023 vendored, which are re-encoded by this repo's own `testdata/gen` tool using a
matching-but-wrong assumption (`encodeV1` wrote `Length = len(pixel)` and appended palette bytes
separately, after it) — a case of the test fixture and the code under test sharing the same incorrect
premise, so a mismatch against genuine on-disk files never surfaced until now.

**Decision:** `ResolveV1Palette` (`sff/palette.go`) now reads an owning sprite's embedded palette block from
the last `v1PaletteBlockSize` (768) bytes of its own `[Offset, Offset+Length)` span. `resolveV1Pixels`
(`sff/load.go`) correspondingly treats the real pixel-only byte count as `Length - 768` for a sprite that
owns its palette (`SharedPalette == false`), and as `Length` unchanged for one that shares (no palette bytes
embedded to subtract). `testdata/gen/main.go`'s `encodeV1` (the trimmed-fixture-writing tool, not part of
the package's own public API) is corrected to match — writing `Length = len(pixel) + len(palette)` for a
non-shared entry — and every v1 fixture under `sff/testdata/files` is regenerated with it, so the vendored
fixtures now encode a sprite's own palette exactly the way genuine files do. `sff/testdata/gen`'s own
*reading* helpers (`v1SubfileBounds`, `findOwnPalette`) needed no change — they already located a real
source file's trailing palette via "the 768 bytes ending at the next sprite's subheader", which is the
correct convention; only their *output*-writing counterpart (`encodeV1`) had the bug.

This does not touch the v1 write path's public API (`SerializeV1`/`V1WriteSprite`): it does not yet support
embedding a sprite's own palette at all (see
`.vibe/decisions/005-sff-v1-serialize-is-semantic-not-byte-exact-round-trip.md`), so no existing
`SerializeV1`/`Load` round-trip test exercised real palette *color* content — only `resolveV1Pixels`'s
derived width/height and `loadV1`'s derived numeric `Sprite.Palette` index, both unaffected by where the
literal RGB bytes live on disk.

**Reason:** `ResolveV1Palette`'s whole purpose is producing the same final colors the reference decoder
does (per ADR 014/015/016's established parity goal); a location bug defeats that regardless of how
correct the surrounding linking/inheritance logic (ADR 017) is.

**Rejected alternatives:**
- *Leave `ResolveV1Palette` as-is and only regenerate the trimmed fixtures to match its (wrong) assumption.*
  Rejected — the fixtures would then no longer reflect genuine `.sff` v1 file layout, defeating the point of
  fixture-driven testing against real data, and would misbehave the moment a real, non-trimmed file is
  loaded through this same function (as this investigation's own real-file check demonstrated).
- *Mark ADR 014 as superseded.* Not done — the API shape ADR 014 actually decided (separate
  `ResolveV1Palette`/`DecodeV1Palette`, `AlphaRule`, etc.) remains correct; only its "Reason" section's
  empirical byte-layout claim was based on a self-consistently-wrong fixture and is corrected here instead.
