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
	// StateFiles lists additional state definition (.st) files this
	// character includes beyond ConstantsFile (.def [Files] "st", "st1",
	// "st2", ...), in the order they appear in the .def file.
	StateFiles []string `json:"stateFiles"`
	// Palettes lists this character's palette (.act) files, in palette
	// number order (.def [Files] "pal1", "pal2", ...).
	Palettes []string `json:"palettes"`
}
