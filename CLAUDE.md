# OpenKakutou — character

## Project context

OpenKakutou is an open-source alternative to Fighter Factory Studio (which is closed-source), the reference editor for creating MUGEN / Ikemen GO fighting game characters.

Long-term scope of the OpenKakutou org (github.com/openkakutou):
- `character` (this repo) — read/write library for MUGEN/Ikemen GO character files
- `editor` — web-based character editor (built on top of `character`)
- `engine` (future) — a game engine in the spirit of Ikemen GO, built on the same extracted libraries

This repo (`character`) is the foundation everything else depends on. It must stay independent from any rendering/graphics backend — Ikemen GO's own source was found to tightly couple file parsing with OpenGL texture generation, which makes it unsuitable to extract from directly. This library should be built from the MUGEN/Ikemen file format specs instead.

## Scope of this repo

Read **and** write support for MUGEN/Ikemen GO character files:
- `.def` — character definition, references the other files
- `.sff` — sprites (binary format)
- `.air` — animations (text format) — **priority for the first implementation**
- `.cns` — combat logic/state machine (text format) — later, not required for the sprite/animation editor

## Design constraints

1. **No rendering dependency.** This module only parses/serializes data — no OpenGL, no canvas, nothing rendering-related. It must compile cleanly to WASM for the web editor, and later be usable as-is by both the web editor and, eventually, a game engine.

2. **Read and write live in the same repo, but as clearly separated sub-concerns** (not necessarily separate packages from day one — separate the concerns internally):
   - **Read path**: should expose a minimal, stable, pure-data API — this is the surface a future game engine would consume. Design it as if the engine were already a real client, not as "whatever's left after removing write."
   - **Write path**: needs to preserve as much of the original file structure as possible on round-trip (ordering, comments) for `.air`/`.cns` — otherwise every save from the editor produces a huge, unreadable Git diff, which hurts community collaboration on character files.
   - Because Go only compiles what's actually imported, a consumer that only imports the read-oriented sub-package won't pull in write-only dependencies (e.g. formatting-preservation logic) — this already gives engine-side isolation without needing a separate repo.

3. **Formats are interdependent, not independent.** `.air` references sprite groups defined in `.sff` — there's essentially no real use case for parsing one without the other. Suggested internal layout:
   ```
   character/
     def/   → .def parsing (entry point, references other files)
     sff/   → sprite parsing (binary)
     air/   → animation parsing (text, depends on sff) — start here
     cns/   → combat logic parsing (text) — later
   ```
   A root `character` package should assemble these into a single `Character{}` struct (sprites + animations + hitboxes) — the unit a library consumer actually wants to work with, not raw per-format structs.

## Stack

- Go, compiled to WebAssembly (`GOOS=js GOARCH=wasm`) for the web editor
- No UI code in this repo — UI (the editor itself) will be a separate repo, likely JS/TS-based, calling into this module via WASM
- Later, the same Go code should be reusable in a desktop build via Wails, without rewriting core logic

## Where to start

Begin with `.air` parsing (`character/air`), since it's the first editor feature (sprite/animation editor). `.sff` support (`character/sff`) will be needed alongside it since animations reference sprites. `.cns` can wait.
