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

// --- save*: write/serialize path (item 039) ---

// saveDef: no-edits round trip is byte-exact.
const editedInfo = { ...character, palettes: [], stateFiles: [] };
const saveDefNoEditsResult = globalThis.OpenKakutouCharacter.saveDef(defBytes, JSON.stringify(editedInfo));
assert(saveDefNoEditsResult.error === null, `saveDef (no edits) reports no error (got: ${saveDefNoEditsResult.error})`);
assert(saveDefNoEditsResult.bytes instanceof Uint8Array, "saveDef (no edits) returns a Uint8Array");
assert(
	new TextDecoder().decode(saveDefNoEditsResult.bytes) === new TextDecoder().decode(defBytes),
	"saveDef (no edits) is byte-identical to the original .def",
);

// saveDef: an edit is reflected in the output.
const renamedInfo = { ...editedInfo, name: "Renamed via WASM" };
const saveDefEditedResult = globalThis.OpenKakutouCharacter.saveDef(defBytes, JSON.stringify(renamedInfo));
assert(saveDefEditedResult.error === null, `saveDef (edited) reports no error (got: ${saveDefEditedResult.error})`);
const savedDefText = new TextDecoder().decode(saveDefEditedResult.bytes);
assert(savedDefText.includes("Renamed via WASM"), "saveDef (edited) output contains the edited name");

// saveDef: malformed JSON returns a descriptive error instead of throwing.
const saveDefBadJSONResult = globalThis.OpenKakutouCharacter.saveDef(defBytes, "not json");
assert(saveDefBadJSONResult.bytes === null, "saveDef (malformed JSON) returns null bytes");
assert(typeof saveDefBadJSONResult.error === "string" && saveDefBadJSONResult.error.length > 0, "saveDef (malformed JSON) reports an error");

// saveAir: no-edits round trip is byte-exact, an edit is reflected.
const saveAirNoEditsResult = globalThis.OpenKakutouCharacter.saveAir(airBytes, JSON.stringify(character.animations));
assert(saveAirNoEditsResult.error === null, `saveAir (no edits) reports no error (got: ${saveAirNoEditsResult.error})`);
assert(
	new TextDecoder().decode(saveAirNoEditsResult.bytes) === new TextDecoder().decode(airBytes),
	"saveAir (no edits) is byte-identical to the original .air",
);

const editedAnimations = JSON.parse(JSON.stringify(character.animations));
assert(editedAnimations[0].loopStart !== 0, `test fixture sanity: original loopStart isn't already 0 (got: ${editedAnimations[0].loopStart})`);
editedAnimations[0].loopStart = 0;
const saveAirEditedResult = globalThis.OpenKakutouCharacter.saveAir(airBytes, JSON.stringify(editedAnimations));
assert(saveAirEditedResult.error === null, `saveAir (edited) reports no error (got: ${saveAirEditedResult.error})`);
assert(
	new TextDecoder().decode(saveAirEditedResult.bytes) !== new TextDecoder().decode(airBytes),
	"saveAir (edited) output differs from the original",
);

// saveCns: no-edits round trip is byte-exact, an edit is reflected.
const saveCnsNoEditsResult = globalThis.OpenKakutouCharacter.saveCns(cnsBytes, JSON.stringify(character.stateDefs));
assert(saveCnsNoEditsResult.error === null, `saveCns (no edits) reports no error (got: ${saveCnsNoEditsResult.error})`);
assert(
	new TextDecoder().decode(saveCnsNoEditsResult.bytes) === new TextDecoder().decode(cnsBytes),
	"saveCns (no edits) is byte-identical to the original .cns",
);

const editedStateDefs = JSON.parse(JSON.stringify(character.stateDefs));
editedStateDefs[0].ctrl = !editedStateDefs[0].ctrl;
const saveCnsEditedResult = globalThis.OpenKakutouCharacter.saveCns(cnsBytes, JSON.stringify(editedStateDefs));
assert(saveCnsEditedResult.error === null, `saveCns (edited) reports no error (got: ${saveCnsEditedResult.error})`);
assert(
	new TextDecoder().decode(saveCnsEditedResult.bytes) !== new TextDecoder().decode(cnsBytes),
	"saveCns (edited) output differs from the original",
);

// --- loadCmd: read/parse path (item 056) ---
// .cmd isn't wired into `load`'s Character/JSON contract either, so this
// exercises cmd bytes directly rather than through `character`.

