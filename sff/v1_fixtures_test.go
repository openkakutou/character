package sff

// Fixture-driven v1 sprite test suite (backlog item 028): ports each real
// scenario from the reference project's own "extract-v1.test.mjs" (see
// .vibe/backlog/done/028-fixture-driven-v1-sprite-test-suite.md) as a Go
// test. Each one decodes a sprite from a real, trimmed .sff v1 fixture
// under testdata/files (see testdata/README.md — the target sprite is
// always the last entry in the trimmed table) and compares the fully
// resolved RGBA pixel buffer against the reference project's own expected
// PNG under testdata/sprites, decoded via the standard library so no
// PNG-encoding-choice difference can produce a false mismatch.

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// v1FixtureCase describes one ported scenario: which trimmed fixture to
// decode, an optional external .act palette override, and the expected
// reference PNG to compare against.
type v1FixtureCase struct {
	name    string
	sff     string
	actFile string // "" if the scenario uses the fixture's own palette
	png     string
}

var v1FixtureCases = []v1FixtureCase{
	{"BasicSprite", "v1-basic.sff", "", "v1-sprite-001.png"},
	{"LastGroupSprite", "v1-last-sprite.sff", "", "v1-sprite-002.png"},
	{"ZeroLengthCopyOfPreviousSprite", "v1-zero-length-copy.sff", "", "v1-sprite-003.png"},
	{"IndexEqualsLinkedIndex", "v1-self-link.sff", "greenarrow-v1-palette1.act", "v1-sprite-004.png"},
	{"LinkedIndexGreaterThanCurrentIndex", "v1-forward-link.sff", "", "v1-sprite-005.png"},
	{"InvalidSpriteSize", "v1-invalid-size.sff", "", "v1-sprite-006.png"},
	{"ExternalPalette", "v1-external-palette-source.sff", "cyclops-v1-palette1.act", "v1-sprite-007.png"},
	{"SamePaletteAsFirstSpriteInGroupZero", "v1-same-palette-a.sff", "", "v1-sprite-008.png"},
	{"SamePaletteAsFirstSpriteInGroupZero_SecondFile", "v1-same-palette-b.sff", "", "v1-sprite-009.png"},
	{"SamePaletteAsFirstSpriteInGroupZero_DifferentGroup", "v1-same-palette-a-group5.sff", "", "v1-sprite-010.png"},
}

// TestV1Fixtures_DecodedPixelsMatchReferencePNGs ports every real-file
// scenario above: it must produce a pixel-for-pixel identical image to the
// one the reference project's own decoder produced for the same real
// source sprite.
func TestV1Fixtures_DecodedPixelsMatchReferencePNGs(t *testing.T) {
	for _, c := range v1FixtureCases {
		t.Run(c.name, func(t *testing.T) {
			f := openTestdataFile(t, c.sff)
			defer f.Close()

			table, err := ParseV1(f)
			if err != nil {
				t.Fatalf("ParseV1(%s): %v", c.sff, err)
			}
			target := len(table.Sprites) - 1

			var override *Palette
			if c.actFile != "" {
				actData, err := os.ReadFile(filepath.Join("testdata", "files", c.actFile))
				if err != nil {
					t.Fatalf("reading external palette %s: %v", c.actFile, err)
				}
				p, err := DecodeExternalPalette(actData)
				if err != nil {
					t.Fatalf("DecodeExternalPalette(%s): %v", c.actFile, err)
				}
				override = &p
			}

			img, err := ResolveV1Pixels(table, f, target)
			if err != nil {
				t.Fatalf("resolving pixel data: %v", err)
			}
			pal, err := ResolveV1Palette(table, f, target, override)
			if err != nil {
				t.Fatalf("ResolveV1Palette: %v", err)
			}
			got := ResolvePixels(img.Pixels, pal, AlphaForceTransparentAtIndexZero)

			wantImg := decodeExpectedPNG(t, c.png)
			assertPixelsMatch(t, got, img.Width, img.Height, wantImg)
		})
	}
}

