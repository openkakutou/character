package character

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openkakutou/character/air"
	"github.com/openkakutou/character/cns"
	"github.com/openkakutou/character/def"
	"github.com/openkakutou/sff"
)

// Load reads the .def character definition file at path and assembles a
// fully populated Character: its identifying info, its referenced
// animations (.air), its referenced sprites (.sff, either on-disk version),
// and its referenced combat logic (.cns). Referenced file paths are
// resolved relative to the directory containing path itself, matching
// CharacterInfo's own documented convention and real MUGEN/Ikemen character
// folder layout.
//
// A missing or unreadable .def, .air, .sff, or .cns file returns a
// descriptive error naming which file and step failed, rather than
// panicking. See
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

	if err := validateFilesSection(path, info); err != nil {
		return nil, err
	}

	dir := filepath.Dir(path)

	animations, err := loadAnimations(dir, info.AnimationFile)
	if err != nil {
		return nil, err
	}

	sprites, err := loadSprites(dir, info.SpriteFile)
	if err != nil {
		return nil, err
	}

	stateDefs, err := loadStateDefs(dir, info.ConstantsFile)
	if err != nil {
		return nil, err
	}

	return &Character{
		Name:          info.Name,
		Author:        info.Author,
		SpriteFile:    info.SpriteFile,
		AnimationFile: info.AnimationFile,
		SoundFile:     info.SoundFile,
		CommandFile:   info.CommandFile,
		ConstantsFile: info.ConstantsFile,
		StateFiles:    info.StateFiles,
		Palettes:      info.Palettes,
		Animations:    animations,
		Sprites:       sprites,
		StateDefs:     stateDefs,
	}, nil
}

// validateFilesSection returns a descriptive error naming path when info's
// referenced file fields are all empty, which happens when a .def file has
// no [Files] section at all — most commonly an Ikemen GO
// storyboard/intro/ending screen definition (a [SceneDef]/[Scene N]-based
// file, an entirely different sub-format) that happens to sit inside a
// character's own folder alongside the real character .def. def.Parse
// correctly returns a zero-value CharacterInfo for these per its own
// documented "skip unrecognized sections" behavior, but leaving that
// unchecked means an empty referenced path later resolves to the .def's own
// containing directory (filepath.Join(dir, "")), which fails with an
// opaque "is a directory" filesystem error instead of naming the real
// problem. A .def with at least one populated [Files] entry is left
// untouched: a file missing only some entries (e.g. no cns) still gets the
// existing per-file "opening ... file" error further down the load path.
// See backlog item 051.
func validateFilesSection(path string, info def.CharacterInfo) error {
	if info.AnimationFile != "" || info.SpriteFile != "" || info.ConstantsFile != "" {
		return nil
	}
	return fmt.Errorf("character: %q doesn't look like a character definition: no [Files] section (or it has no recognized entries)", path)
}

// resolveReferencedPath joins dir with a file path referenced from inside a
// .def file (e.g. a [Files] entry), normalizing Windows-style backslash
// separators to forward slashes first. Real .def files are authored on
// Windows, where a backslash in a path is unambiguously a separator (never
// a literal filename character, which Windows itself disallows) — but
// filepath.Join treats an unnormalized backslash-containing string as a
// single literal path component on a forward-slash filesystem (Linux, and
// the browser/WASM target this library must also support), so the
// referenced subdirectory is never traversed. dir itself (derived from
// Load's own path argument, not from .def content) is left untouched: it is
// already in the host's native convention. See backlog item 049.
func resolveReferencedPath(dir, referenced string) string {
	return filepath.Join(dir, strings.ReplaceAll(referenced, "\\", "/"))
}

// openReferencedFile opens the file a .def references (dir + referenced,
// resolved via resolveReferencedPath) and, if no file exists under that
// exact casing, falls back to a case-insensitive match against the actual
// on-disk directory entries, one path component at a time, before giving
// up. Real .def files are authored on Windows, where filesystem paths are
// case-insensitive, and many reference a file under different letter
// casing than its actual on-disk name — a mismatch Linux and the
// browser/WASM target this library must also support do not tolerate. The
// exact-case path is always tried first via a direct os.Open, so a
// correctly-cased reference never pays for a directory listing. When no
// case-insensitive match exists either, the original os.Open error against
// the exact-case path is returned unchanged, preserving the existing
// descriptive "no such file or directory"-style error. See backlog item
// 050.
func openReferencedFile(dir, referenced string) (*os.File, error) {
	path := resolveReferencedPath(dir, referenced)
	f, err := os.Open(path)
	if err == nil {
		return f, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	if resolvedPath, ok := resolveCaseInsensitive(dir, strings.ReplaceAll(referenced, "\\", "/")); ok {
		if f, ferr := os.Open(resolvedPath); ferr == nil {
			return f, nil
		}
	}
	return nil, err
}

// resolveCaseInsensitive walks relPath component by component starting at
// base, matching each component against the actual directory entries at
// that level case-insensitively when no exact match exists. It returns the
// fully resolved on-disk path and true once every component (including the
// final file) has been matched, or ("", false) as soon as any component
// has no match under either casing.
func resolveCaseInsensitive(base, relPath string) (string, bool) {
	current := base
	for _, component := range strings.Split(relPath, "/") {
		if component == "" || component == "." {
			continue
		}
		candidate := filepath.Join(current, component)
		if _, err := os.Stat(candidate); err == nil {
			current = candidate
			continue
		}

		entries, err := os.ReadDir(current)
		if err != nil {
			return "", false
		}
		match := ""
		for _, entry := range entries {
			if strings.EqualFold(entry.Name(), component) {
				match = entry.Name()
				break
			}
		}
		if match == "" {
			return "", false
		}
		current = filepath.Join(current, match)
	}
	return current, true
}

// loadAnimations opens and parses the .air file dir+referenced points to.
func loadAnimations(dir, referenced string) ([]air.Animation, error) {
	path := resolveReferencedPath(dir, referenced)
	f, err := openReferencedFile(dir, referenced)
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

// loadSprites opens and loads the .sff file dir+referenced points to.
func loadSprites(dir, referenced string) ([]sff.SpriteGroup, error) {
	path := resolveReferencedPath(dir, referenced)
	f, err := openReferencedFile(dir, referenced)
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

// loadStateDefs opens and parses the .cns file dir+referenced points to.
func loadStateDefs(dir, referenced string) ([]cns.StateDef, error) {
	path := resolveReferencedPath(dir, referenced)
	f, err := openReferencedFile(dir, referenced)
	if err != nil {
		return nil, fmt.Errorf("character: opening combat logic file %q: %w", path, err)
	}
	defer f.Close()

	stateDefs, err := cns.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("character: parsing combat logic file %q: %w", path, err)
	}
	return stateDefs, nil
}
