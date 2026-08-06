package sff

import (
	"fmt"
	"io"
	"sort"
)

// signaturePeekSize is the number of leading bytes Load reads to identify a
// file: the 12-byte signature both .sff versions share, plus the 4 version
// bytes that distinguish them (see ParseV2's Version[3] check).
const signaturePeekSize = 16

// Load reads a full MUGEN/Ikemen GO .sff file from r — version 1 or version
// 2, auto-detected from the file's own signature and version bytes — and
// assembles it into the shared read-path SpriteGroup model, decoding
// whatever pixel data is needed to do so.
//
// v1's sprite table carries no width/height (see
// .vibe/decisions/004-sff-v1-table-is-a-separate-low-level-type.md): Load
// decodes each sprite's PCX pixel data to recover them, resolving a linked
// sprite (one with no pixel data of its own) to its link target first. A
// link chain that is out of range or cycles back on itself returns a
// descriptive error instead of looping or panicking. v1's table also only
// records whether a sprite reuses the previous sprite's palette, not a
// numeric reference, so Sprite.Palette is derived by incrementing a counter
// on every sprite that does not share, and having sharing sprites inherit
// the current value — see
// .vibe/decisions/010-def-loader-assembles-character-from-referenced-files.md.
//
// v2's sprite table already carries width/height and an explicit palette
// bank index, so each V2SpriteEntry maps directly to a Sprite without
// decoding any pixel data.
func Load(r io.ReaderAt) ([]SpriteGroup, error) {
	peek := make([]byte, signaturePeekSize)
	if _, err := r.ReadAt(peek, 0); err != nil {
		return nil, fmt.Errorf("sff: reading file header: %w", err)
	}
	if sig := string(peek[0:12]); sig != v1Signature {
		return nil, fmt.Errorf("sff: not a .sff file: unexpected signature %q", sig)
	}

	// Matches ParseV2's own check: verhi (the version's high byte) is
	// stored at this offset and is 2 only for a v2 file.
	if peek[15] == 2 {
		return loadV2(r)
	}
	return loadV1(r)
}

// loadV1 assembles a v1 .sff file's sprite table into SpriteGroups,
// decoding each sprite's PCX pixel data to recover its dimensions and
// deriving a Palette value from the table's per-sprite palette-sharing
// flag.
func loadV1(r io.ReaderAt) ([]SpriteGroup, error) {
	table, err := ParseV1(r)
	if err != nil {
		return nil, err
	}

	sprites := make([]Sprite, len(table.Sprites))
	palette := -1
	for i, e := range table.Sprites {
		if !e.SharedPalette || palette == -1 {
			palette++
		}

		img, err := resolveV1Pixels(table, r, i, nil)
		if err != nil {
			return nil, fmt.Errorf("sff: sprite %d (group %d, image %d): %w", i, e.Group, e.Image, err)
		}

		sprites[i] = Sprite{
			Group:   e.Group,
			Image:   e.Image,
			Width:   img.Width,
			Height:  img.Height,
			AxisX:   e.AxisX,
			AxisY:   e.AxisY,
			Palette: palette,
		}
	}

	return groupSprites(sprites), nil
}

// resolveV1Pixels returns the decoded PCX pixel data belonging to sprite
// index i in table, following its LinkedIndex chain when it stores no pixel
// data of its own (Length == 0). seen tracks indices already visited on the
// current chain, so a link cycle is reported as an error rather than
// recursing forever; pass nil on the initial call.
func resolveV1Pixels(table *V1SpriteTable, r io.ReaderAt, i int, seen map[int]bool) (*PCXImage, error) {
	if seen == nil {
		seen = make(map[int]bool, len(table.Sprites))
	}
	if seen[i] {
		return nil, fmt.Errorf("linked sprite chain cycles back to sprite index %d", i)
	}
	seen[i] = true

	e := table.Sprites[i]
	if e.Length > 0 {
		raw := make([]byte, e.Length)
		if _, err := r.ReadAt(raw, e.Offset); err != nil {
			return nil, fmt.Errorf("reading pixel data: %w", err)
		}
		return DecodePCX(raw)
	}

	if e.LinkedIndex < 0 || e.LinkedIndex >= len(table.Sprites) {
		return nil, fmt.Errorf("links to out-of-range sprite index %d", e.LinkedIndex)
	}
	return resolveV1Pixels(table, r, e.LinkedIndex, seen)
}

// loadV2 assembles a v2 .sff file's sprite table into SpriteGroups. Every
// field a Sprite needs is already present in a V2SpriteEntry, so no pixel
// data is decoded.
func loadV2(r io.ReaderAt) ([]SpriteGroup, error) {
	table, err := ParseV2(r)
	if err != nil {
		return nil, err
	}

	sprites := make([]Sprite, len(table.Sprites))
	for i, e := range table.Sprites {
		sprites[i] = Sprite{
			Group:   e.Group,
			Image:   e.Image,
			Width:   e.Width,
			Height:  e.Height,
			AxisX:   e.AxisX,
			AxisY:   e.AxisY,
			Palette: e.PaletteIndex,
		}
	}

	return groupSprites(sprites), nil
}

// groupSprites buckets sprites by Group into SpriteGroup values, preserving
// each sprite's relative order within its group, and returns the groups
// sorted by ascending group index.
func groupSprites(sprites []Sprite) []SpriteGroup {
	var order []int
	byGroup := make(map[int][]Sprite)
	for _, s := range sprites {
		if _, ok := byGroup[s.Group]; !ok {
			order = append(order, s.Group)
		}
		byGroup[s.Group] = append(byGroup[s.Group], s)
	}

	sort.Ints(order)
	groups := make([]SpriteGroup, len(order))
	for i, g := range order {
		groups[i] = SpriteGroup{Index: g, Sprites: byGroup[g]}
	}
	return groups
}
