#!/usr/bin/env node
// smoke.mjs is a Node.js verification harness for the WASM entrypoint built
// from this directory (see main.go) — it exercises the module the same way
// a browser consumer would (fetch/instantiate the .wasm, call the exposed
// global function, read back the result), without requiring an actual
// browser. It is not part of `go test` — syscall/js glue cannot run under
// the plain Go toolchain — and doubles as a minimal usage example for a JS
// consumer.
//
// Usage: node cmd/wasm/smoke.mjs [path/to/character.wasm]
// (defaults to ./character.wasm, relative to the repo root)

import { execSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..", "..");
const wasmPath = path.resolve(process.argv[2] || path.join(repoRoot, "character.wasm"));

const goroot = execSync("go env GOROOT").toString().trim();
const wasmExecPath = path.join(goroot, "lib", "wasm", "wasm_exec.js");

// wasm_exec.js defines a global `Go` constructor; importing it for its
// side effect is the same pattern used to load it in a browser <script> tag.
await import(`file://${wasmExecPath}`);

const go = new globalThis.Go();
const { instance } = await WebAssembly.instantiate(readFileSync(wasmPath), go.importObject);
go.run(instance); // does not return: keeps the Go runtime (and its registered functions) alive

function toUint8Array(relativePath) {
	return new Uint8Array(readFileSync(path.join(repoRoot, relativePath)));
}

function assert(condition, message) {
	if (!condition) {
		console.error(`FAIL: ${message}`);
		process.exitCode = 1;
	} else {
		console.log(`ok - ${message}`);
	}
}

const defBytes = new TextEncoder().encode("[Info]\nname = Smoke Test\nauthor = Someone\n");
const airBytes = toUint8Array("air/testdata/sample.air");
const sffBytes = toUint8Array("sff/testdata/files/v1-basic.sff");
const cnsBytes = toUint8Array("cns/testdata/sample.cns");

// --- nominal path ---
const okResult = globalThis.OpenKakutouCharacter.load(defBytes, airBytes, sffBytes, cnsBytes);
assert(okResult.error === null, `nominal load reports no error (got: ${okResult.error})`);
assert(typeof okResult.character === "string", "nominal load returns a character JSON string");

const character = JSON.parse(okResult.character ?? "null");
assert(character?.name === "Smoke Test", `character name is "Smoke Test" (got: ${character?.name})`);
assert(Array.isArray(character?.animations) && character.animations.length === 2, "character has 2 animations");
assert(Array.isArray(character?.sprites) && character.sprites.length === 1, "character has 1 sprite group");
assert(Array.isArray(character?.stateDefs) && character.stateDefs.length === 3, "character has 3 state defs");
assert(Array.isArray(character?.animations?.[0]?.frames?.[0]?.clsn1), "a frame's empty clsn1 is an array, not null/undefined");

// --- error path: malformed sff bytes ---
const errResult = globalThis.OpenKakutouCharacter.load(defBytes, airBytes, new TextEncoder().encode("garbage"), cnsBytes);
assert(errResult.character === null, "malformed sff bytes: character is null");
assert(typeof errResult.error === "string" && errResult.error.length > 0, `malformed sff bytes: error is a non-empty string (got: ${errResult.error})`);
assert(errResult.error.includes("sprite"), `malformed sff bytes: error identifies the sprite stage (got: ${errResult.error})`);

// --- error path: wrong argument count, must not crash the module ---
const argCountResult = globalThis.OpenKakutouCharacter.load(defBytes, airBytes);
assert(argCountResult.character === null, "missing arguments: character is null");
assert(typeof argCountResult.error === "string" && argCountResult.error.length > 0, "missing arguments: error is a non-empty string");

// The module must still respond correctly after an error — proves the
// earlier failures didn't leave the Go runtime in a broken state.
const afterErrorResult = globalThis.OpenKakutouCharacter.load(defBytes, airBytes, sffBytes, cnsBytes);
assert(afterErrorResult.error === null, "module still works after a prior error");

if (process.exitCode) {
	console.error("\nsmoke test FAILED");
} else {
	console.log("\nsmoke test passed");
}
process.exit(process.exitCode ?? 0);
