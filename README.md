# character

A read/write Go library for MUGEN/Ikemen GO fighting-game character files (`.def`, `.sff`, `.air`, `.cns`), built as the foundation of the [OpenKakutou](https://github.com/openkakutou) project — an open-source alternative to Fighter Factory Studio.

<!-- vibe:begin:features -->
This project is in early-stage development. Shipped so far:

- Reading MUGEN/Ikemen GO animation (`.air`) files into structured animation data — actions, frame sequences, collision boxes, and loop points
- Malformed or unusual `.air` input (bad headers, missing/negative values, comment lines, empty files) is caught with a clear, line-numbered error instead of crashing or producing wrong data
- Writing animation data back out to valid `.air` text, ready to be read by MUGEN/Ikemen GO or read back in by this library
- Loading a `.air` file and saving it back out unchanged reproduces the original file exactly, comments included — so re-saving a file you haven't edited never creates a noisy diff
- Defined the sprite and sprite group data model that will represent a character's sprites once `.sff` file reading is implemented
- Reading the header and sprite index table of MUGEN/Ikemen GO sprite sheet (`.sff`) files in their original (v1) format, locating every sprite's image data by group and image number; malformed or truncated sprite sheets are caught with a descriptive error instead of crashing
- Decoding the compressed pixel data of `.sff` v1 sprites into a plain pixel buffer with its width and height; corrupted or cut-off sprite image data is caught with a descriptive error instead of crashing
- Saving sprites back out to a valid `.sff` v1 sprite sheet file, including their pixel data, sprite-linking (sprites that reuse another sprite's image data), and palette-sharing settings; saved files load back correctly with no image data lost
- Reading the header and sprite/palette index tables of the newer, Ikemen-compatible (v2) `.sff` sprite sheet format, locating every sprite's image data and every palette bank's color data, including sprites/palette banks that reuse another one's data; malformed or truncated sprite sheets are caught with a descriptive error instead of crashing
- Decoding `.sff` v2 sprite image data — both uncompressed and PNG-encoded (indexed and true-color) — into a plain pixel buffer; an unrecognized or not-yet-supported encoding is caught with a descriptive error instead of crashing
- Saving sprites back out to a valid `.sff` v2 sprite sheet file, including uncompressed and PNG-encoded pixel data, sprite-linking, and palette bank data (with palette-linking); saved files load back correctly with no image data lost and every palette bank reference intact

Planned:

- Decoding the remaining, less common `.sff` v2 compressed pixel formats (RLE-based)
- Reading the remaining character file formats (`.def`, `.cns`) into a single, pure-data `Character` representation
- Preserving comments and ordering when the saved file actually differs from the original (today this is guaranteed only when nothing changed)
- No rendering dependency, so the library can compile to WebAssembly for web-based tooling
<!-- vibe:end:features -->

<!-- vibe:begin:install -->
Requires [Go](https://go.dev/) 1.26 or later.

```sh
go get github.com/openkakutou/character
```

Verify the install by importing the module in a Go file and running `go build`:

```go
import "github.com/openkakutou/character"
```

To update to the latest version:

```sh
go get -u github.com/openkakutou/character
```
<!-- vibe:end:install -->

<!-- vibe:begin:usage -->
Import the root package and use the `Character` type:

```go
package main

import "github.com/openkakutou/character"

func main() {
	var c character.Character
	c.Name = "Kung Fu Man"
}
```

Read a `.air` animation file into structured animation data with the `air` sub-package:

```go
package main

import (
	"fmt"
	"os"

	"github.com/openkakutou/character/air"
)

func main() {
	f, err := os.Open("kfm.air")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	animations, err := air.Parse(f)
	if err != nil {
		panic(err)
	}

	for _, a := range animations {
		fmt.Printf("action %d: %d frames\n", a.Number, len(a.Frames))
	}
}
```

Write animation data back out to `.air` text with `air.Serialize`:

```go
package main

import (
	"os"

	"github.com/openkakutou/character/air"
)

func main() {
	animations := []air.Animation{
		{
			Number: 0,
			Frames: []air.Frame{
				{Group: 0, Image: 0, X: 0, Y: 0, Time: 5},
			},
		},
	}

	f, err := os.Create("kfm.air")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := air.Serialize(f, animations); err != nil {
		panic(err)
	}
}
```

If you just want to load a `.air` file and save it back out unchanged (no data edits), use `air.ParseDocument`/`Document.Serialize` instead of `Parse`/`Serialize` — it keeps the file's comments intact:

```go
package main

import (
	"os"

	"github.com/openkakutou/character/air"
)

func main() {
	f, err := os.Open("kfm.air")
	if err != nil {
		panic(err)
	}
	doc, err := air.ParseDocument(f)
	f.Close()
	if err != nil {
		panic(err)
	}

	out, err := os.Create("kfm-copy.air")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Reproduces kfm.air exactly, including its comments.
	if err := doc.Serialize(out); err != nil {
		panic(err)
	}
}
```

Read a `.sff` v1 sprite sheet's header and sprite index table with `sff.ParseV1` — this locates each sprite's image data by (group, image) without decoding pixel data yet:

```go
package main

import (
	"fmt"
	"os"

	"github.com/openkakutou/character/sff"
)

func main() {
	f, err := os.Open("kfm.sff")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	table, err := sff.ParseV1(f)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d sprites across %d groups\n", table.Header.ImageCount, table.Header.GroupCount)

	if offset, ok := table.Offset(0, 0); ok {
		fmt.Printf("sprite (0,0) image data starts at byte %d\n", offset)
	}
}
```

Once you have a sprite's file offset and length from `table`, decode its pixel data with `sff.DecodePCX`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/openkakutou/character/sff"
)

func main() {
	f, err := os.Open("kfm.sff")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	table, err := sff.ParseV1(f)
	if err != nil {
		panic(err)
	}

	entry := table.Sprites[0]
	data := make([]byte, entry.Length)
	if _, err := f.ReadAt(data, entry.Offset); err != nil {
		panic(err)
	}

	img, err := sff.DecodePCX(data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("sprite (%d,%d) is %dx%d pixels\n", entry.Group, entry.Image, img.Width, img.Height)
}
```

Save sprites back out to a `.sff` v1 file with `sff.SerializeV1`, encoding each sprite's pixel data with `sff.EncodePCX` first:

```go
package main

import (
	"os"

	"github.com/openkakutou/character/sff"
)

func main() {
	pixels, err := sff.EncodePCX(&sff.PCXImage{
		Width: 2, Height: 2,
		Pixels: []byte{0, 0, 1, 1},
	})
	if err != nil {
		panic(err)
	}

	f, err := os.Create("kfm-copy.sff")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sprites := []sff.V1WriteSprite{
		{Group: 0, Image: 0, PixelData: pixels},
	}
	if err := sff.SerializeV1(f, [4]byte{1, 0, 0, 1}, false, sprites); err != nil {
		panic(err)
	}
}
```

Read a `.sff` v2 sprite sheet's header and sprite/palette index tables with `sff.ParseV2` — this locates each sprite's image data by (group, image) and each palette bank's color data by (group, number), without decoding pixel/color data yet:

```go
package main

import (
	"fmt"
	"os"

	"github.com/openkakutou/character/sff"
)

func main() {
	f, err := os.Open("kfm.sff")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	table, err := sff.ParseV2(f)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d sprites, %d palette banks\n", table.Header.SpriteCount, table.Header.PaletteCount)

	if offset, ok := table.Offset(0, 0); ok {
		fmt.Printf("sprite (0,0) image data starts at byte %d\n", offset)
	}
}
```

Once you have a v2 sprite's encoded pixel data (via `table.Offset` and the sprite table entry's `Length`), decode it with `sff.DecodeV2Sprite`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/openkakutou/character/sff"
)

func main() {
	f, err := os.Open("kfm.sff")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	table, err := sff.ParseV2(f)
	if err != nil {
		panic(err)
	}

	entry := table.Sprites[0]
	data := make([]byte, entry.Length)
	if _, err := f.ReadAt(data, entry.Offset); err != nil {
		panic(err)
	}

	img, err := sff.DecodeV2Sprite(entry.Format, entry.Width, entry.Height, entry.ColorDepth, data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("sprite (%d,%d) is %dx%d pixels, %d bytes per pixel\n", entry.Group, entry.Image, img.Width, img.Height, img.BytesPerPixel)
}
```

Save sprites back out to a `.sff` v2 file with `sff.SerializeV2`, encoding each sprite's pixel data with `sff.EncodeV2Sprite` first:

```go
package main

import (
	"os"

	"github.com/openkakutou/character/sff"
)

func main() {
	pixels, err := sff.EncodeV2Sprite(sff.V2FormatRaw, &sff.V2Image{
		Width: 2, Height: 2, BytesPerPixel: 1,
		Pixels: []byte{0, 0, 1, 1},
	})
	if err != nil {
		panic(err)
	}

	f, err := os.Create("kfm-copy.sff")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sprites := []sff.V2WriteSprite{
		{Group: 0, Image: 0, Width: 2, Height: 2, Format: sff.V2FormatRaw, ColorDepth: 8, PixelData: pixels},
	}
	if err := sff.SerializeV2(f, [4]byte{0, 1, 0, 2}, sprites, nil); err != nil {
		panic(err)
	}
}
```

Parsing/serialization for `.def` and `.cns` files, and decoding/encoding the remaining `.sff` v2 compressed pixel formats (RLE-based), are not implemented yet — this API surface will grow as those pieces are added.
<!-- vibe:end:usage -->

<!-- vibe:begin:docs-index -->
- [docs/api.md](docs/api.md) — the library's public API, package by package
- [docs/architecture.md](docs/architecture.md) — how the packages fit together and the read/write split
- [docs/testing.md](docs/testing.md) — how to run the test suite and what it covers
<!-- vibe:end:docs-index -->
