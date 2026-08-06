package sff

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// buildPNG8 encodes a paletted (8-bit indexed) PNG from a row-major index
// buffer, mirroring how a .sff v2 PNG8 sprite's pixel data is stored.
func buildPNG8(t *testing.T, width, height int, indices []byte) []byte {
	t.Helper()

	palette := make(color.Palette, 256)
	for i := range palette {
		palette[i] = color.RGBA{R: uint8(i), G: uint8(i), B: uint8(i), A: 255}
	}

	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetColorIndex(x, y, indices[y*width+x])
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encoding test PNG8 fixture: %v", err)
	}
	return buf.Bytes()
}

// buildPNG24 encodes an RGB (no alpha) PNG from a row-major RGB byte
// buffer, mirroring how a .sff v2 PNG24 sprite's pixel data is stored.
func buildPNG24(t *testing.T, width, height int, rgb []byte) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 3
			img.SetRGBA(x, y, color.RGBA{R: rgb[i], G: rgb[i+1], B: rgb[i+2], A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encoding test PNG24 fixture: %v", err)
	}
	return buf.Bytes()
}

// buildPNG32 encodes an RGBA PNG from a row-major RGBA byte buffer,
// mirroring how a .sff v2 PNG32 sprite's pixel data is stored.
func buildPNG32(t *testing.T, width, height int, rgba []byte) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 4
			img.SetNRGBA(x, y, color.NRGBA{R: rgba[i], G: rgba[i+1], B: rgba[i+2], A: rgba[i+3]})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encoding test PNG32 fixture: %v", err)
	}
	return buf.Bytes()
}

func TestDecodeV2Sprite_DecodesRawIndexedData(t *testing.T) {
	// 3x2 raw (uncompressed) 8-bit indexed sprite: literally one byte per
	// pixel, row-major, no header of its own.
	data := []byte{10, 20, 30, 40, 50, 60}

	img, err := DecodeV2Sprite(V2FormatRaw, 3, 2, 8, data)
	if err != nil {
		t.Fatalf("DecodeV2Sprite returned error: %v", err)
	}
	if img.Width != 3 || img.Height != 2 {
		t.Fatalf("got Width=%d Height=%d, want Width=3 Height=2", img.Width, img.Height)
	}
	if img.BytesPerPixel != 1 {
		t.Fatalf("got BytesPerPixel=%d, want 1", img.BytesPerPixel)
	}
	if !bytesEqual(img.Pixels, data) {
		t.Fatalf("got Pixels=%v, want %v", img.Pixels, data)
	}
}

func TestDecodeV2Sprite_DecodesPNG8IndexedData(t *testing.T) {
	indices := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	data := buildPNG8(t, 4, 2, indices)

	img, err := DecodeV2Sprite(V2FormatPNG8, 4, 2, 8, data)
	if err != nil {
		t.Fatalf("DecodeV2Sprite returned error: %v", err)
	}
	if img.Width != 4 || img.Height != 2 {
		t.Fatalf("got Width=%d Height=%d, want Width=4 Height=2", img.Width, img.Height)
	}
	if img.BytesPerPixel != 1 {
		t.Fatalf("got BytesPerPixel=%d, want 1", img.BytesPerPixel)
	}
	if !bytesEqual(img.Pixels, indices) {
		t.Fatalf("got Pixels=%v, want %v", img.Pixels, indices)
	}
}

func TestDecodeV2Sprite_DecodesPNG24ColorData(t *testing.T) {
	// 2x1 image: pixel 0 pure red, pixel 1 pure blue.
	rgb := []byte{255, 0, 0, 0, 0, 255}
	data := buildPNG24(t, 2, 1, rgb)

	img, err := DecodeV2Sprite(V2FormatPNG24, 2, 1, 24, data)
	if err != nil {
		t.Fatalf("DecodeV2Sprite returned error: %v", err)
	}
	if img.BytesPerPixel != 3 {
		t.Fatalf("got BytesPerPixel=%d, want 3", img.BytesPerPixel)
	}
	if !bytesEqual(img.Pixels, rgb) {
		t.Fatalf("got Pixels=%v, want %v", img.Pixels, rgb)
	}
}