// loadCmd: nominal, MUGEN-style .cmd (explicit [Statedef -1], [Remap]).
const mugenCmdBytes = new TextEncoder().encode(
	`[Remap]\na = a\nb = b\n\n[Defaults]\ncommand.time = 15\ncommand.buffer.time = 1\n\n[Command]\nname = "QCF_a"\ncommand = ~D, DF, F, a\n\n[Statedef -1]\n\n[State -1, QCF Special]\ntype = ChangeState\nvalue = 1000\ntrigger1 = command = "QCF_a"\n`,
);
const loadMugenCmdResult = globalThis.OpenKakutouCharacter.loadCmd(mugenCmdBytes);
assert(loadMugenCmdResult.error === null, `loadCmd (MUGEN-style) reports no error (got: ${loadMugenCmdResult.error})`);
assert(typeof loadMugenCmdResult.commandFile === "string", "loadCmd (MUGEN-style) returns a commandFile JSON string");
const mugenCommandFile = JSON.parse(loadMugenCmdResult.commandFile ?? "null");
assert(mugenCommandFile?.remap?.a === "a" && mugenCommandFile?.remap?.b === "b", `loadCmd (MUGEN-style) parses [Remap] (got: ${JSON.stringify(mugenCommandFile?.remap)})`);
assert(mugenCommandFile?.defaults?.time === 15 && mugenCommandFile?.defaults?.bufferTime === 1, `loadCmd (MUGEN-style) parses [Defaults] (got: ${JSON.stringify(mugenCommandFile?.defaults)})`);
assert(Array.isArray(mugenCommandFile?.commands) && mugenCommandFile.commands.length === 1 && mugenCommandFile.commands[0].name === "QCF_a", `loadCmd (MUGEN-style) parses the [Command] section (got: ${JSON.stringify(mugenCommandFile?.commands)})`);
assert(Array.isArray(mugenCommandFile?.states) && mugenCommandFile.states.length === 1 && mugenCommandFile.states[0].number === -1, `loadCmd (MUGEN-style) links the Statedef -1 state (got: ${JSON.stringify(mugenCommandFile?.states)})`);

// loadCmd: Ikemen GO-style .cmd — omits the "[Statedef -1]" header entirely
// (cmd.Parse synthesizes it) and uses charge-input notation.
const ikemenCmdBytes = new TextEncoder().encode(
	`[Command]\nname = "charge_hs"\ncommand = ~$D, /$U, a+b~\ntime = 15\n\n[State -1, Charge Special]\ntype = ChangeState\nvalue = 2000\ntrigger1 = command = "charge_hs"\n`,
);
const loadIkemenCmdResult = globalThis.OpenKakutouCharacter.loadCmd(ikemenCmdBytes);
assert(loadIkemenCmdResult.error === null, `loadCmd (Ikemen-style) reports no error (got: ${loadIkemenCmdResult.error})`);
const ikemenCommandFile = JSON.parse(loadIkemenCmdResult.commandFile ?? "null");
assert(ikemenCommandFile?.commands?.[0]?.input === "~$D, /$U, a+b~", `loadCmd (Ikemen-style) preserves the charge input verbatim (got: ${ikemenCommandFile?.commands?.[0]?.input})`);
assert(Array.isArray(ikemenCommandFile?.states) && ikemenCommandFile.states.length === 1 && ikemenCommandFile.states[0].number === -1, `loadCmd (Ikemen-style) still synthesizes and links the implicit Statedef -1 (got: ${JSON.stringify(ikemenCommandFile?.states)})`);

// loadCmd: malformed .cmd bytes return a descriptive error, not a throw.
const loadCmdMalformedResult = globalThis.OpenKakutouCharacter.loadCmd(new TextEncoder().encode("[Command\nname = \"a\"\n"));
assert(loadCmdMalformedResult.commandFile === null, "loadCmd (malformed) returns null commandFile");
assert(typeof loadCmdMalformedResult.error === "string" && loadCmdMalformedResult.error.length > 0, `loadCmd (malformed) reports a descriptive error (got: ${loadCmdMalformedResult.error})`);

// loadCmd: wrong argument count, must not crash the module.
const loadCmdArgCountResult = globalThis.OpenKakutouCharacter.loadCmd();
assert(loadCmdArgCountResult.commandFile === null, "loadCmd (missing argument) returns null commandFile");
assert(typeof loadCmdArgCountResult.error === "string" && loadCmdArgCountResult.error.length > 0, "loadCmd (missing argument) reports an error");

