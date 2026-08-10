// Package character is the root package of OpenKakutou's character library.
//
// It assembles the format-specific sub-packages (def, sff, air, cns — added
// incrementally, see CLAUDE.md) into a single Character struct: the unit a
// library consumer (editor, engine) actually wants to work with, rather than
// raw per-format structs.
package character

import (
	"github.com/openkakutou/character/air"
	"github.com/openkakutou/character/cns"
	"github.com/openkakutou/sff"
)

// Character is the in-memory representation of a MUGEN/Ikemen GO character,
// combining its definition, sprites, animations, and combat logic.
//
// Animations, Sprites, and StateDefs are exposed through the air/sff/cns
// read-path types only (air.Animation, sff.SpriteGroup, cns.StateDef) —
// never a write-only, format-preservation type — per CLAUDE.md's read/write
// separation constraint.
type Character struct {
	Name string `json:"name"`
	// Author is the character's author/creator, from the .def [Info]
	// section's "author" key. Empty when the source .def doesn't set it.
	Author string `json:"author"`
	// SpriteFile, AnimationFile, SoundFile, CommandFile, and ConstantsFile
	// are the referenced file paths from the .def [Files] section, exactly
	// as written there (see def.CharacterInfo). They are metadata only —
	// LoadBytes does not use them to locate the animations/sprites/state
	// defs it was handed directly as byte buffers, and Load already
	// resolved them against the filesystem before returning.
	SpriteFile    string `json:"spriteFile"`
	AnimationFile string `json:"animationFile"`
	SoundFile     string `json:"soundFile"`
	CommandFile   string `json:"commandFile"`
	ConstantsFile string `json:"constantsFile"`
	// StateFiles lists additional state definition (.st) files beyond
	// ConstantsFile, and Palettes lists palette (.act) files — both from
	// the .def [Files] section, see def.CharacterInfo.
	StateFiles []string          `json:"stateFiles"`
	Palettes   []string          `json:"palettes"`
	Animations []air.Animation   `json:"animations"`
	Sprites    []sff.SpriteGroup `json:"sprites"`
	StateDefs  []cns.StateDef    `json:"stateDefs"`
}

// ResolveSprite returns the Sprite that frame's (Group, Image) reference
// names within c.Sprites. It returns a descriptive error, rather than a
// zero Sprite, when no loaded sprite matches — including when c.Sprites is
// empty (e.g. a zero-value Character). Resolution is delegated to
// air.SpriteResolver, the single place the air/sff frame-to-sprite lookup
// is implemented; see .vibe/decisions/008-air-sprite-resolution-lives-in-air-package.md.
func (c *Character) ResolveSprite(frame air.Frame) (sff.Sprite, error) {
	return air.NewSpriteResolver(c.Sprites).Resolve(frame)
}