func TestDecodeV2Sprite_DecodesPNG32ColorDataWithAlpha(t *testing.T) {
	// 2x1 image: pixel 0 opaque green, pixel 1 half-transparent white.
	rgba := []byte{0, 255, 0, 255, 255, 255, 255, 128}
	data := buildPNG32(t, 2, 1, rgba)

	img, err := DecodeV2Sprite(V2FormatPNG32, 2, 1, 32, data)
	if err != nil {
		t.Fatalf("DecodeV2Sprite returned error: %v", err)
	}
	if img.BytesPerPixel != 4 {
		t.Fatalf("got BytesPerPixel=%d, want 4", img.BytesPerPixel)
	}
	if !bytesEqual(img.Pixels, rgba) {
		t.Fatalf("got Pixels=%v, want %v", img.Pixels, rgba)
	}
}

func TestDecodeV2Sprite_DecodesSinglePixelRawSprite(t *testing.T) {
	// 1x1 boundary case.
	data := []byte{42}

	img, err := DecodeV2Sprite(V2FormatRaw, 1, 1, 8, data)
	if err != nil {
		t.Fatalf("DecodeV2Sprite returned error: %v", err)
	}
	if img.Width != 1 || img.Height != 1 {
		t.Fatalf("got Width=%d Height=%d, want Width=1 Height=1", img.Width, img.Height)
	}
	if !bytesEqual(img.Pixels, data) {
		t.Fatalf("got Pixels=%v, want %v", img.Pixels, data)
	}
}

func TestDecodeV2Sprite_ReturnsErrorOnRawDataLengthMismatch(t *testing.T) {
	// Declares a 3x2 sprite (6 bytes needed) but supplies only 4.
	data := []byte{1, 2, 3, 4}

	_, err := DecodeV2Sprite(V2FormatRaw, 3, 2, 8, data)
	if err == nil {
		t.Fatal("expected an error for truncated raw pixel data, got nil")
	}
}

func TestDecodeV2Sprite_ReturnsErrorOnUnrecognizedFormat(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6}

	_, err := DecodeV2Sprite(99, 3, 2, 8, data)
	if err == nil {
		t.Fatal("expected an error for an unrecognized format code, got nil")
	}
}

func TestDecodeV2Sprite_ReturnsErrorOnUnimplementedCompressedFormat(t *testing.T) {
	// RLE8 is a real .sff v2 format code (see ParseV2's V2Format*
	// constants) but is not decoded by this item; it must still fail
	// descriptively rather than being silently misinterpreted as raw data.
	data := []byte{1, 2, 3, 4, 5, 6}

	_, err := DecodeV2Sprite(V2FormatRLE8, 3, 2, 8, data)
	if err == nil {
		t.Fatal("expected an error for the unimplemented RLE8 format, got nil")
	}
}

func TestDecodeV2Sprite_ReturnsErrorOnCorruptPNGData(t *testing.T) {
	data := []byte{0x89, 'P', 'N', 'G', 0, 0, 0, 0, 'g', 'a', 'r', 'b', 'a', 'g', 'e'}

	_, err := DecodeV2Sprite(V2FormatPNG8, 4, 2, 8, data)
	if err == nil {
		t.Fatal("expected an error for corrupt PNG data, got nil")
	}
}

func TestDecodeV2Sprite_ReturnsErrorOnPNGDimensionMismatch(t *testing.T) {
	// Encodes a 4x2 PNG but declares (via the table entry's own
	// width/height) that it should be 2x2.
	data := buildPNG8(t, 4, 2, []byte{1, 2, 3, 4, 5, 6, 7, 8})

	_, err := DecodeV2Sprite(V2FormatPNG8, 2, 2, 8, data)
	if err == nil {
		t.Fatal("expected an error for a PNG whose dimensions disagree with the declared sprite size, got nil")
	}
}
