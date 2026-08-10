# character

A read/write Go library for MUGEN/Ikemen GO fighting-game character files (`.def`, `.sff`, `.air`, `.cns`), built as the foundation of the [OpenKakutou](https://github.com/openkakutou) project — an open-source alternative to Fighter Factory Studio.

<!-- vibe:begin:features -->
This project is in early-stage development. Shipped so far:

- Reading MUGEN/Ikemen GO animation (`.air`) files into structured animation data — actions, frame sequences, collision boxes, and loop points
- Malformed or unusual `.air` input (bad headers, missing/negative values, comment lines, empty files) is caught with a clear, line-numbered error instead of crashing or producing wrong data
- Frames using the "no sprite shown" convention widely found in real MUGEN/Ikemen GO characters (any negative group/image value, not just `-1,-1`) are read successfully instead of being rejected, and resolve to no sprite instead of a failed lookup
- Writing animation data back out to valid `.air` text, ready to be read by MUGEN/Ikemen GO or read back in by this library
- Loading a `.air` file and saving it back out unchanged reproduces the original file exactly, comments included — so re-saving a file you haven't edited never creates a noisy diff
- Defined the sprite and sprite group data model that will represent a character's sprites once `.sff` file reading is implemented
- Reading the header and sprite index table of MUGEN/Ikemen GO sprite sheet (`.sff`) files in their original (v1) format, locating every sprite's image data by group and image number; malformed or truncated sprite sheets are caught with a descriptive error instead of crashing
- Decoding the compressed pixel data of `.sff` v1 sprites into a plain pixel buffer with its width and height; corrupted or cut-off sprite image data is caught with a descriptive error instead of crashing
- Saving sprites back out to a valid `.sff` v1 sprite sheet file, including their pixel data, sprite-linking (sprites that reuse another sprite's image data), and palette-sharing settings; saved files load back correctly with no image data lost
- Reading the header and sprite/palette index tables of the newer, Ikemen-compatible (v2) `.sff` sprite sheet format, locating every sprite's image data and every palette bank's color data, including sprites/palette banks that reuse another one's data; malformed or truncated sprite sheets are caught with a descriptive error instead of crashing
- Decoding `.sff` v2 sprite image data — uncompressed, run-length-compressed ("RLE8"), dictionary-compressed ("LZ5"), and PNG-encoded (indexed and true-color) — into a plain pixel buffer; an unrecognized or not-yet-supported encoding is caught with a descriptive error instead of crashing
- Saving sprites back out to a valid `.sff` v2 sprite sheet file, including uncompressed and PNG-encoded pixel data, sprite-linking, and palette bank data (with palette-linking); saved files load back correctly with no image data lost and every palette bank reference intact
- Matching each animation frame to its actual sprite image from a loaded sprite sheet (either `.sff` version, no version-specific handling needed); a frame pointing to a sprite that doesn't exist is caught with a clear error instead of silently showing nothing
- Assembling a character's animations and sprites into a single `Character`, so you can look up the actual sprite shown by any animation frame directly from it
- Defined the character definition data model that will represent a character's name, author, and the file paths it references (sprite, animation, sound, commands, combat logic, additional states, and palettes) once `.def` file reading is implemented
- Reading MUGEN/Ikemen GO character definition (`.def`) files into that data model: name, author, and every referenced file (sprite, animation, sound, commands, combat logic, additional states, palettes); sections the library doesn't recognize are skipped instead of aborting the read, and a malformed line is caught with a clear, line-numbered error instead of crashing
- Writing character definitions back out to valid `.def` text, ready to be read by MUGEN/Ikemen GO or read back in by this library
- Loading a `.def` file and saving it back out unchanged reproduces the original file exactly, including comments, section ordering, and any sections the library doesn't otherwise recognize — so re-saving a file you haven't edited never creates a noisy diff
- Loading a full character in one step from its `.def` file — its name, animations, sprites (either `.sff` version), and combat logic are automatically read from the files it references and assembled into a ready-to-use `Character`; a missing or unreadable referenced file is caught with a clear error instead of crashing
- Reading MUGEN/Ikemen GO combat logic (`.cns`) files into structured state data — every state and the behaviors it runs, with their trigger conditions and parameters kept as raw data rather than evaluated; sections the library doesn't recognize are skipped instead of aborting the read, and a malformed state or behavior header is caught with a clear, line-numbered error instead of crashing
- Writing combat logic data back out to valid `.cns` text, ready to be read by MUGEN/Ikemen GO or read back in by this library
- Loading a `.cns` file and saving it back out unchanged reproduces the original file exactly, including comments, block ordering, and any sections the library doesn't otherwise recognize — so re-saving a file you haven't edited never creates a noisy diff
- Reading real-world `.cns` files that write a behavior header with no number at all (a common authoring style) instead of rejecting them, and reading a state header field given a formula instead of a plain number (its original formula text is kept alongside the state) instead of failing to load — while a genuinely malformed file is still caught with a clear, line-numbered error
- Reading real-world `.cns` files that write a state header with its closing bracket missing — a common authoring typo — instead of rejecting them, while a bracket line unrelated to a state header is still caught with a clear, line-numbered error
- Resolving a decoded sprite's pixel data into its actual on-screen colors using the sprite's palette — including reading a `.sff` v1 sprite's own embedded palette and following `.sff` v2 palette bank sharing/linking — with the correct transparency rule applied depending on how the sprite was originally encoded
- Recoloring a sprite with an external `.act` palette file instead of its own — reading the `.act` file's colors in the correct order with the correct transparency, and using it in place of the sprite's own palette when resolving its on-screen colors; a wrongly-sized `.act` file is caught with a descriptive error instead of crashing
- Resolving a `.sff` v1 sprite's actual decoded pixel data through the same public interface as its palette, validated against real, unmodified MUGEN/Ikemen character files rather than only hand-built test data — this caught and fixed two decoding inaccuracies affecting some real sprites: a sprite whose stored "linked sprite" reference points at itself or a later sprite now correctly falls back to its own image, and a sprite's own color palette is now located correctly for every real file layout, not just the common case; a sprite with a corrupted or nonsensical declared size falls back to a blank placeholder image instead of risking a crash

- A character (name, animations, sprites, combat logic) can now be loaded directly in a web browser via a WebAssembly build — no local Go installation needed; tagging a new release automatically publishes a downloadable module ready for a web app to load
- A sprite's actual image can now be resolved directly in the browser via that same WebAssembly build — not just its dimensions and metadata — including recoloring it with an external palette file; multiple sprites can be resolved in a single call so browsing a whole sprite sheet doesn't require one round trip per sprite
- `.sff` v2 sprites that carry no image data of their own — a shorthand some real character files use to mean "reuse the character's very first sprite" — now resolve to that actual image instead of reporting an error, validated against real, unmodified MUGEN/Ikemen character files rather than only hand-built test data; that same validation also caught and fixed two color-accuracy bugs affecting some real sprites: `.sff` v2 sprites using the 32-bit PNG format now show correct colors wherever they're partially see-through, and `.sff` v2 sprites using the 8-bit indexed PNG format now always show fully opaque colors outside their transparent areas
- `.sff` v2 sprites using the less common "RLE5" compressed pixel format can now be resolved too, alongside the other supported encodings — unlike those, this one hasn't yet been validated against a real character file (none using it has been found so far), so treat it as unproven until one turns up

Planned:

- Preserving comments and ordering when the saved `.def`/`.air`/`.cns` file actually differs from the original (today this is guaranteed only when nothing changed)
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

	"github.com/openkakutou/sff"
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

	"github.com/openkakutou/sff"
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

Resolve a decoded sprite's actual on-screen colors from its palette with `sff.ResolvePixels` — this works the same for `.sff` v1 (`sff.ResolveV1Palette`) and v2 (`sff.ResolveV2Palette`) sprites, only the palette lookup differs:

```go
package main

import (
	"fmt"
	"os"

	"github.com/openkakutou/sff"
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

	palette, err := sff.ResolveV1Palette(table, f, 0, nil)
	if err != nil {
		panic(err)
	}

	// PCX-decoded sprites use the "index 0 is always transparent" rule.
	pixels := sff.ResolvePixels(img.Pixels, palette, sff.AlphaForceTransparentAtIndexZero)
	fmt.Printf("pixel (0,0) color: %v\n", pixels[0])
}
```

Recolor a sprite with an external `.act` palette file instead of its own: decode it with `sff.DecodeExternalPalette`, then pass it as the last argument to `sff.ResolveV1Palette`/`sff.ResolveV2Palette` in place of `nil`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/openkakutou/sff"
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

	actBytes, err := os.ReadFile("kfm-alt.act")
	if err != nil {
		panic(err)
	}
	altPalette, err := sff.DecodeExternalPalette(actBytes)
	if err != nil {
		panic(err)
	}

	// Same sprite, but resolved against the alternate palette instead of its own.
	palette, err := sff.ResolveV1Palette(table, f, 0, &altPalette)
	if err != nil {
		panic(err)
	}
	fmt.Printf("recolored palette entry 1: %v\n", palette[1])
}
```

Save sprites back out to a `.sff` v1 file with `sff.SerializeV1`, encoding each sprite's pixel data with `sff.EncodePCX` first:

```go
package main

import (
	"os"

	"github.com/openkakutou/sff"
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

	"github.com/openkakutou/sff"
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

	"github.com/openkakutou/sff"
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

	"github.com/openkakutou/sff"
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

Match an animation's frames to their actual sprites with `air.NewSpriteResolver`, built from the sprite groups you loaded via `sff` — it works the same whether those sprites came from a `.sff` v1 or v2 file:

```go
package main

import (
	"fmt"

	"github.com/openkakutou/character/air"
	"github.com/openkakutou/sff"
)

func main() {
	animations := []air.Animation{
		{Number: 0, Frames: []air.Frame{{Group: 0, Image: 0, Time: 5}}},
	}
	spriteGroups := []sff.SpriteGroup{
		{Index: 0, Sprites: []sff.Sprite{{Group: 0, Image: 0, Width: 64, Height: 128}}},
	}

	resolver := air.NewSpriteResolver(spriteGroups)

	for _, frame := range animations[0].Frames {
		sprite, err := resolver.Resolve(frame)
		if err != nil {
			// e.g. the frame references a sprite that isn't in spriteGroups.
			panic(err)
		}
		fmt.Printf("frame shows sprite (%d,%d), %dx%d pixels\n", sprite.Group, sprite.Image, sprite.Width, sprite.Height)
	}
}
```

Assemble a full `Character` from animations and sprites you've already loaded, then look up the sprite shown by any frame directly from it with `Character.ResolveSprite`:

```go
package main

import (
	"fmt"

	"github.com/openkakutou/character"
	"github.com/openkakutou/character/air"
	"github.com/openkakutou/sff"
)

func main() {
	c := character.Character{
		Name: "Kung Fu Man",
		Animations: []air.Animation{
			{Number: 0, Frames: []air.Frame{{Group: 0, Image: 0, Time: 5}}},
		},
		Sprites: []sff.SpriteGroup{
			{Index: 0, Sprites: []sff.Sprite{{Group: 0, Image: 0, Width: 64, Height: 128}}},
		},
	}

	for _, frame := range c.Animations[0].Frames {
		sprite, err := c.ResolveSprite(frame)
		if err != nil {
			// e.g. the frame references a sprite that isn't in c.Sprites.
			panic(err)
		}
		fmt.Printf("frame shows sprite (%d,%d), %dx%d pixels\n", sprite.Group, sprite.Image, sprite.Width, sprite.Height)
	}
}
```

Read a `.def` character definition file into a `CharacterInfo` with `def.Parse`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/openkakutou/character/def"
)

func main() {
	f, err := os.Open("kfm.def")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	info, err := def.Parse(f)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s by %s, sprites in %s\n", info.Name, info.Author, info.SpriteFile)
}
```

Write a `CharacterInfo` back out to `.def` text with `def.Serialize`:

```go
package main

import (
	"os"

	"github.com/openkakutou/character/def"
)

func main() {
	info := def.CharacterInfo{
		Name:       "Kung Fu Man",
		Author:     "Elecbyte",
		SpriteFile: "kfm.sff",
	}

	f, err := os.Create("kfm.def")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := def.Serialize(f, info); err != nil {
		panic(err)
	}
}
```

If you just want to load a `.def` file and save it back out unchanged (no data edits), use `def.ParseDocument`/`Document.Serialize` instead of `Parse`/`Serialize` — it keeps the file's comments, section ordering, and unrecognized sections intact:

```go
package main

import (
	"os"

	"github.com/openkakutou/character/def"
)

func main() {
	f, err := os.Open("kfm.def")
	if err != nil {
		panic(err)
	}
	doc, err := def.ParseDocument(f)
	f.Close()
	if err != nil {
		panic(err)
	}

	out, err := os.Create("kfm-copy.def")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Reproduces kfm.def exactly, including its comments.
	if err := doc.Serialize(out); err != nil {
		panic(err)
	}
}
```

Load a full character directly from its `.def` file with `character.Load` — this reads the `.def`, resolves and reads its referenced `.air`/`.sff`/`.cns` files, and hands you back a ready-to-use `Character`:

```go
package main

import (
	"fmt"

	"github.com/openkakutou/character"
)

func main() {
	c, err := character.Load("kfm.def")
	if err != nil {
		// e.g. the .def file, or a file it references, is missing or unreadable.
		panic(err)
	}

	fmt.Printf("%s: %d animations, %d sprite groups, %d states\n", c.Name, len(c.Animations), len(c.Sprites), len(c.StateDefs))

	for _, frame := range c.Animations[0].Frames {
		sprite, err := c.ResolveSprite(frame)
		if err != nil {
			panic(err)
		}
		fmt.Printf("frame shows sprite (%d,%d), %dx%d pixels\n", sprite.Group, sprite.Image, sprite.Width, sprite.Height)
	}
}
```

Read a `.cns` combat logic file into its states with `cns.Parse`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/openkakutou/character/cns"
)

func main() {
	f, err := os.Open("kfm.cns")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	states, err := cns.Parse(f)
	if err != nil {
		panic(err)
	}

	for _, s := range states {
		fmt.Printf("state %d: %d controllers\n", s.Number, len(s.Controllers))
	}
}
```

Write states back out to `.cns` text with `cns.Serialize`:

```go
package main

import (
	"os"

	"github.com/openkakutou/character/cns"
)

func main() {
	states := []cns.StateDef{
		{
			Number:   0,
			Type:     cns.StateTypeStanding,
			MoveType: cns.MoveTypeIdle,
			Physics:  cns.PhysicsStanding,
			Ctrl:     true,
		},
	}

	f, err := os.Create("kfm.cns")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := cns.Serialize(f, states); err != nil {
		panic(err)
	}
}
```

If you just want to load a `.cns` file and save it back out unchanged (no data edits), use `cns.ParseDocument`/`Document.Serialize` instead of `Parse`/`Serialize` — it keeps the file's comments, block ordering, and unrecognized sections intact:

```go
package main

import (
	"os"

	"github.com/openkakutou/character/cns"
)

func main() {
	f, err := os.Open("kfm.cns")
	if err != nil {
		panic(err)
	}
	doc, err := cns.ParseDocument(f)
	f.Close()
	if err != nil {
		panic(err)
	}

	out, err := os.Create("kfm-copy.cns")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Reproduces kfm.cns exactly, including its comments.
	if err := doc.Serialize(out); err != nil {
		panic(err)
	}
}
```

Wiring `.cns` into `Character`, and decoding/encoding the remaining `.sff` v2 compressed pixel formats (RLE-based), are not implemented yet — this API surface will grow as those pieces are added.

### Loading a character in a web browser (WebAssembly)

A web app with no Go toolchain of its own can load a character too, using a pre-built WebAssembly module downloaded from a tagged release's assets (`character.wasm` + `wasm_exec.js`):

```html
<script src="wasm_exec.js"></script>
<script>
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("character.wasm"), go.importObject)
    .then((result) => go.run(result.instance))
    .then(async () => {
      const [defBytes, airBytes, sffBytes, cnsBytes] = await Promise.all(
        ["kfm.def", "kfm.air", "kfm.sff", "kfm.cns"].map(
          (name) => fetch(name).then((r) => r.arrayBuffer()).then((buf) => new Uint8Array(buf)),
        ),
      );

      const result = globalThis.OpenKakutouCharacter.load(defBytes, airBytes, sffBytes, cnsBytes);
      if (result.error) {
        throw new Error(result.error);
      }

      const character = JSON.parse(result.character);
      console.log(`${character.name}: ${character.animations.length} animations`);
    });
</script>
```

See [docs/wasm.md](docs/wasm.md) for the full JS API contract and how to build the module locally.
<!-- vibe:end:usage -->

<!-- vibe:begin:docs-index -->
- [docs/api.md](docs/api.md) — the library's public API, package by package
- [docs/architecture.md](docs/architecture.md) — how the packages fit together and the read/write split
- [docs/testing.md](docs/testing.md) — how to run the test suite and what it covers
- [docs/wasm.md](docs/wasm.md) — the WebAssembly entrypoint's JS API, how to build it locally, and the release pipeline that publishes it
<!-- vibe:end:docs-index -->
