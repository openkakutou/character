---
status: todo
---
# Expose every state file in `Load`

## Description
`Load` parses only the `.def` `cns` entry into `StateDefs`; `st`, `stN` and `stcommon` are left to each consumer, so the engine's conformance harness parses them itself (engine ADR 019) and `game` would have to again. Load them in Ikemen GO's order with its duplicate rule (first definition wins, `src/compiler.go`), resolving `stcommon` exactly as Ikemen GO searches it (character folder, then the engine's own data folder; cite `src/char.go`/`src/compiler.go`), where the engine's data folder is supplied by the caller (`game`'s `common/`, roadmap decision `031`; the engine conformance harness, which then loads Ikemen GO's bundled `kfm`).

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO compatibility, verified against Ikemen GO (roadmap decision `038`).

## Acceptance Criteria
- [ ] `Character` exposes the merged states with their source file
- [ ] Duplicate and order rules match Ikemen GO, cited
- [ ] `stcommon` falls back to a caller-given folder

## Notes
Found by the conformance harness (item 030, `docs/conformance.md`, 2026-11-17).