// loadCmd -> edit -> saveCmd round trip: the edited field persists, the
// rest is unchanged.
const roundTripEdited = JSON.parse(JSON.stringify(mugenCommandFile));
roundTripEdited.commands[0].name = "QCF_a_renamed";
const roundTripSaveResult = globalThis.OpenKakutouCharacter.saveCmd(mugenCmdBytes, JSON.stringify(roundTripEdited));
assert(roundTripSaveResult.error === null, `loadCmd -> saveCmd round trip reports no error (got: ${roundTripSaveResult.error})`);
const roundTripReloadResult = globalThis.OpenKakutouCharacter.loadCmd(roundTripSaveResult.bytes);
assert(roundTripReloadResult.error === null, `loadCmd -> saveCmd -> loadCmd round trip reports no error (got: ${roundTripReloadResult.error})`);
const roundTrippedCommandFile = JSON.parse(roundTripReloadResult.commandFile ?? "null");
assert(roundTrippedCommandFile?.commands?.[0]?.name === "QCF_a_renamed", `round trip: edited command name persists (got: ${roundTrippedCommandFile?.commands?.[0]?.name})`);
assert(roundTrippedCommandFile?.remap?.a === "a", "round trip: unedited fields (remap) are unchanged");

// The module must still respond correctly after loadCmd errors too.
const afterLoadCmdErrorResult = globalThis.OpenKakutouCharacter.load(defBytes, airBytes, sffBytes, cnsBytes);
assert(afterLoadCmdErrorResult.error === null, "module still works after a prior loadCmd error");

// saveCmd: not yet wired into `load`'s Character/JSON contract (.cmd isn't
// part of it), so the baseline comes from a fresh, empty original — a new
// command file — rather than round-tripping character.commandFile.
const newCommandFile = { remap: {}, defaults: { time: 0, bufferTime: 0 }, commands: [{ name: "QCF_a", input: "~D, DF, F, a", time: 0, bufferTime: 0 }], states: [] };
const saveCmdNewResult = globalThis.OpenKakutouCharacter.saveCmd(new Uint8Array(0), JSON.stringify(newCommandFile));
assert(saveCmdNewResult.error === null, `saveCmd (new file) reports no error (got: ${saveCmdNewResult.error})`);
assert(new TextDecoder().decode(saveCmdNewResult.bytes).includes("QCF_a"), "saveCmd (new file) output contains the new command's name");

// saveZss: same "not wired into Character" reasoning as saveCmd.
const zssBytes = new TextEncoder().encode("[Statedef 200; type: S; ctrl: 0;]\nif AnimElem = 1 {\n\tcallSuper{}\n}\n");
const parsedZssResult = { preamble: "", blocks: [{ kind: "Statedef", number: 200, headerParams: { type: "S", ctrl: "0" }, body: "if AnimElem = 1 {\n\tcallSuper{}\n}\n" }] };
const saveZssNoEditsResult = globalThis.OpenKakutouCharacter.saveZss(zssBytes, JSON.stringify(parsedZssResult));
assert(saveZssNoEditsResult.error === null, `saveZss (no edits) reports no error (got: ${saveZssNoEditsResult.error})`);
assert(
	new TextDecoder().decode(saveZssNoEditsResult.bytes) === new TextDecoder().decode(zssBytes),
	"saveZss (no edits) is byte-identical to the original .zss",
);

const editedZss = JSON.parse(JSON.stringify(parsedZssResult));
editedZss.blocks[0].headerParams.ctrl = "1";
const saveZssEditedResult = globalThis.OpenKakutouCharacter.saveZss(zssBytes, JSON.stringify(editedZss));
assert(saveZssEditedResult.error === null, `saveZss (edited) reports no error (got: ${saveZssEditedResult.error})`);
assert(
	new TextDecoder().decode(saveZssEditedResult.bytes) !== new TextDecoder().decode(zssBytes),
	"saveZss (edited) output differs from the original",
);

// saveCns: wrong argument count, must not crash the module.
const saveArgCountResult = globalThis.OpenKakutouCharacter.saveCns(cnsBytes);
assert(saveArgCountResult.bytes === null, "saveCns (missing argument) returns null bytes");
assert(typeof saveArgCountResult.error === "string" && saveArgCountResult.error.length > 0, "saveCns (missing argument) reports an error");

// The module must still respond correctly after the save* calls above too.
const afterSaveResult = globalThis.OpenKakutouCharacter.load(defBytes, airBytes, sffBytes, cnsBytes);
assert(afterSaveResult.error === null, "module still works after the save* calls");

if (process.exitCode) {
	console.error("\nsmoke test FAILED");
} else {
	console.log("\nsmoke test passed");
}
process.exit(process.exitCode ?? 0);
