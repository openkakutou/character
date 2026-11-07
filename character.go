// Package character is the root package of OpenKakutou's character library.
//
// It assembles the format-specific sub-packages (def, sff, air, cns — added
// incrementally, see CLAUDE.md) into a single Character struct: the unit a
// library consumer (editor, engine) actually wants to work with, rather than
// raw per-format structs.
package character

import (
	"github.com/openkakutou/character/air"
	"github.com/openkakutou/character/cmd"
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
	// StateFiles lists additional state files (.cns or .zss) beyond
	// ConstantsFile, and Palettes lists palette (.act) files — both from
	// the .def [Files] section, see def.CharacterInfo.
	StateFiles []string `json:"stateFiles"`
	// StateCommonFile is the .def "stcommon" entry, see def.CharacterInfo.
	StateCommonFile string            `json:"stateCommonFile"`
	Palettes        []string          `json:"palettes"`
	Animations      []air.Animation   `json:"animations"`
	Sprites         []sff.SpriteGroup `json:"sprites"`
	StateDefs       []cns.StateDef    `json:"stateDefs"`
	// Scripts holds every .zss state script referenced by StateFiles or
	// StateCommonFile, in Ikemen GO's load order, each with its own
	// non-fatal Error. It is empty for a MUGEN character, which has no
	// .zss. See .vibe/decisions/032.
	Scripts []ScriptFile `json:"scripts"`
	// Sounds is the character's decoded sound effects, resolved from
	// SoundFile via the external github.com/openkakutou/snd module.
	// Unlike Sprites/Animations/StateDefs, SoundFile is optional: it is
	// empty (rather than an error) for a character with no SoundFile
	// declared at all — see .vibe/decisions/029.
	Sounds []SoundGroup `json:"sounds"`
	// Commands is the character's parsed input command file, resolved from
	// CommandFile (Load) or the cmdBytes argument (LoadBytes). It is the
	// normalized empty value (no nil slice/map) when no .cmd was declared or
	// supplied, or when it could not be read — see CommandError and
	// .vibe/decisions/031.
	Commands cmd.CommandFile `json:"commands"`
	// CommandError is non-empty when a .cmd file was declared or supplied
	// but could not be opened or parsed. It is non-fatal: the load still
	// succeeds, mirroring Sound.Error (.vibe/decisions/030).
	CommandError string `json:"commandError,omitempty"`
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
