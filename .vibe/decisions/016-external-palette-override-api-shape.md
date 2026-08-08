---
date: 2026-08-08
status: accepted
---
# External .act palette override API shape (backlog item 027)

**Context:** Adding support for an external `.act` palette (a 768-byte,
256×RGB buffer used by MUGEN/Ikemen to recolor a character without
touching its sprites) and letting it be used in place of a sprite's own
palette when resolving pixel colors — item 026's `ResolveV1Palette`/
`ResolveV2Palette`.

**Decision:**
- New exported function `DecodeExternalPalette(data []byte) (Palette, error)`
  in `sff/palette.go`, alongside `DecodeV1Palette`/`DecodeV2Palette` — same
  "decode raw bytes already in hand" shape, kept separate from any
  table/link-aware resolve step since an external `.act` buffer has none.
  It reproduces `convertExternalPaletteToRGBA.mjs`'s two quirks: the file's
  256 RGB triplets are stored in reverse index order (the first triplet in
  the file becomes palette index 255, the last becomes index 0), and only
  the resulting index 0 is forced to alpha 0 — every other entry stays
  opaque, unlike `DecodeV1Palette`, which is always fully opaque, and
  `DecodeV2Palette`, which always uses the file's own literal alpha.
- `ResolveV1Palette`/`ResolveV2Palette` each gain a new trailing
  `override *Palette` parameter. When non-nil, the function returns
  `*override` immediately, bypassing the table lookup (and, for v1, the
  `io.ReaderAt` read) entirely — "instead of the sprite's own", not merged
  with it. `nil` preserves the exact existing behavior, so every current
  call site only needs `nil` appended.

**Reason:** `override *Palette` mirrors the codebase's existing style for
optional/absent references (`*V1SpriteTable`, `io.ReaderAt` results) rather
than introducing a second family of functions or a variadic/options-struct
parameter — for two functions with a single optional input, that extra
machinery buys nothing. Both call sites live only in this package's own
tests today (no external consumers yet), so extending the signature
in place is a strictly cheaper change than adding parallel
`ResolveV1PaletteWithOverride`-style functions, which would duplicate the
link-walking logic's doc comments and double the surface area for a single
extra parameter.

**Rejected alternatives:**
- A second set of functions (`ResolveV1PaletteWithOverride`, ...) — rejected
  as needless duplication for one optional parameter, and it would leave
  two entry points to keep in sync as the link-walking logic evolves.
- A functional-options parameter (`...Option`) — rejected as
  over-engineered for a single optional value; this package uses plain
  parameters everywhere else (e.g. `AlphaRule`, not an options struct).
- Merging the override *into* the sprite's own palette (e.g. only
  overriding non-zero entries) — rejected: the reference project and the
  backlog item both describe a full replacement, not a partial merge.
