---
date: 2026-08-09
status: accepted
---
# v2 zero-length sprite pixel/palette resolution always copies table index 0

**Context:** Item 029 required `ResolveSpritePixels` to resolve a v2 sprite
whose table entry declares `Length == 0` (no pixel data of its own) — the
"zero-length copy" scenario ported from the reference JS project's own test
suite (`extractSpritesFromSFFV2.mjs`, fixture `piccolo-v2.sff` group 186).
v1 already has an equivalent case, resolved positionally: a zero-length v1
sprite always inherits the immediately preceding table entry's image (see
`.vibe/decisions/017-v1-sprite-linking-and-palette-inheritance-rules.md`).

**Decision:** A v2 sprite with `Length == 0` (and, per the reference
project's own logic, only ever observed with `LinkedIndex == 0`) resolves
its pixel data, format, color depth, and palette bank index from the
sprite table's own index 0 entry — not from the immediately preceding
table entry, and not by following its own `LinkedIndex` field. This
reproduces the reference project's `extractSpritesFromSFFV2.mjs` exactly:
its zero-length branch spreads `{...sprites[0], index, group, number, x,
y}`, literally copying the first entry of the full (unfiltered) sprite
table regardless of the target's own position or declared `LinkedIndex`.

A v2 sprite with `Length == 0` and `LinkedIndex > 0` remains unsupported
(still reports a descriptive error) — the reference project itself does
not resolve this case within its own sprite-extraction pass (it patches it
in afterwards, indexing into the *filtered* output array by a *raw table*
index, a mismatch that only works by accident when no group filter is
applied); no fixture in this repo's corpus exercises it, so it is left as
a distinct, explicit future gap rather than guessed at.

**Reason:** Positional inheritance (v1's rule) and "always copy table
index 0" (v2's rule, per this decision) are both real, independently
confirmed behaviors of their respective formats/reference implementations
— not a single unified "zero-length" rule. Matching the reference
project's actual code (verified by running its own test suite against the
real, unmodified source files) was preferred over assuming the two
formats share one resolution rule.

Separately, this item also confirmed — by decoding the real, untrimmed
`kazuki-v2.sff` end to end and diffing every pixel against the reference
project's own expected PNG (zero mismatches across 69,834 pixels) — that
`ParseV2`'s existing offset-flag branch (`literalDataOffset` vs. the field
it reads at header byte 60, labeled `translatedDataOffset`) already
computes the exact same absolute offset the reference project's own
`loadMode === 1` branch does (`onDemandDataSizeTotal + dataOffset` — the
same 4 header bytes, just named differently by each implementation). No
`ParseV2` change was needed for on-demand (`loadMode = 1`) addressing,
despite the backlog item and `testdata/README.md`'s "known caveat"
anticipating one might be.

**Rejected alternatives:** Extending v1's positional-inheritance rule to
v2 as well — rejected because it does not match the real reference
behavior (confirmed against `piccolo-v2.sff`'s actual table layout, where
the donor is table index 0, which is also the immediately preceding entry
in that specific fixture, so the two rules happen to coincide there and
would not have been distinguishable without checking the reference
project's own source).
