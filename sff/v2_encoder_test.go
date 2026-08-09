package sff

import "testing"

func TestEncodeV2Sprite_RawFormat_RoundTripsThroughDecodeV2Sprite(t *testing.T) {
	img := &V2Image{Width: 3, Height: 2, BytesPerPixel: 1, Pixels: []byte{10, 20, 30, 40, 50, 60}}

	data, err := EncodeV2Sprite(V2FormatRaw, img)
	if err != nil {
		t.Fatalf("EncodeV2Sprite returned error: %v", err)
	}

	got, err := DecodeV2Sprite(V2FormatRaw, img.Width, img.Height, 8, data)
	if err != nil {
		t.Fatalf("DecodeV2Sprite on encoded raw data: %v", err)
	}
	if !bytesEqual(got.Pixels, img.Pixels) {
		t.Errorf("expected pixels %v, got %v", img.Pixels, got.Pixels)
	}
}

func TestEncodeV2Sprite_PNG8Format_RoundTripsThroughDecodeV2Sprite(t *testing.T) {
	img := &V2Image{Width: 4, Height: 2, BytesPerPixel: 1, Pixels: []byte{1, 2, 3, 4, 5, 6, 7, 8}}

	data, err := EncodeV2Sprite(V2FormatPNG8, img)
	if err != nil {
		t.Fatalf("EncodeV2Sprite returned error: %v", err)
	}

	got, err := DecodeV2Sprite(V2FormatPNG8, img.Width, img.Height, 8, data)
	if err != nil {
		t.Fatalf("DecodeV2Sprite on encoded PNG8 data: %v", err)
	}
	if !bytesEqual(got.Pixels, img.Pixels) {
		t.Errorf("expected pixels %v, got %v", img.Pixels, got.Pixels)
	}
}

func TestEncodeV2Sprite_PNG24Format_RoundTripsThroughDecodeV2Sprite(t *testing.T) {
	// 2x1: pure red, pure blue.
	img := &V2Image{Width: 2, Height: 1, BytesPerPixel: 3, Pixels: []byte{255, 0, 0, 0, 0, 255}}

	data, err := EncodeV2Sprite(V2FormatPNG24, img)
	if err != nil {
		t.Fatalf("EncodeV2Sprite returned error: %v", err)
	}

	got, err := DecodeV2Sprite(V2FormatPNG24, img.Width, img.Height, 24, data)
	if err != nil {
		t.Fatalf("DecodeV2Sprite on encoded PNG24 data: %v", err)
	}
	if !bytesEqual(got.Pixels, img.Pixels) {
		t.Errorf("expected pixels %v, got %v", img.Pixels, got.Pixels)
	}
}

func TestEncodeV2Sprite_PNG32Format_RoundTripsThroughDecodeV2SpriteWithAlpha(t *testing.T) {
	// 2x1: opaque green, half-transparent white — exercises non-255 alpha.
	// Pixels are straight (non-premultiplied) alpha, matching the file's
	// own on-disk representation (.vibe/decisions/006-...md, confirmed by
	// backlog item 029's real-file testing): straight-alpha white
	// (255,255,255) at alpha 128 stays (255,255,255,128) — never scaled by
	// its own alpha.
	img := &V2Image{Width: 2, Height: 1, BytesPerPixel: 4, Pixels: []byte{0, 255, 0, 255, 255, 255, 255, 128}}

	data, err := EncodeV2Sprite(V2FormatPNG32, img)
	if err != nil {
		t.Fatalf("EncodeV2Sprite returned error: %v", err)
	}

	got, err := DecodeV2Sprite(V2FormatPNG32, img.Width, img.Height, 32, data)
	if err != nil {
		t.Fatalf("DecodeV2Sprite on encoded PNG32 data: %v", err)
	}
	// Byte-exact: writing and reading straight alpha directly (no
	// premultiply/un-premultiply conversion in either direction) has no
	// lossy rounding step, unlike the premultiplied round trip this test
	// previously exercised.
	if !bytesEqual(got.Pixels, img.Pixels) {
		t.Errorf("expected pixels %v, got %v", img.Pixels, got.Pixels)
	}
}

func TestEncodeV2Sprite_SinglePixelRawSprite_RoundTrips(t *testing.T) {
	img := &V2Image{Width: 1, Height: 1, BytesPerPixel: 1, Pixels: []byte{42}}

	data, err := EncodeV2Sprite(V2FormatRaw, img)
	if err != nil {
		t.Fatalf("EncodeV2Sprite returned error: %v", err)
	}
	if !bytesEqual(data, img.Pixels) {
		t.Errorf("expected raw encoded data %v, got %v", img.Pixels, data)
	}
}

func TestEncodeV2Sprite_ReturnsErrorOnUnsupportedFormat(t *testing.T) {
	img := &V2Image{Width: 2, Height: 2, BytesPerPixel: 1, Pixels: []byte{1, 2, 3, 4}}

	// RLE8 is a real .sff v2 format code but is not encoded by this item,
	// mirroring DecodeV2Sprite's own scope — see
	// .vibe/decisions/007-sff-v2-serialize-shape-and-scope.md.
	_, err := EncodeV2Sprite(V2FormatRLE8, img)
	if err == nil {
		t.Fatal("expected an error for the unsupported RLE8 format, got nil")
	}
}

func TestEncodeV2Sprite_ReturnsErrorOnBytesPerPixelMismatch(t *testing.T) {
	// Raw format expects 1 byte per pixel (indexed); this image claims 3
	// (RGB), which raw pixel data can never represent.
	img := &V2Image{Width: 2, Height: 1, BytesPerPixel: 3, Pixels: []byte{1, 2, 3, 4, 5, 6}}

	_, err := EncodeV2Sprite(V2FormatRaw, img)
	if err == nil {
		t.Fatal("expected an error for a BytesPerPixel mismatch, got nil")
	}
}