// TestV1Fixtures_VersionStringIsReadFromHeader asserts version-string
// extraction (the eleventh ported scenario, "Extract v1 metadata"): the
// reference project reads the same 4 raw bytes as major.minor.patch.build
// (reversed field order) and reports "1.0.1.0" for cvsryu-v1.sff, the real
// source v1-basic.sff was trimmed from — testdata/gen/main.go's encodeV1
// copies a trimmed fixture's version bytes verbatim from its real source
// file, so that value survives trimming unchanged.
func TestV1Fixtures_VersionStringIsReadFromHeader(t *testing.T) {
	f := openTestdataFile(t, "v1-basic.sff")
	defer f.Close()

	table, err := ParseV1(f)
	if err != nil {
		t.Fatalf("ParseV1: %v", err)
	}
	// Version holds the raw on-disk byte order (build, patch, minor,
	// major) — see V1Header.Version's own doc comment. Reversed, this is
	// major.minor.patch.build = "1.0.1.0", matching the reference
	// project's own reported version string for cvsryu-v1.sff.
	want := [4]byte{0, 1, 0, 1}
	if table.Header.Version != want {
		t.Errorf("got version %v, want %v (= \"1.0.1.0\" reversed)", table.Header.Version, want)
	}
}

// decodeExpectedPNG reads and decodes one of the reference project's own
// expected-output PNGs under testdata/sprites. It is type-asserted to
// *image.NRGBA (non-alpha-premultiplied) rather than returned as the
// generic image.Image the standard decoder interface offers: the
// reference project's own PNGs are always written with colorType 6
// (truecolor+alpha, no premultiplication — see
// convertSpriteDecodedBufferToPng.mjs), so Go's decoder always produces
// *image.NRGBA for them, and reading its raw pixel struct directly (rather
// than calling .At(x,y).RGBA(), whose interface contract always returns
// alpha-premultiplied values) is required to compare a fully transparent
// pixel's color channels at all: alpha-premultiplying a (R,G,B,0) pixel
// always yields (0,0,0,0), silently destroying real color information a
// non-black "transparent" palette entry (e.g. an applied external
// palette's own index 0) still carries.
func decodeExpectedPNG(t *testing.T, name string) *image.NRGBA {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "sprites", name))
	if err != nil {
		t.Fatalf("reading expected PNG %s: %v", name, err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decoding expected PNG %s: %v", name, err)
	}
	nrgba, ok := img.(*image.NRGBA)
	if !ok {
		t.Fatalf("expected PNG %s decoded as %T, want *image.NRGBA (non-alpha-premultiplied)", name, img)
	}
	return nrgba
}

// assertPixelsMatch compares a resolved RGBA pixel buffer (row-major,
// width*height entries) against an expected reference image, dimension by
// dimension and then pixel by pixel, reporting the first handful of
// mismatches rather than every one. Both sides are compared as
// non-alpha-premultiplied color channels — see decodeExpectedPNG's doc
// comment for why that distinction matters here.
func assertPixelsMatch(t *testing.T, got []color.RGBA, width, height int, want *image.NRGBA) {
	t.Helper()
	wb := want.Bounds()
	if width != wb.Dx() || height != wb.Dy() {
		t.Fatalf("dimensions: got %dx%d, want %dx%d", width, height, wb.Dx(), wb.Dy())
	}

	mismatches := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			wc := want.NRGBAAt(wb.Min.X+x, wb.Min.Y+y)
			gc := got[y*width+x]
			if gc.R != wc.R || gc.G != wc.G || gc.B != wc.B || gc.A != wc.A {
				mismatches++
				if mismatches <= 5 {
					t.Errorf("pixel (%d,%d): got %+v, want %+v", x, y, gc, wc)
				}
			}
		}
	}
	if mismatches > 5 {
		t.Errorf("... and %d more mismatching pixels (%d/%d total)", mismatches-5, mismatches, width*height)
	}
}
