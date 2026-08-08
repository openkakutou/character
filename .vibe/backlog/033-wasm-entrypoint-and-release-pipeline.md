---
status: todo
---
# WASM Entrypoint And Release Pipeline

## Description
The repo's own design constraint states this library "must compile cleanly to WASM," and the README lists WebAssembly compilation under *Planned* — but no WASM entrypoint, build, or release artifact exists yet. A downstream web consumer, `character-viewer-web` (a static page for inspecting a character's sprites, palettes, animations, and specials), needs a pre-built `.wasm` it can download and load directly, without requiring its own Go toolchain. The future `editor` project (per the OpenKakutou roadmap) will need the same artifact.

Add a minimal `cmd/wasm/main.go` entrypoint that exposes JS-callable bindings over the existing read-path API (e.g. load a character from `.def`/`.sff`/`.air`/`.cns` bytes and return its data in a JS/JSON-friendly shape) — glue code only, no rendering, consistent with the "no rendering dependency" design constraint. Then add a GitHub Actions workflow that builds it with `GOOS=js GOARCH=wasm go build` on tagged releases and attaches the resulting `.wasm` plus the matching `wasm_exec.js` glue file as release assets, so consumers can download a version-pinned artifact instead of building from source.

## Acceptance Criteria
- [ ] `cmd/wasm/main.go` builds successfully with `GOOS=js GOARCH=wasm go build`
- [ ] The WASM module exposes JS-callable functions to load a character (from `.def` + its referenced files) and read back its animations/sprites/metadata, without any canvas/OpenGL/rendering code
- [ ] A tagged release triggers a GitHub Actions workflow that builds the wasm and publishes `.wasm` + `wasm_exec.js` as downloadable release assets
- [ ] A consumer can download a specific version's release assets and load the module in a browser (documented with a minimal usage snippet in the README or `docs/`)

## Notes
Raised from `character-viewer-web` (sibling repo, static TS/Vite viewer for character files) during its own `/vibe:init`: it wants to download a pinned WASM build rather than vendor Go source or require a local Go toolchain. Until this item ships, `character-viewer-web` falls back to documenting a local build via a relative path to this repo.
