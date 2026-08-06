// Package character is the root package of OpenKakutou's character library.
//
// It assembles the format-specific sub-packages (def, sff, air, cns — added
// incrementally, see CLAUDE.md) into a single Character struct: the unit a
// library consumer (editor, engine) actually wants to work with, rather than
// raw per-format structs.
package character

// Character is the in-memory representation of a MUGEN/Ikemen GO character,
// combining its definition, sprites, animations, and combat logic once the
// corresponding sub-packages are implemented.
type Character struct {
	Name string
}
