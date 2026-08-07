// Command gen regenerates the trimmed .sff test fixtures under
// sff/testdata/files from real, full-size character files. It is a
// standalone regeneration tool, not part of the sff package's public API:
// see .vibe/backlog/done/023-vendor-real-sff-test-fixtures.md for context.
//
// It is not meant to run in CI. Run it manually, pointing SRC_DIR at a
// local checkout of the source files (see testdata/README.md for where
// they come from), whenever a fixture needs to be regenerated:
//
//	SRC_DIR=/path/to/sff-extractor/tests/files go run ./sff/testdata/gen
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openkakutou/character/sff"
)

// scenario describes one trimmed fixture to produce: which real source
// file to pull from, which sprite the test needs (selected the same way
// the reference JS project's own test suite selects it: filter the sprite
// table to one group, then take the nth match), and the output filename.
type scenario struct {
	name       string // output file, under testdata/files
	src        string // source file, under SRC_DIR
	version    int    // 1 or 2
	group      int    // group filter, matching the JS test's spriteGroups option
	nthInGroup int    // index into the group-filtered list (0-based)
	// includePrevOwnPalette also copies the nearest earlier sprite that
	// owns its palette (v1 only), needed when the target sprite shares a
	// palette rather than carrying its own.
}

func main() {
	srcDir := os.Getenv("SRC_DIR")
	if srcDir == "" {
		fmt.Fprintln(os.Stderr, "SRC_DIR environment variable must point at the real .sff files (see testdata/README.md)")
		os.Exit(1)
	}
	outDir := "testdata/files"
	if _, err := os.Stat(outDir); err != nil {
		fmt.Fprintln(os.Stderr, "run this from the sff/ package directory:", err)
		os.Exit(1)
	}

	scenarios := []scenario{
		{"v1-basic.sff", "cvsryu-v1.sff", 1, 0, 0},
		{"v1-last-sprite.sff", "arale-v1.sff", 1, 10302, 0},
		{"v1-zero-length-copy.sff", "crab-v1.sff", 1, 0, 75},
		{"v1-self-link.sff", "greenarrow-v1.sff", 1, 5040, 2},
		{"v1-forward-link.sff", "cvssakura-v1.sff", 1, 5072, 0},
		{"v1-invalid-size.sff", "vivi-v1.sff", 1, 10017, 7},
		{"v1-external-palette-source.sff", "cyclops-v1.sff", 1, 0, 0},
		{"v1-same-palette-a.sff", "gray-v1.sff", 1, 0, 0},
		{"v1-same-palette-a-group5.sff", "gray-v1.sff", 1, 5, 0},
		{"v1-same-palette-b.sff", "lucifer-v1.sff", 1, 0, 0},

		{"v2-rle8.sff", "kfm-v2.sff", 2, 9000, 1},
		{"v2-lz5.sff", "kfm-v2.sff", 2, 0, 0},
		{"v2-png8.sff", "kim-v2.sff", 2, 0, 0},
		{"v2-png24.sff", "ruby-v2.sff", 2, 6053, 0},
		{"v2-png32.sff", "batman-v2.sff", 2, 9000, 2},
		{"v2-zero-length-copy.sff", "piccolo-v2.sff", 2, 186, 0},
		{"v2-png8-forced-alpha.sff", "kfm720-v2.sff", 2, 9000, 0},
		{"v2-external-palette-source.sff", "ruby-v2.sff", 2, 0, 0},
		{"v2-empty-palette-use-first.sff", "makina-v2.sff", 2, 0, 0},
		{"v2-empty-palette-use-previous.sff", "sonic-v2.sff", 2, 0, 0},
		{"v2-empty-palette-use-previous2.sff", "piccolo-sd-v2.sff", 2, 0, 0},
		{"v2-png8-b.sff", "kumagawa-v2.sff", 2, 9000, 1},
		{"v2-loadmode1.sff", "kazuki-v2.sff", 2, 3096, 14},
	}

	for _, sc := range scenarios {
		srcPath := filepath.Join(srcDir, sc.src)
		data, err := os.ReadFile(srcPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v (skipping)\n", sc.name, err)
			continue
		}

		var out []byte
		if sc.version == 1 {
			out, err = trimV1(data, sc)
		} else {
			out, err = trimV2(data, sc)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", sc.name, err)
			continue
		}

		if err := os.WriteFile(filepath.Join(outDir, sc.name), out, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "%s: writing: %v\n", sc.name, err)
			continue
		}
		fmt.Printf("%s: %d bytes (from %d)\n", sc.name, len(out), len(data))
	}
}

