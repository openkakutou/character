package sff

import (
	"fmt"
	"image/color"
	"io"
)

// v1PaletteBlockSize is the fixed size, in bytes, of a .sff v1 sprite's
// embedded 256-color RGB palette block.
const v1PaletteBlockSize = 768

// Palette is a resolved 256-entry RGBA color table, indexed by a decoded
// sprite's palette index bytes (PCXImage.Pixels / V2Image.Pixels with
// BytesPerPixel: 1).
//
// It is kept separate from PCXImage/V2Image/Sprite — the read-path
// pure-data types — as an explicit opt-in helper, mirroring how DecodePCX
// and DecodeV2Sprite are already separate from Load; see
// .vibe/decisions/014-palette-resolution-api-shape.md.
type Palette [256]color.RGBA

// AlphaRule selects how ResolvePixels determines a resolved pixel's alpha
// channel at palette index 0. The reference decoders this package tracks
// apply one of two rules depending on the sprite's own encoding, not a
// single universal rule.
type AlphaRule int

const (
	// AlphaForceTransparentAtIndexZero forces palette index 0 to resolve to
	// fully transparent ((0,0,0,0)), regardless of the palette's own stored
	// value there. Used by PCX (v1) and PNG8 (v2) decoded pixel data.
	AlphaForceTransparentAtIndexZero AlphaRule = iota
	// AlphaLiteral uses the palette's own stored alpha value unmodified,
	// including at index 0. Used by RLE8/LZ5 (v2) decoded pixel data.
	AlphaLiteral
)

// ResolvePixels resolves a row-major palette-index pixel buffer — as
// decoded by DecodePCX or DecodeV2Sprite for an indexed pixel format
// (BytesPerPixel: 1) — against palette into a row-major buffer of final
// RGBA colors, one per index, applying rule to determine index 0's alpha.
func ResolvePixels(indices []byte, palette Palette, rule AlphaRule) []color.RGBA {
	pixels := make([]color.RGBA, len(indices))
	for i, idx := range indices {
		c := palette[idx]
		if rule == AlphaForceTransparentAtIndexZero && idx == 0 {
			c = color.RGBA{}
		}
		pixels[i] = c
	}
	return pixels
}

// DecodeV1Palette decodes a .sff v1 sprite's embedded palette block — 256
// RGB triplets (3 bytes/color), always opaque — into a Palette. data must
// be exactly v1PaletteBlockSize (768) bytes: the palette block itself, not
// the sprite's pixel data it follows (see ResolveV1Palette, which locates
// it within a sprite's raw file bytes).
func DecodeV1Palette(data []byte) (Palette, error) {
	if len(data) != v1PaletteBlockSize {
		return Palette{}, fmt.Errorf("sff: v1 palette block is %d bytes, want exactly %d", len(data), v1PaletteBlockSize)
	}

	var p Palette
	for i := range p {
		p[i] = color.RGBA{R: data[i*3], G: data[i*3+1], B: data[i*3+2], A: 255}
	}
	return p, nil
}

// ResolveV1Palette resolves the palette used by sprite index i in table:
// the nearest sprite at or before i (walking backward) whose SharedPalette
// is false, i.e. the one that carries its own embedded palette rather than
// reusing the previous sprite's. That owning sprite's palette block is
// read via r from immediately after its own pixel data (Offset+Length,
// v1PaletteBlockSize bytes) — a v1 sprite's declared pixel-data Length does
// not include its trailing palette block, confirmed against the real
// fixtures in testdata/files (see
// .vibe/decisions/014-palette-resolution-api-shape.md) — and decoded via
// DecodeV1Palette.
func ResolveV1Palette(table *V1SpriteTable, r io.ReaderAt, i int) (Palette, error) {
	if i < 0 || i >= len(table.Sprites) {
		return Palette{}, fmt.Errorf("sff: v1 palette: sprite index %d out of range (have %d sprites)", i, len(table.Sprites))
	}

	owner := -1
	for j := i; j >= 0; j-- {
		if !table.Sprites[j].SharedPalette {
			owner = j
			break
		}
	}
	if owner == -1 {
		return Palette{}, fmt.Errorf("sff: v1 palette: sprite %d: no earlier sprite owns a palette", i)
	}

	e := table.Sprites[owner]
	block := make([]byte, v1PaletteBlockSize)
	if _, err := r.ReadAt(block, e.Offset+int64(e.Length)); err != nil {
		return Palette{}, fmt.Errorf("sff: v1 palette: reading sprite %d's embedded palette block: %w", owner, err)
	}
	return DecodeV1Palette(block)
}

// DecodeV2Palette decodes a .sff v2 palette bank's raw color data — already
// stored as RGBA on disk, 4 bytes/color, unlike v1's opaque-only RGB
// triplets — into a Palette. data's length must be a multiple of 4 and
// declare at most 256 colors; any entries beyond the declared color count
// are left at their zero value (fully transparent black).
func DecodeV2Palette(data []byte) (Palette, error) {
	if len(data)%4 != 0 {
		return Palette{}, fmt.Errorf("sff: v2 palette color data length %d is not a multiple of 4", len(data))
	}
	count := len(data) / 4
	if count > 256 {
		return Palette{}, fmt.Errorf("sff: v2 palette color data declares %d colors, more than the maximum 256", count)
	}

	var p Palette
	for i := 0; i < count; i++ {
		p[i] = color.RGBA{R: data[i*4], G: data[i*4+1], B: data[i*4+2], A: data[i*4+3]}
	}
	return p, nil
}

// ResolveV2Palette resolves the palette bank at index within table.Palettes
// to its final Palette: its own RGBA color data (read via r) when it
// stores one (Length > 0), or the colors of the bank its LinkedIndex points
// to otherwise — following the chain across further zero-length banks if
// needed. A link chain that is out of range or cycles back on itself
// returns a descriptive error instead of looping or panicking, mirroring
// v1 linked-sprite resolution (resolveV1Pixels).
func ResolveV2Palette(table *V2SpriteTable, r io.ReaderAt, index int) (Palette, error) {
	return resolveV2Palette(table, r, index, nil)
}

func resolveV2Palette(table *V2SpriteTable, r io.ReaderAt, index int, seen map[int]bool) (Palette, error) {
	if index < 0 || index >= len(table.Palettes) {
		return Palette{}, fmt.Errorf("sff: v2 palette: bank index %d out of range (have %d banks)", index, len(table.Palettes))
	}
	if seen == nil {
		seen = make(map[int]bool, len(table.Palettes))
	}
	if seen[index] {
		return Palette{}, fmt.Errorf("sff: v2 palette: linked bank chain cycles back to bank index %d", index)
	}
	seen[index] = true

	e := table.Palettes[index]
	if e.Length > 0 {
		raw := make([]byte, e.Length)
		if _, err := r.ReadAt(raw, e.Offset); err != nil {
			return Palette{}, fmt.Errorf("sff: v2 palette: reading bank %d color data: %w", index, err)
		}
		return DecodeV2Palette(raw)
	}

	return resolveV2Palette(table, r, e.LinkedIndex, seen)
}
