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
const sffBytes = toUint8Array("cmd/wasm/testdata/v1-basic.sff");
const cnsBytes = toUint8Array("cns/testdata/sample.cns");

// --- nominal path ---
const okResult = globalThis.OpenKakutouCharacter.load(defBytes, airBytes, sffBytes, cnsBytes);
assert(okResult.error === null, `nominal load reports no error (got: ${okResult.error})`);
assert(typeof okResult.character === "string", "nominal load returns a character JSON string");

const character = JSON.parse(okResult.character ?? "null");
assert(character?.name === "Smoke Test", `character name is "Smoke Test" (got: ${character?.name})`);
assert(character?.author === "Someone", `character author is "Someone" (got: ${character?.author})`);
assert(Array.isArray(character?.animations) && character.animations.length === 2, "character has 2 animations");
assert(Array.isArray(character?.sprites) && character.sprites.length === 1, "character has 1 sprite group");
assert(Array.isArray(character?.stateDefs) && character.stateDefs.length === 3, "character has 3 state defs");
assert(Array.isArray(character?.animations?.[0]?.frames?.[0]?.clsn1), "a frame's empty clsn1 is an array, not null/undefined");

// --- backlog item 038: the rest of CharacterInfo's fields (unset in this
// fixture's .def, which has no [Files] section) must still be present in
// the JSON contract, empty/zero rather than missing or null ---
assert(character?.spriteFile === "", `unset spriteFile is an empty string (got: ${JSON.stringify(character?.spriteFile)})`);
assert(character?.animationFile === "", `unset animationFile is an empty string (got: ${JSON.stringify(character?.animationFile)})`);
assert(character?.soundFile === "", `unset soundFile is an empty string (got: ${JSON.stringify(character?.soundFile)})`);
assert(character?.commandFile === "", `unset commandFile is an empty string (got: ${JSON.stringify(character?.commandFile)})`);
assert(character?.constantsFile === "", `unset constantsFile is an empty string (got: ${JSON.stringify(character?.constantsFile)})`);
assert(Array.isArray(character?.stateFiles) && character.stateFiles.length === 0, "unset stateFiles is an empty array, not null");
assert(Array.isArray(character?.palettes) && character.palettes.length === 0, "unset palettes is an empty array, not null");

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

// --- resolveSprites: batched sprite pixel resolution (item 034) ---
// v1-basic.sff carries exactly one real sprite, at (group 0, image 0).

// --- nominal batch: one real sprite, one nonexistent (group, image) ---
const spritesResult = globalThis.OpenKakutouCharacter.resolveSprites(sffBytes, [[0, 0], [999, 999]], null);
assert(Array.isArray(spritesResult) && spritesResult.length === 2, "resolveSprites returns one result per request");

const [found, notFound] = spritesResult;
assert(found.error === null, `resolveSprites: real sprite reports no error (got: ${found.error})`);
assert(found.pixels instanceof Uint8Array, "resolveSprites: real sprite returns a pixel buffer");
assert(found.pixels.length === found.width * found.height * 4, "resolveSprites: pixel buffer length is width*height*4 (RGBA)");
assert(found.width > 0 && found.height > 0, `resolveSprites: real sprite has positive dimensions (got: ${found.width}x${found.height})`);

assert(notFound.pixels === null, "resolveSprites: nonexistent sprite returns null pixels");
assert(notFound.width === 0 && notFound.height === 0, "resolveSprites: nonexistent sprite reports 0x0 dimensions");
assert(typeof notFound.error === "string" && notFound.error.startsWith("sprite not found: "), `resolveSprites: nonexistent sprite error is distinguishable (got: ${notFound.error})`);

// --- external palette override recolors the sprite ---
const actBytes = toUint8Array("cmd/wasm/testdata/cyclops-v1-palette1.act");
const overriddenResult = globalThis.OpenKakutouCharacter.resolveSprites(sffBytes, [[0, 0]], actBytes);
assert(overriddenResult[0].error === null, `resolveSprites: override reports no error (got: ${overriddenResult[0].error})`);
const differs = overriddenResult[0].pixels.some((b, i) => b !== found.pixels[i]);
assert(differs, "resolveSprites: external palette override changes the resolved colors");

// --- undefined and null overrideBytes are equivalent to "no override" ---
const undefinedOverrideResult = globalThis.OpenKakutouCharacter.resolveSprites(sffBytes, [[0, 0]], undefined);
const nullOverrideResult = globalThis.OpenKakutouCharacter.resolveSprites(sffBytes, [[0, 0]], null);
assert(
	undefinedOverrideResult[0].pixels.every((b, i) => b === nullOverrideResult[0].pixels[i]),
	"resolveSprites: undefined and null overrideBytes produce identical output",
);
assert(
	nullOverrideResult[0].pixels.every((b, i) => b === found.pixels[i]),
	"resolveSprites: no override matches the sprite's own palette",
);

// --- an explicitly empty overrideBytes is an error, not a silent fallback ---
const emptyOverrideResult = globalThis.OpenKakutouCharacter.resolveSprites(sffBytes, [[0, 0]], new Uint8Array(0));
assert(emptyOverrideResult[0].pixels === null, "resolveSprites: empty overrideBytes returns null pixels");
assert(typeof emptyOverrideResult[0].error === "string" && emptyOverrideResult[0].error.length > 0, "resolveSprites: empty overrideBytes reports an error");

// --- malformed sffBytes: no throw, every request in the batch reports an error ---
const malformedBatchResult = globalThis.OpenKakutouCharacter.resolveSprites(new TextEncoder().encode("garbage"), [[0, 0]], null);
assert(malformedBatchResult[0].pixels === null, "resolveSprites: malformed sffBytes returns null pixels");
assert(typeof malformedBatchResult[0].error === "string" && malformedBatchResult[0].error.length > 0, "resolveSprites: malformed sffBytes reports an error");

// The module must still respond correctly after resolveSprites errors too.
const afterResolveErrorResult = globalThis.OpenKakutouCharacter.load(defBytes, airBytes, sffBytes, cnsBytes);
assert(afterResolveErrorResult.error === null, "module still works after a prior resolveSprites error");

if (process.exitCode) {
	console.error("\nsmoke test FAILED");
} else {
	console.log("\nsmoke test passed");
}
process.exit(process.exitCode ?? 0);