// resolveGroupV1 replicates the reference JS project's own sprite
// selection: it filters ParseV1's sprite table to entries whose Group
// matches, in original table order, and returns the table index of the
// nth (0-based) match.
func resolveGroupV1(table *sff.V1SpriteTable, group, nth int) (int, error) {
	n := 0
	for i, s := range table.Sprites {
		if s.Group != group {
			continue
		}
		if n == nth {
			return i, nil
		}
		n++
	}
	return 0, fmt.Errorf("group %d: only %d matching sprites, want index %d", group, n, nth)
}

func resolveGroupV2(table *sff.V2SpriteTable, group, nth int) (int, error) {
	n := 0
	for i, s := range table.Sprites {
		if s.Group != group {
			continue
		}
		if n == nth {
			return i, nil
		}
		n++
	}
	return 0, fmt.Errorf("group %d: only %d matching sprites, want index %d", group, n, nth)
}

// v1SubfileBounds returns the [start, end) byte range of sprite index i's
// own on-disk subfile (its 32-byte subheader plus everything up to the
// next sprite's subheader, or EOF for the last sprite) — the region a
// non-shared sprite's trailing embedded palette lives at the end of.
func v1SubfileBounds(table *sff.V1SpriteTable, data []byte, i int) (start, end int64) {
	start = table.Sprites[i].Offset - 32
	if i+1 < len(table.Sprites) {
		end = table.Sprites[i+1].Offset - 32
	} else {
		end = int64(len(data))
	}
	return start, end
}

// findOwnPalette walks backward from sprite index i to the nearest sprite
// (including i itself) that carries its own embedded palette
// (SharedPalette == false), and returns that sprite's real trailing
// 768-byte RGB palette block.
func findOwnPalette(table *sff.V1SpriteTable, data []byte, i int) ([]byte, error) {
	for j := i; j >= 0; j-- {
		if table.Sprites[j].SharedPalette {
			continue
		}
		_, end := v1SubfileBounds(table, data, j)
		start := end - 768
		if start < 0 || end > int64(len(data)) {
			return nil, fmt.Errorf("sprite %d: palette block out of range [%d,%d)", j, start, end)
		}
		return data[start:end], nil
	}
	return nil, fmt.Errorf("sprite %d: no earlier sprite owns a palette", i)
}

// resolvePixelOwnerV1 follows sprite index i's LinkedIndex chain (only
// consulted when Length == 0 — a sprite with its own pixel data always
// wins regardless of whatever raw value its LinkedIndex field happens to
// hold, matching sff.Load's own resolveV1Pixels) until it reaches a sprite
// that stores its own pixel data.
func resolvePixelOwnerV1(table *sff.V1SpriteTable, i int) (int, error) {
	owner := i
	seen := map[int]bool{}
	for table.Sprites[owner].Length == 0 {
		if seen[owner] {
			return 0, fmt.Errorf("sprite %d: linked chain cycles", i)
		}
		seen[owner] = true
		li := table.Sprites[owner].LinkedIndex
		if li < 0 || li >= len(table.Sprites) {
			return 0, fmt.Errorf("sprite %d: linked index %d out of range", owner, li)
		}
		owner = li
	}
	return owner, nil
}

