---
status: in_progress
---
# Migrate To Depend On Standalone `sff` Repo

## Description
This repo's `sff/` package is being extracted into a new standalone repo, `sff` (see `roadmap`'s `.vibe/decisions/007`), because `stage` and `lifebar` also need sprite parsing and shouldn't depend on this repo to get it. This item removes the internal `sff/` package from this repo and replaces it with a dependency on the external `github.com/openkakutou/sff` Go module, updating every internal reference (`character.go`'s `ResolveSprite` delegation, `load.go`, `load_bytes.go`, `cmd/wasm/main.go`'s sprite-related WASM exports) to use the external package instead.

## Acceptance Criteria
- [ ] `sff/` package removed from this repo; `go.mod` depends on `github.com/openkakutou/sff` instead
- [ ] All existing behavior (JSON contract, WASM exports) is unchanged from a consumer's point of view — no breaking change to `character-viewer-web` or the published WASM API
- [ ] All existing tests that previously exercised the internal `sff` package now pass against the external module without modification to their assertions
- [ ] A malformed/incompatible `sff` module version is caught at build time (`go.mod` version pin), not silently

## Notes
Blocked on the `sff` repo publishing an initial tagged release with functionality equivalent to this repo's current internal package (cross-repo blocker, not expressible via `depends_on` frontmatter since that only tracks same-repo item numbers). See `roadmap`'s `.vibe/decisions/007`.
