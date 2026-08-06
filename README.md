# character

A read/write Go library for MUGEN/Ikemen GO fighting-game character files (`.def`, `.sff`, `.air`, `.cns`), built as the foundation of the [OpenKakutou](https://github.com/openkakutou) project — an open-source alternative to Fighter Factory Studio.

<!-- vibe:begin:features -->
This project is in early-stage development — no features have shipped yet. The planned capability set covers:

- Reading MUGEN/Ikemen GO character files (`.def`, `.sff`, `.air`, `.cns`) into a single, pure-data `Character` representation
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
Import the package and use the `Character` type:

```go
package main

import "github.com/openkakutou/character"

func main() {
	var c character.Character
	c.Name = "Kung Fu Man"
}
```

Parsing/serialization for `.def`, `.sff`, `.air`, and `.cns` files is not implemented yet — this API surface will grow as those sub-packages are added.
<!-- vibe:end:usage -->

<!-- vibe:begin:docs-index -->
No additional documentation yet.
<!-- vibe:end:docs-index -->