// trimV1 builds a minimal, real-bytes-only .sff v1 file exposing exactly
// the sprite scenario sc asks for. It keeps up to three distinct real
// on-disk entries — the target sprite itself, its pixel-data owner
// (resolved via LinkedIndex when the target's own Length is 0), and its
// palette owner (resolved via SharedPalette, independently of pixel
// ownership since v1 tracks palette sharing positionally, not via an
// explicit index field) — deduplicating whichever of the three coincide,
// and orders them so the palette owner always immediately precedes the
// target with no other non-shared entry between them (palette
// inheritance is purely positional in this format).
func trimV1(data []byte, sc scenario) ([]byte, error) {
	table, err := sff.ParseV1(readerAt(data))
	if err != nil {
		return nil, err
	}

	idx, err := resolveGroupV1(table, sc.group, sc.nthInGroup)
	if err != nil {
		return nil, err
	}
	target := table.Sprites[idx]

	pixelOwner, err := resolvePixelOwnerV1(table, idx)
	if err != nil {
		return nil, err
	}

	paletteOwner := idx
	for j := idx; j >= 0; j-- {
		if !table.Sprites[j].SharedPalette {
			paletteOwner = j
			break
		}
	}

	// realEntry builds a self-contained v1Sprite for real sprite index i:
	// its own embedded real pixel bytes (following i's own pixel-owner
	// chain, since i itself might be a zero-length copy) plus its own
	// real palette bytes (via findOwnPalette).
	realEntry := func(i int) (v1Sprite, error) {
		po, err := resolvePixelOwnerV1(table, i)
		if err != nil {
			return v1Sprite{}, err
		}
		pe := table.Sprites[po]
		pal, err := findOwnPalette(table, data, i)
		if err != nil {
			return v1Sprite{}, err
		}
		e := table.Sprites[i]
		return v1Sprite{
			group: e.Group, image: e.Image, axisX: e.AxisX, axisY: e.AxisY,
			linkedIndex: -1, shared: false,
			pixel:   data[pe.Offset : pe.Offset+int64(pe.Length)],
			palette: pal,
		}, nil
	}

	var sprites []v1Sprite
	positions := map[int]int{} // real sprite index -> position in sprites

	if pixelOwner != idx && pixelOwner != paletteOwner {
		e, err := realEntry(pixelOwner)
		if err != nil {
			return nil, err
		}
		positions[pixelOwner] = len(sprites)
		sprites = append(sprites, e)
	}
	if paletteOwner != idx {
		e, err := realEntry(paletteOwner)
		if err != nil {
			return nil, err
		}
		positions[paletteOwner] = len(sprites)
		sprites = append(sprites, e)
	}

	// The target entry itself.
	te := v1Sprite{group: target.Group, image: target.Image, axisX: target.AxisX, axisY: target.AxisY}
	if pixelOwner == idx {
		pe := table.Sprites[idx]
		te.linkedIndex = -1
		te.pixel = data[pe.Offset : pe.Offset+int64(pe.Length)]
	} else {
		te.linkedIndex = positions[pixelOwner]
	}
	if paletteOwner == idx {
		te.shared = false
		pal, err := findOwnPalette(table, data, idx)
		if err != nil {
			return nil, err
		}
		te.palette = pal
	} else {
		te.shared = true
	}
	sprites = append(sprites, te)

	return encodeV1(table.Header, sprites)
}

// v1Sprite is the trimming tool's own minimal per-sprite write shape,
// carrying real bytes copied verbatim from a source file.
type v1Sprite struct {
	group, image, axisX, axisY int
	linkedIndex                int // -1 = none, own pixel+palette
	shared                     bool
	pixel                      []byte
	palette                    []byte
}

// encodeV1 hand-writes a minimal, valid .sff v1 file: a 512-byte header
// followed by one 32-byte subheader + pixel data (+ trailing 768-byte
// palette, for sprites that carry their own) per sprite, chained via each
// subheader's next-subfile offset. This bypasses SerializeV1 because
// V1WriteSprite has no field for embedding a sprite's own real palette
// bytes (that capability doesn't exist yet — see backlog item 026); this
// tool needs the real bytes preserved verbatim, not decoded/re-encoded.
func encodeV1(h sff.V1Header, sprites []v1Sprite) ([]byte, error) {
	var buf bytes.Buffer
	header := make([]byte, 512)
	copy(header[0:12], []byte("ElecbyteSpr\x00"))
	copy(header[12:16], h.Version[:])
	binary.LittleEndian.PutUint32(header[16:20], uint32(len(sprites)))
	binary.LittleEndian.PutUint32(header[20:24], uint32(len(sprites)))
	binary.LittleEndian.PutUint32(header[24:28], 512)
	if h.SharedPalette {
		header[32] = 1
	}
	buf.Write(header)

	offsets := make([]int, len(sprites))
	pos := 512
	for i, s := range sprites {
		offsets[i] = pos
		pos += 32 + len(s.pixel)
		if !s.shared {
			pos += len(s.palette)
		}
	}

	for i, s := range sprites {
		sub := make([]byte, 32)
		if i+1 < len(sprites) {
			binary.LittleEndian.PutUint32(sub[0:4], uint32(offsets[i+1]))
		}
		binary.LittleEndian.PutUint32(sub[4:8], uint32(len(s.pixel)))
		binary.LittleEndian.PutUint16(sub[8:10], uint16(int16(s.axisX)))
		binary.LittleEndian.PutUint16(sub[10:12], uint16(int16(s.axisY)))
		binary.LittleEndian.PutUint16(sub[12:14], uint16(s.group))
		binary.LittleEndian.PutUint16(sub[14:16], uint16(s.image))
		li := s.linkedIndex
		if li < 0 {
			li = 0
		}
		binary.LittleEndian.PutUint16(sub[16:18], uint16(li))
		if s.shared {
			sub[18] = 1
		}
		buf.Write(sub)
		buf.Write(s.pixel)
		if !s.shared {
			buf.Write(s.palette)
		}
	}

	return buf.Bytes(), nil
}

