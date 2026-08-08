//go:build js && wasm

// Command wasm is the WASM entrypoint for the character library: thin
// syscall/js glue exposing character.LoadBytes to a browser (or any JS
// host) as a single global function, so a consumer can load a MUGEN/Ikemen
// character without a Go toolchain of its own.
//
// It carries no logic beyond argument conversion, calling LoadBytes, and
// marshaling the result to JSON — all real behavior lives in the root
// character package, which is unit-tested independently of this file (see
// .vibe/decisions/019-wasm-entrypoint-byte-buffer-loading-and-json-contract.md).
// This file's own behavior is instead verified by smoke.mjs, a Node.js
// script that loads the built module the way a real JS consumer would —
// syscall/js code cannot run under the plain `go test` toolchain.
package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	character "github.com/openkakutou/character"
)

func main() {
	globalName := "OpenKakutouCharacter"
	js.Global().Set(globalName, js.ValueOf(map[string]any{
		"load": js.FuncOf(load),
	}))

	// Registering js.FuncOf callbacks does not keep the Go runtime alive on
	// its own; block forever so OpenKakutouCharacter.load keeps working for
	// the lifetime of the page.
	select {}
}

// load is OpenKakutouCharacter.load(defBytes, airBytes, sffBytes, cnsBytes)
// as seen from JS: each argument is a Uint8Array (or any JS value
// js.CopyBytesToGo accepts), holding that file's raw bytes. It always
// returns a JS object shaped { character: string|null, error: string|null }
// — exactly one of the two fields is non-null — never throws and never lets
// an internal panic escape to the JS caller.
func load(this js.Value, args []js.Value) any {
	defer func() {
		// A panic here would otherwise propagate out of the js.Func
		// callback and tear down the whole page's WASM instance; recover
		// is this boundary's own responsibility, not LoadBytes's (which
		// already returns descriptive errors for every malformed-input
		// path it knows about).
		recover()
	}()

	if len(args) != 4 {
		return result(nil, fmt.Errorf("OpenKakutouCharacter.load: expected 4 arguments (defBytes, airBytes, sffBytes, cnsBytes), got %d", len(args)))
	}

	defBytes, err := bytesFromJS(args[0])
	if err != nil {
		return result(nil, fmt.Errorf("OpenKakutouCharacter.load: defBytes: %w", err))
	}
	airBytes, err := bytesFromJS(args[1])
	if err != nil {
		return result(nil, fmt.Errorf("OpenKakutouCharacter.load: airBytes: %w", err))
	}
	sffBytes, err := bytesFromJS(args[2])
	if err != nil {
		return result(nil, fmt.Errorf("OpenKakutouCharacter.load: sffBytes: %w", err))
	}
	cnsBytes, err := bytesFromJS(args[3])
	if err != nil {
		return result(nil, fmt.Errorf("OpenKakutouCharacter.load: cnsBytes: %w", err))
	}

	c, err := character.LoadBytes(defBytes, airBytes, sffBytes, cnsBytes)
	if err != nil {
		return result(nil, err)
	}

	data, err := json.Marshal(c)
	if err != nil {
		return result(nil, fmt.Errorf("OpenKakutouCharacter.load: encoding result as JSON: %w", err))
	}

	return result(data, nil)
}

// bytesFromJS copies a JS Uint8Array-like value into a Go []byte via
// js.CopyBytesToGo, the standard syscall/js conversion. It returns a
// descriptive error instead of panicking if v is not a byte-array-like
// value (e.g. undefined, or missing a numeric "length").
func bytesFromJS(v js.Value) (b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			b, err = nil, fmt.Errorf("expected a byte array, got %v (%v)", v, r)
		}
	}()

	length := v.Get("length").Int()
	buf := make([]byte, length)
	js.CopyBytesToGo(buf, v)
	return buf, nil
}

// result builds this module's { character, error } JS return shape.
// Exactly one field is ever non-null.
func result(characterJSON []byte, err error) map[string]any {
	if err != nil {
		return map[string]any{"character": nil, "error": err.Error()}
	}
	return map[string]any{"character": string(characterJSON), "error": nil}
}
