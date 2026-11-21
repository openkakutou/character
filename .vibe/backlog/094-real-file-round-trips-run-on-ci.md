---
status: todo
depends_on: [065]
---
# Real-file round trips run on CI

## Description
Found by the audit of 2026-11-21 (medium). The only real-file round-trip tests depend on the environment and skip silently (`corpus_test.go:26`, `zss/document_test.go:114`, which falls back to a path on the author's machine). The project's rule is validation against real community files, yet CI never runs one. Tests that only assert Go struct semantics also inflate the count (`def/character_info_test.go`, `air/animation_test.go:5-58`, `cmd/command_test.go`, `cns/statedef_test.go`, `character_test.go`).

## Acceptance Criteria
- [ ] A few small real-character files whose licence allows it are committed under `testdata/` with byte-exact and single-edit assertions in the default suite
- [ ] The external-corpus tests can be made to fail instead of skip with an environment flag
- [ ] The struct-semantics-only tests are removed
