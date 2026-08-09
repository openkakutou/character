package sff

// Fixture-driven v2 sprite test suite (backlog item 029): ports every real
// scenario from the reference project's own "extract-v2.test.mjs" (see
// .vibe/backlog/done/029-fixture-driven-v2-sprite-test-suite.md), commit
// 2d4af64d26441bf4d692bb479275d64b11869678. Each one resolves a sprite from
// a real, trimmed .sff v2 fixture under testdata/files (see
// testdata/README.md) through the public ResolveSpritePixels combinator,
// and compares the resulting RGBA pixel buffer against the reference
// project's own expected PNG under testdata/sprites — pixel for pixel, so
// no PNG-encoding-choice difference can produce a false mismatch.
//
// Scenario-to-fixture mapping was derived by running the reference
// project's own test suite (node --test) against its real, unmodified
// source files rather than guessed from file/test names alone: two of its
// test names ("RLE8", "LZ5") do not actually describe the compression
// format of the sprite each one ends up decoding (a quirk of how the
// reference project's own group filter/array-index selection lands), but
// the ported scenario here is defined by matching decoded pixel output,
// not by the label.

import (
	"os"
	"path/filepath"
	"testing"
)

// v2FixtureCase describes one ported scenario: which trimmed fixture to
// resolve a sprite from, that sprite's (group, image) key, an optional
// external .act palette override, and the expected reference PNG to
// compare against.
type v2FixtureCase struct {
	name    string
	sff     string
	group   int
	image   int
	actFile string // "" if the scenario uses the sprite's own palette bank
	png     string
}

var v2FixtureCases = []v2FixtureCase{
	{"RLE8", "v2-rle8.sff", 9000, 0, "", "v2-sprite-001.png"},
	{"LZ5", "v2-lz5.sff", 0, 0, "", "v2-sprite-002.png"},
	{"PNG8", "v2-png8.sff", 0, 0, "", "v2-sprite-003.png"},
	{"PNG24", "v2-png24.sff", 6053, 0, "", "v2-sprite-004.png"},
	{"PNG32", "v2-png32.sff", 9000, 45, "", "v2-sprite-005.png"},
	{"ZeroLengthCopyOfFirstSprite", "v2-zero-length-copy.sff", 186, 0, "", "v2-sprite-006.png"},
	{"PNG8ForceAlpha255", "v2-png8-forced-alpha.sff", 9000, 0, "", "v2-sprite-007.png"},
	{"ExternalPalette", "v2-external-palette-source.sff", 0, 0, "ruby-v2-palette1.act", "v2-sprite-008.png"},
	{"EmptyPaletteUseFirstPalette", "v2-empty-palette-use-first.sff", 0, 0, "", "v2-sprite-009.png"},
	{"ExternalPaletteFirstColorTransparent", "v2-empty-palette-use-first.sff", 0, 0, "makina-v2-palette1.act", "v2-sprite-010.png"},
	{"EmptyPaletteUsePreviousPalette", "v2-empty-palette-use-previous.sff", 0, 0, "", "v2-sprite-011.png"},
	{"EmptyPaletteUsePreviousPalette2", "v2-empty-palette-use-previous2.sff", 0, 0, "", "v2-sprite-012.png"},
	{"PNG8Second", "v2-png8-b.sff", 9000, 1, "", "v2-sprite-013.png"},
	{"PNG8LoadMode1", "v2-loadmode1.sff", 3096, 14, "", "v2-sprite-014.png"},
}

// TestV2Fixtures_DecodedPixelsMatchReferencePNGs ports every real-file
// scenario above: ResolveSpritePixels must produce a pixel-for-pixel
// identical image to the one the reference project's own decoder produced
// for the same real source sprite.
func TestV2Fixtures_DecodedPixelsMatchReferencePNGs(t *testing.T) {
	for _, c := range v2FixtureCases {
		t.Run(c.name, func(t *testing.T) {
			f := openTestdataFile(t, c.sff)
			defer f.Close()

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

			got, w, h, err := ResolveSpritePixels(f, c.group, c.image, override)
			if err != nil {
				t.Fatalf("ResolveSpritePixels(group=%d, image=%d): %v", c.group, c.image, err)
			}

			wantImg := decodeExpectedPNG(t, c.png)
			assertPixelsMatch(t, got, w, h, wantImg)
		})
	}
}

// TestV2Fixtures_VersionStringIsReadFromHeader asserts version-string
// extraction (the reference project's own "Extract v2 metadata" scenario):
// it reads the same 4 raw bytes as major.minor.patch.build and reports
// "2.0.1.0" for kfm-v2.sff, the real source every v2-rle8.sff/v2-lz5.sff
// fixture was trimmed from — testdata/gen/main.go's trimV2 copies a
// trimmed fixture's version bytes verbatim from its real source file, so
// that value survives trimming unchanged.
func TestV2Fixtures_VersionStringIsReadFromHeader(t *testing.T) {
	f := openTestdataFile(t, "v2-rle8.sff")
	defer f.Close()

	table, err := ParseV2(f)
	if err != nil {
		t.Fatalf("ParseV2: %v", err)
	}
	// Version holds the raw on-disk byte order (build, patch, minor,
	// major) — see V2Header.Version's own doc comment. Reversed, this is
	// major.minor.patch.build = "2.0.1.0", matching the reference
	// project's own reported version string for kfm-v2.sff.
	want := [4]byte{0, 1, 0, 2}
	if table.Header.Version != want {
		t.Errorf("got version %v, want %v (= \"2.0.1.0\" reversed)", table.Header.Version, want)
	}
}
