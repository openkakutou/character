---
date: 2026-08-11
status: accepted
---
# WASM save path: per-format exports, byte-exact when unmodified, fresh-serialize when edited

**Context:** Item 039 asks for a WASM save/serialize entrypoint accepting an
edited in-memory character representation and returning serialized bytes
for `.def`/`.air`/`.cns`/`.cmd`/`.zss`, with two acceptance criteria that
pull in different directions: (1) it must accept and reflect an *edited*
representation, and (2) an unmodified load→save round trip must be
byte-identical to the original, "matching the existing Go-level guarantee".
The existing Go-level guarantee for byte-identical round trips is
`Document`/`ParseDocument` (per format), which explicitly does **not**
regenerate output around edits — mutating its decoded field has no effect
on `Serialize`, by design (see decisions 002/003/009/012 and the `zss`
package doc comment). The other existing write path, each format's plain
`Serialize(w, model)`, does reflect the model's current values but is
explicitly *not* byte-exact (a "first-pass write path", per `cmd`/`zss`
`Serialize`'s own doc comments). Also, `Character` (the struct `load`
already exposes) has no field for `.cmd`/`.zss` data — those two formats
were never wired into the read path, only into their own packages.

**Decision:** Expose five independent WASM save functions, one per format
(`saveDef`, `saveAir`, `saveCns`, `saveCmd`, `saveZss`), each taking that
format's original file bytes (empty when creating a new file) plus the
edited in-memory model as JSON, rather than a single save call taking one
unified "edited Character". Each implementation: parses the original bytes
into a `Document` (skipped entirely when original bytes are empty); if the
edited model, once normalized (nil slices/maps → empty, matching the same
normalization `LoadBytes` already applies for its own JSON contract),
deep-equals the freshly-parsed baseline (normalized the same way) — writes
the `Document`'s retained source back out verbatim (byte-exact); otherwise
falls back to the format's own `Serialize(w, edited)` (reflects the edit,
not byte-exact). The normalization step exists specifically so a caller
that round-trips an untouched model through JSON (e.g. `null` becoming `[]`
on the JS side) is never spuriously treated as "edited".

**Reason:** Reuses the two write paths this repo already has instead of
inventing a third; each keeps doing exactly the one job it was designed
for. Byte-exact preservation is honored for the specific case the
acceptance criteria actually describe (no edits) without silently
resurrecting the "regenerate around edits while preserving unrelated
formatting" problem every existing `Document` type explicitly declines to
solve. Per-format exports avoid extending `Character`/`LoadBytes`'s JSON
contract with `.cmd`/`.zss` fields it was never designed to carry — wiring
those into the read path is a separate concern with its own scope (e.g. a
character uses `.cns` *or* `.zss`, never both, which a shared `Character`
field would need to model), not implied by this item's own title
("write/serialize path"), and every consumer of the write path
(`character-editor`'s per-format editors, items 007/008) already works
against one format's model at a time.

**Rejected alternatives:**
- **Extend `Character`/`LoadBytes` with `Command`/`Script` fields and add one
  unified `save(editedCharacterJSON)` call**: rejected — conflates a
  read-path scope change (wiring `.cmd`/`.zss` into `Character`, including
  the "never both `.cns` and `.zss`" invariant) with this item's own
  write-path scope, and none of the acceptance criteria ask for it.
- **Always use the format's plain `Serialize(w, model)`, drop the
  byte-exact case entirely**: rejected — directly fails the "byte-identical
  on no edits" acceptance criterion, which explicitly calls out the
  existing Go-level guarantee as the bar to match.
- **Regenerate `Document` output around an edited model (diff-and-patch the
  retained source)**: rejected as out of scope — every existing `Document`
  type already documents this as a heavier per-line reconciliation it does
  not attempt; solving it here, as a side effect of one WASM item, would
  contradict that established, repeated design stance without its own
  dedicated design pass.
