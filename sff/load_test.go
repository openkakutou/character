package sff

import (
	"bytes"
	"testing"
)

func TestLoad_V1File_AssemblesSpriteGroupsWithDecodedDimensionsAndDerivedPalettes(t *testing.T) {
	sprites := []V1WriteSprite{
		{
			Group: 0, Image: 0, AxisX: 4, AxisY: -2, SharedPalette: false,
			PixelData: mustEncodePCX(t, &PCXImage{Width: 4, Height: 2, Pixels: []byte{1, 1, 1, 1, 2, 2, 2, 2}}),
		},
		{
			// Shares sprite 0's palette: same derived Palette value.
			Group: 0, Image: 1, AxisX: 5, AxisY: -1, SharedPalette: true,
			PixelData: mustEncodePCX(t, &PCXImage{Width: 3, Height: 1, Pixels: []byte{3, 3, 3}}),
		},
		{
			// Its own palette again: derived Palette value bumps up.
			Group: 1, Image: 0, AxisX: 0, AxisY: 0, SharedPalette: false,
			PixelData: mustEncodePCX(t, &PCXImage{Width: 2, Height: 2, Pixels: []byte{9, 9, 9, 9}}),
		},
	}

	var buf bytes.Buffer
	if err := SerializeV1(&buf, [4]byte{1, 0, 0, 1}, false, sprites); err != nil {
		t.Fatalf("test setup: SerializeV1 failed: %v", err)
	}

	groups, err := Load(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("expected 2 sprite groups, got %d", len(groups))
	}
	if groups[0].Index != 0 || groups[1].Index != 1 {
		t.Fatalf("expected groups in ascending index order [0, 1], got [%d, %d]", groups[0].Index, groups[1].Index)
	}

	if len(groups[0].Sprites) != 2 {
		t.Fatalf("expected 2 sprites in group 0, got %d", len(groups[0].Sprites))
	}
	want := Sprite{Group: 0, Image: 0, Width: 4, Height: 2, AxisX: 4, AxisY: -2, Palette: 0}
	if got := groups[0].Sprites[0]; got != want {
		t.Errorf("sprite (0,0): expected %+v, got %+v", want, got)
	}
	want = Sprite{Group: 0, Image: 1, Width: 3, Height: 1, AxisX: 5, AxisY: -1, Palette: 0}
	if got := groups[0].Sprites[1]; got != want {
		t.Errorf("sprite (0,1): expected %+v, got %+v", want, got)
	}

	if len(groups[1].Sprites) != 1 {
		t.Fatalf("expected 1 sprite in group 1, got %d", len(groups[1].Sprites))
	}
	want = Sprite{Group: 1, Image: 0, Width: 2, Height: 2, AxisX: 0, AxisY: 0, Palette: 1}
	if got := groups[1].Sprites[0]; got != want {
		t.Errorf("sprite (1,0): expected %+v, got %+v", want, got)
	}
}

func TestLoad_V2File_AssemblesSpriteGroupsDirectlyFromTableWithoutDecodingPixels(t *testing.T) {
	sprites := []V2WriteSprite{
		{
			Group: 0, Image: 0, Width: 3, Height: 2, AxisX: 1, AxisY: -1,
			Format: V2FormatRaw, ColorDepth: 8, PaletteIndex: 0,
			// Deliberately unsupported pixel encoding — Load must not need
			// to decode pixel data for v2 to succeed.
			PixelData: []byte{0xFF, 0xFF, 0xFF},
		},
		{
			Group: 1, Image: 0, Width: 1, Height: 1, AxisX: 0, AxisY: 0,
			Format: V2FormatRLE8, ColorDepth: 8, PaletteIndex: 2,
			PixelData: []byte{0x00},
		},
	}

	var buf bytes.Buffer
	if err := SerializeV2(&buf, [4]byte{0, 1, 0, 2}, sprites, nil); err != nil {
		t.Fatalf("test setup: SerializeV2 failed: %v", err)
	}

	groups, err := Load(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("expected 2 sprite groups, got %d", len(groups))
	}

	want := Sprite{Group: 0, Image: 0, Width: 3, Height: 2, AxisX: 1, AxisY: -1, Palette: 0}
	if got := groups[0].Sprites[0]; got != want {
		t.Errorf("sprite (0,0): expected %+v, got %+v", want, got)
	}
	want = Sprite{Group: 1, Image: 0, Width: 1, Height: 1, AxisX: 0, AxisY: 0, Palette: 2}
	if got := groups[1].Sprites[0]; got != want {
		t.Errorf("sprite (1,0): expected %+v, got %+v", want, got)
	}
}

func TestLoad_V1LinkedSprite_ResolvesDimensionsFromLinkTarget(t *testing.T) {
	targetPixels := &PCXImage{Width: 5, Height: 3, Pixels: make([]byte, 15)}
	for i := range targetPixels.Pixels {
		targetPixels.Pixels[i] = byte(i)
	}

	sprites := []V1WriteSprite{
		{Group: 0, Image: 0, PixelData: mustEncodePCX(t, targetPixels)},
		{Group: 0, Image: 1, SharedPalette: true, LinkedIndex: 0}, // no PixelData, links to sprite 0
	}

	var buf bytes.Buffer
	if err := SerializeV1(&buf, [4]byte{1, 0, 0, 1}, true, sprites); err != nil {
		t.Fatalf("test setup: SerializeV1 failed: %v", err)
	}

	groups, err := Load(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(groups) != 1 || len(groups[0].Sprites) != 2 {
		t.Fatalf("expected 1 group with 2 sprites, got %+v", groups)
	}
	linked := groups[0].Sprites[1]
	if linked.Width != 5 || linked.Height != 3 {
		t.Errorf("expected linked sprite to inherit dimensions 5x3 from its link target, got %dx%d", linked.Width, linked.Height)
	}
}

func TestLoad_V1MutualLinkCycle_ReturnsDescriptiveErrorInsteadOfHangingOrPanicking(t *testing.T) {
	sprites := []V1WriteSprite{
		{Group: 0, Image: 0, LinkedIndex: 1}, // links to sprite 1, no PixelData
		{Group: 0, Image: 1, LinkedIndex: 0}, // links to sprite 0, no PixelData
	}

	var buf bytes.Buffer
	if err := SerializeV1(&buf, [4]byte{1, 0, 0, 1}, false, sprites); err != nil {
		t.Fatalf("test setup: SerializeV1 failed: %v", err)
	}

	_, err := Load(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatal("expected an error for a mutual link cycle with no real pixel data anywhere, got nil")
	}
}

func TestLoad_UnrecognizedSignature_ReturnsDescriptiveError(t *testing.T) {
	data := bytes.Repeat([]byte("not-an-sff-file!"), 4)

	_, err := Load(bytes.NewReader(data))
	if err == nil {
		t.Fatal("expected an error for data with an unrecognized signature, got nil")
	}
}

func TestLoad_TruncatedFile_ReturnsErrorRatherThanPanicking(t *testing.T) {
	data := []byte{1, 2, 3} // shorter than the 16-byte signature/version peek

	_, err := Load(bytes.NewReader(data))
	if err == nil {
		t.Fatal("expected an error for a file too short to contain a header, got nil")
	}
}
