// Package def defines the pure-data model for MUGEN/Ikemen GO character
// definition (.def) files: CharacterInfo.
//
// This is the read-path surface — the stable vocabulary a library consumer
// (editor, engine) works with. It carries no INI parsing, file I/O, or
// write-only (format-preservation) logic; per CLAUDE.md's read/write
// separation constraint, that lives elsewhere so importing this data model
// alone never pulls in write-only dependencies.
//
// A .def file is the entry point of a MUGEN/Ikemen character: an INI-style
// text file whose [Info] section carries identifying information and whose
// [Files] section points to the other files (sprite, animation, sound,
// commands, combat logic, extra state files, palettes) that make up the
// character. CharacterInfo mirrors that shape as plain data; no parser
// populates it yet (see backlog item 016).
package def

// CharacterInfo is a MUGEN/Ikemen character definition: the identifying
// information from a .def file's [Info] section, and the file paths it
// references from its [Files] section. Paths are stored exactly as written
// in the .def file (typically relative to the character's own directory);
// resolving them against a filesystem is a concern for the package that
// eventually loads a .def file (see backlog item 018), not this model.
type CharacterInfo struct {
	// Name is the character's display name (.def [Info] "name").
	Name string `json:"name"`
	// Author is the character's author/creator (.def [Info] "author").
	Author string `json:"author"`
	// MugenVersion and IkemenVersion are the .def [Info] "mugenversion" and
	// "ikemenversion" as written ("" when absent); an engine reads them to
	// choose MUGEN or Ikemen GO behaviours.
	MugenVersion  string `json:"mugenVersion,omitempty"`
	IkemenVersion string `json:"ikemenVersion,omitempty"`
	// SpriteFile is the path to this character's sprite sheet (.sff) file
	// (.def [Files] "sprite").
	SpriteFile string `json:"spriteFile"`
	// AnimationFile is the path to this character's animation (.air) file
	// (.def [Files] "anim").
	AnimationFile string `json:"animationFile"`
	// SoundFile is the path to this character's sound (.snd) file
	// (.def [Files] "sound").
	SoundFile string `json:"soundFile"`
	// CommandFile is the path to this character's command input (.cmd)
	// file (.def [Files] "cmd").
	CommandFile string `json:"commandFile"`
	// ConstantsFile is the path to this character's main combat logic
	// (.cns) file (.def [Files] "cns").
	ConstantsFile string `json:"constantsFile"`
	// StateCommonFile is the path to this character's common state file
	// (.def [Files] "stcommon"), usually the engine's own common1.cns, which
	// is typically not inside the character folder. Ikemen GO also accepts
	// a .zss file here.
	StateCommonFile string `json:"stateCommonFile"`
	// StateFiles lists additional state definition files this character
	// includes beyond ConstantsFile (.def [Files] "st", "st1", "st2", ...),
	// in Ikemen GO's load order: "st" (no number) first, then by numeric
	// suffix ascending, ties in file order. An entry may be a .zss script.
	StateFiles []string `json:"stateFiles"`
	// Palettes lists this character's palette (.act) files, in palette
	// number order (.def [Files] "pal1", "pal2", ...).
	Palettes []string `json:"palettes"`
	// Movelists lists this character's Ikemen GO move list declarations
	// (.def [Files] "movelist", "movelist0", "movelist1", ...), one per
	// declared key with a non-empty value, ordered by Index and, within one
	// Index, with "movelist" before "movelist0", then in file order. MUGEN
	// has no move list: it is empty for a MUGEN character.
	Movelists []MovelistRef `json:"movelists"`
}

// MovelistRef is one move list declaration of a .def [Files] section: the
// key as written, the index it stands for and the file path as written.
//
// Ikemen GO treats "movelist" and "movelist0" as index 0 and "movelistN" as
// index N, N being what a ChangeMovelist state controller selects, with
// index 0 as the fallback when the selected index has no move list
// (src/char.go:3425-3452 loadMovelists, src/script.go:3996-4014
// getMovelist, commit 54d4148a728153c2097a0fb899a06812b2d3e193). For index
// 0, "movelist" is tried first and "movelist0" only when it did not load.
type MovelistRef struct {
	// Key is the key as written in the .def, letter case included.
	Key string `json:"key"`
	// Index is the move list number the key stands for.
	Index int `json:"index"`
	// File is the path of the move list text file, as written.
	File string `json:"file"`
}
