# character

A read/write Go library for MUGEN/Ikemen GO fighting-game character files (`.def`, `.sff`, `.air`, `.cns`), built as the foundation of the [OpenKakutou](https://github.com/openkakutou) project — an open-source alternative to Fighter Factory Studio.

<!-- vibe:begin:features -->
This project is in early-stage development. Shipped so far:

- Reading MUGEN/Ikemen GO animation (`.air`) files into structured animation data — actions, frame sequences, collision boxes, and loop points

Planned:

- Reading the remaining character file formats (`.def`, `.sff`, `.cns`) into a single, pure-data `Character` representation
- Writing changes back out while preserving the original file structure (ordering, comments), so edits produce clean, reviewable diffs
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

Parsing/serialization for `.def`, `.sff`, and `.cns` files is not implemented yet, and `.air` writing is not implemented yet either — this API surface will grow as those pieces are added.
<!-- vibe:end:usage -->

<!-- vibe:begin:docs-index -->
- [docs/api.md](docs/api.md) — the library's public API, package by package
- [docs/architecture.md](docs/architecture.md) — how the packages fit together and the read/write split
- [docs/testing.md](docs/testing.md) — how to run the test suite and what it covers
<!-- vibe:end:docs-index -->
