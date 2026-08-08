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
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"syscall/js"

	character "github.com/openkakutou/character"
	"github.com/openkakutou/character/sff"
)

func main() {
	globalName := "OpenKakutouCharacter"
	js.Global().Set(globalName, js.ValueOf(map[string]any{
		"load":           js.FuncOf(load),
		"resolveSprites": js.FuncOf(resolveSprites),
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

// resolveSprites is
// OpenKakutouCharacter.resolveSprites(sffBytes, requests, overrideBytes) as
// seen from JS: sffBytes is a Uint8Array holding a .sff file's raw bytes
// (transferred once for the whole call, not once per sprite); requests is
// an array of [group, image] number pairs; overrideBytes is an optional
// Uint8Array holding an external .act palette file — undefined or null
// means "use each sprite's own palette", any other value (including an
// empty Uint8Array) is decoded, and if invalid, reported as an error for
// every request in the batch rather than silently falling back. Returns
// one { pixels, width, height, error } object per request, in the same
// order — pixels is a flat, row-major RGBA byte buffer
// (width*height*4 bytes); on error, pixels/width/height are
// nil/0/0. Like load, this never throws and never leaves the module in a
// broken state after an error. See
// .vibe/decisions/020-wasm-sprite-pixel-resolution-batched-stateless-contract.md.
func resolveSprites(this js.Value, args []js.Value) any {
	defer func() {
		// See load's identical recover() — a panic here would otherwise
		// tear down the whole page's WASM instance.
		recover()
	}()

	if len(args) != 3 {
		return []any{spriteResult(nil, 0, 0, fmt.Errorf("OpenKakutouCharacter.resolveSprites: expected 3 arguments (sffBytes, requests, overrideBytes), got %d", len(args)))}
	}

	sffBytes, err := bytesFromJS(args[0])
	if err != nil {
		return []any{spriteResult(nil, 0, 0, fmt.Errorf("OpenKakutouCharacter.resolveSprites: sffBytes: %w", err))}
	}

	override, overrideErr := overridePaletteFromJS(args[2])

	n := args[1].Get("length").Int()
	r := bytes.NewReader(sffBytes)
	results := make([]any, n)
	for i := 0; i < n; i++ {
		if overrideErr != nil {
			results[i] = spriteResult(nil, 0, 0, fmt.Errorf("OpenKakutouCharacter.resolveSprites: overrideBytes: %w", overrideErr))
			continue
		}
		group, image, err := spriteRequestFromJS(args[1].Index(i))
		if err != nil {
			results[i] = spriteResult(nil, 0, 0, fmt.Errorf("OpenKakutouCharacter.resolveSprites: requests[%d]: %w", i, err))
			continue
		}
		pixels, width, height, err := sff.ResolveSpritePixels(r, group, image, override)
		results[i] = spriteResult(pixels, width, height, err)
	}
	return results
}

// overridePaletteFromJS decodes v as an external .act palette override, or
// returns (nil, nil) when v is JS undefined/null — the two values a
// caller uses to mean "no override, use the sprite's own palette". Any
// other value, including an empty Uint8Array, is decoded via
// sff.DecodeExternalPalette and its own validation (wrong size, etc.)
// surfaces as err, never a silent fallback to no override.
func overridePaletteFromJS(v js.Value) (*sff.Palette, error) {
	if v.IsUndefined() || v.IsNull() {
		return nil, nil
	}
	b, err := bytesFromJS(v)
	if err != nil {
		return nil, err
	}
	p, err := sff.DecodeExternalPalette(b)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// spriteRequestFromJS reads one [group, image] pair from requests[i] as
// seen from JS. Returns a descriptive error instead of panicking if v is
// not a length-2 array-like value.
func spriteRequestFromJS(v js.Value) (group, image int, err error) {
	defer func() {
		if r := recover(); r != nil {
			group, image, err = 0, 0, fmt.Errorf("expected a [group, image] pair, got %v (%v)", v, r)
		}
	}()

	if length := v.Get("length").Int(); length != 2 {
		return 0, 0, fmt.Errorf("expected a [group, image] pair (length 2), got length %d", length)
	}
	return v.Index(0).Int(), v.Index(1).Int(), nil
}

// spriteResult builds this module's { pixels, width, height, error } JS
// return shape for one resolved sprite. On error, pixels/width/height are
// explicitly nil/0/0 rather than left undefined.
func spriteResult(pixels []color.RGBA, width, height int, err error) map[string]any {
	if err != nil {
		return map[string]any{"pixels": nil, "width": 0, "height": 0, "error": err.Error()}
	}
	return map[string]any{"pixels": rgbaToJS(pixels), "width": width, "height": height, "error": nil}
}

// rgbaToJS flattens a row-major []color.RGBA buffer into a flat, straight-
// alpha JS Uint8Array (width*height*4 bytes: R, G, B, A per pixel) —
// directly usable as new ImageData(pixels, width, height) by a caller.
func rgbaToJS(pixels []color.RGBA) js.Value {
	buf := make([]byte, len(pixels)*4)
	for i, p := range pixels {
		buf[i*4], buf[i*4+1], buf[i*4+2], buf[i*4+3] = p.R, p.G, p.B, p.A
	}
	arr := js.Global().Get("Uint8Array").New(len(buf))
	js.CopyBytesToJS(arr, buf)
	return arr
}