func trimV2(data []byte, sc scenario) ([]byte, error) {
	table, err := sff.ParseV2(readerAt(data))
	if err != nil {
		return nil, err
	}

	idx, err := resolveGroupV2(table, sc.group, sc.nthInGroup)
	if err != nil {
		return nil, err
	}
	target := table.Sprites[idx]

	pixelOwner := idx
	seen := map[int]bool{}
	for table.Sprites[pixelOwner].Length == 0 {
		if seen[pixelOwner] {
			return nil, fmt.Errorf("sprite %d: linked chain cycles", idx)
		}
		seen[pixelOwner] = true
		li := table.Sprites[pixelOwner].LinkedIndex
		if li < 0 || li >= len(table.Sprites) {
			return nil, fmt.Errorf("sprite %d: linked index %d out of range", pixelOwner, li)
		}
		pixelOwner = li
	}
	pe := table.Sprites[pixelOwner]
	pixelBytes := data[pe.Offset : pe.Offset+int64(pe.Length)]

	// Resolve the palette bank's real color bytes, following its own
	// LinkedIndex chain when it stores no color data of its own.
	pi := target.PaletteIndex
	if pi < 0 || pi >= len(table.Palettes) {
		return nil, fmt.Errorf("sprite %d: palette index %d out of range", idx, pi)
	}
	palOwner := pi
	seenP := map[int]bool{}
	for table.Palettes[palOwner].Length == 0 && len(table.Palettes) > 0 {
		if seenP[palOwner] {
			return nil, fmt.Errorf("palette %d: linked chain cycles", pi)
		}
		seenP[palOwner] = true
		li := table.Palettes[palOwner].LinkedIndex
		if li == 0 && palOwner == 0 {
			break
		}
		if li < 0 || li >= len(table.Palettes) {
			li = 0
		}
		if li == palOwner {
			break
		}
		palOwner = li
	}
	pal := table.Palettes[palOwner]
	var paletteColors []byte
	if pal.Length > 0 {
		paletteColors = data[pal.Offset : pal.Offset+int64(pal.Length)]
	} else {
		// Truly empty file-wide: shouldn't happen for a real fixture.
		paletteColors = make([]byte, pal.ColorCount*4)
	}

	writeSprites := []sff.V2WriteSprite{{
		Group: pe.Group, Image: pe.Image, Width: pe.Width, Height: pe.Height,
		AxisX: pe.AxisX, AxisY: pe.AxisY, Format: pe.Format, ColorDepth: pe.ColorDepth,
		PaletteIndex: 0, PixelData: pixelBytes,
	}}
	if pixelOwner != idx {
		writeSprites = append(writeSprites, sff.V2WriteSprite{
			Group: target.Group, Image: target.Image,
			PaletteIndex: 0, LinkedIndex: 0,
		})
	}

	var buf bytes.Buffer
	err = sff.SerializeV2(&buf, table.Header.Version, writeSprites, []sff.V2WritePalette{{
		Group: pal.Group, Number: pal.Number, ColorCount: pal.ColorCount, ColorData: paletteColors,
	}})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// readerAt adapts a byte slice to io.ReaderAt.
type readerAtBytes []byte

func (r readerAtBytes) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(r)) {
		return 0, fmt.Errorf("offset %d past end of data (len %d)", off, len(r))
	}
	n := copy(p, r[off:])
	if n < len(p) {
		return n, fmt.Errorf("short read at offset %d: got %d, want %d", off, n, len(p))
	}
	return n, nil
}

func readerAt(data []byte) readerAtBytes { return readerAtBytes(data) }
