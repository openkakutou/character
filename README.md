# character

A read/write Go library for MUGEN/Ikemen GO fighting-game character files (`.def`, `.sff`, `.air`, `.cns`), built as the foundation of the [OpenKakutou](https://github.com/openkakutou) project — an open-source alternative to Fighter Factory Studio.

<!-- vibe:begin:features -->
This project is in early-stage development. Shipped so far:

- Reading MUGEN/Ikemen GO animation (`.air`) files into structured animation data — actions, frame sequences, collision boxes, and loop points
- Malformed or unusual `.air` input (bad headers, missing/negative values, comment lines, empty files) is caught with a clear, line-numbered error instead of crashing or producing wrong data
- Writing animation data back out to valid `.air` text, ready to be read by MUGEN/Ikemen GO or read back in by this library
- Loading a `.air` file and saving it back out unchanged reproduces the original file exactly, comments included — so re-saving a file you haven't edited never creates a noisy diff

Planned:

- Reading the remaining character file formats (`.def`, `.sff`, `.cns`) into a single, pure-data `Character` representation
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

Parsing/serialization for `.def`, `.sff`, and `.cns` files is not implemented yet — this API surface will grow as those pieces are added.
<!-- vibe:end:usage -->

<!-- vibe:begin:docs-index -->
- [docs/api.md](docs/api.md) — the library's public API, package by package
- [docs/architecture.md](docs/architecture.md) — how the packages fit together and the read/write split
- [docs/testing.md](docs/testing.md) — how to run the test suite and what it covers
<!-- vibe:end:docs-index -->
