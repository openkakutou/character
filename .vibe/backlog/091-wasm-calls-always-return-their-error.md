---
status: todo
---
# WASM calls always return their error object

## Description
Found by the audit of 2026-11-21 (medium). Every exported JS function does `defer func(){ recover() }()` with unnamed results (`cmd/wasm/main.go:79-86` and seven more): after a panic the call returns `undefined`, not the documented `{…, error}`, and the panic message is lost. `saveDef` (line 382) also applies a merge rule ("keys absent from the JSON keep their original value") in the build-tagged glue, where `go test` cannot reach it and a native `SerializeDef` caller does not get it.

## Acceptance Criteria
- [ ] A panic in any call returns `{…, error}` with the panic message
- [ ] The `saveDef` overlay lives in the root package with a test; the glue only converts arguments
