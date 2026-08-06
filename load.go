package character

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openkakutou/character/air"
	"github.com/openkakutou/character/def"
	"github.com/openkakutou/character/sff"
)

// Load reads the .def character definition file at path and assembles a
// fully populated Character: its identifying info, its referenced
// animations (.air), and its referenced sprites (.sff, either on-disk
// version). Referenced file paths are resolved relative to the directory
// containing path itself, matching CharacterInfo's own documented
// convention and real MUGEN/Ikemen character folder layout.
//
// A missing or unreadable .def, .air, or .sff file returns a descriptive
// error naming which file and step failed, rather than panicking. See
// .vibe/decisions/010-def-loader-assembles-character-from-referenced-files.md.
func Load(path string) (*Character, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("character: opening character definition file %q: %w", path, err)
	}
	defer f.Close()

	info, err := def.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("character: parsing character definition file %q: %w", path, err)
	}

	dir := filepath.Dir(path)

	animations, err := loadAnimations(filepath.Join(dir, info.AnimationFile))
	if err != nil {
		return nil, err
	}

	sprites, err := loadSprites(filepath.Join(dir, info.SpriteFile))
	if err != nil {
		return nil, err
	}

	return &Character{
		Name:       info.Name,
		Animations: animations,
		Sprites:    sprites,
	}, nil
}

// loadAnimations opens and parses the .air file at path.
func loadAnimations(path string) ([]air.Animation, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("character: opening animation file %q: %w", path, err)
	}
	defer f.Close()

	animations, err := air.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("character: parsing animation file %q: %w", path, err)
	}
	return animations, nil
}

// loadSprites opens and loads the .sff file at path.
func loadSprites(path string) ([]sff.SpriteGroup, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("character: opening sprite file %q: %w", path, err)
	}
	defer f.Close()

	sprites, err := sff.Load(f)
	if err != nil {
		return nil, fmt.Errorf("character: loading sprite file %q: %w", path, err)
	}
	return sprites, nil
}
