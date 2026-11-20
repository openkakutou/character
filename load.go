package character

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/openkakutou/character/air"
	"github.com/openkakutou/character/cmd"
	"github.com/openkakutou/character/cns"
	"github.com/openkakutou/character/def"
	"github.com/openkakutou/sff"
)

// Load reads the .def character definition file at path and assembles a
// fully populated Character: its identifying info, its referenced
// animations (.air), its referenced sprites (.sff, either on-disk version),
// its referenced combat logic (.cns, optional when states come from .zss),
// and its .zss state scripts (st, stN, stcommon; Character.Scripts).
// Referenced file paths are resolved relative to the directory containing path itself, matching
// CharacterInfo's own documented convention and real MUGEN/Ikemen character
// folder layout.
//
// A missing or unreadable .def, .air or .sff file, or a declared .cns that
// is missing or unreadable (an empty cns entry is allowed), returns a
// descriptive error naming which file and step failed, rather than
// panicking. A missing or unparseable .cmd file is the exception: it is
// non-fatal and reported in Character.CommandError instead (see
// .vibe/decisions/031); so is a missing or unparseable .zss, reported in its
// ScriptFile.Error (.vibe/decisions/032); and a missing or unreadable move
// list file (Ikemen GO's "movelist" keys), reported in its Movelists entry's
// Error. See
// .vibe/decisions/010-def-loader-assembles-character-from-referenced-files.md.
func Load(path string) (*Character, error) {
	return loadFrom(osFiles{}, path, filepath.Dir(path))
}

// LoadFS does what Load does for a .def stored in fsys instead of on disk:
// name is the slash-separated path of the .def inside fsys, and every file
// it references (.air, .sff, .snd, .cmd, .cns, .zss, ...) is resolved
// relative to the .def's directory, with Windows backslashes accepted and
// the case-insensitive per-component fallback of backlog item 050. A
// reference that leaves the root of fsys (a leading "..") is reported as a
// missing file. Errors follow Load's. See backlog item 084.
func LoadFS(fsys fs.FS, name string) (*Character, error) {
	return loadFrom(fsFiles{fsys}, name, path.Dir(name))
}

// loadFrom is the single implementation behind Load and LoadFS: files gives
// access to the .def at defPath and to everything it references, resolved
// against dir.
func loadFrom(files fileAccess, defPath, dir string) (*Character, error) {
	path := defPath
	f, err := files.open(path)
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

	src := source{files: files, dir: dir}

	animations, err := loadAnimations(src, info.AnimationFile)
	if err != nil {
		return nil, err
	}

	sprites, err := loadSprites(src, info.SpriteFile)
	if err != nil {
		return nil, err
	}

	// Ikemen GO reads constants only when cns is set, and a character may
	// carry its states in .zss alone (.vibe/decisions/032).
	var stateDefs []cns.StateDef
	if info.ConstantsFile != "" {
		stateDefs, err = loadStateDefs(src, info.ConstantsFile)
		if err != nil {
			return nil, err
		}
	}

	scripts := resolveScripts(info, func(entry string) ([]byte, error) {
		f, err := src.open(entry)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return io.ReadAll(f)
	})

	sounds, err := loadSounds(src, info.SoundFile)
	if err != nil {
		return nil, err
	}

	commands, commandErr := loadCommands(src, info.CommandFile)

	movelists := resolveMovelists(info.Movelists, func(file string) ([]byte, error) {
		f, err := src.open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return io.ReadAll(f)
	})

	return &Character{
		Name:            info.Name,
		Author:          info.Author,
		SpriteFile:      info.SpriteFile,
		AnimationFile:   info.AnimationFile,
		SoundFile:       info.SoundFile,
		CommandFile:     info.CommandFile,
		ConstantsFile:   info.ConstantsFile,
		StateFiles:      info.StateFiles,
		StateCommonFile: info.StateCommonFile,
		Scripts:         scripts,
		Palettes:        info.Palettes,
		Animations:      animations,
		Sprites:         sprites,
		StateDefs:       stateDefs,
		Sounds:          sounds,
		Commands:        commands,
		CommandError:    commandErr,
		Movelists:       movelists,
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
// untouched: a missing cns is allowed (states may come from .zss), while a
// missing anim or sprite entry still gets the per-file "opening ... file"
// error further down the load path.
// See backlog item 051.
func validateFilesSection(path string, info def.CharacterInfo) error {
	if info.AnimationFile != "" || info.SpriteFile != "" || info.ConstantsFile != "" {
		return nil
	}
	return fmt.Errorf("character: %q doesn't look like a character definition: no [Files] section (or it has no recognized entries)", path)
}

// fileAccess is the file system a character is read from. join builds the
// name of a child (and cleans it lexically); open, stat and readDir take
// names built by join. Two implementations exist: the OS, which keeps the
// caller's own path spelling so Load behaves and words its errors exactly
// as it always did, and an fs.FS.
type fileAccess interface {
	join(dir, child string) string
	open(name string) (fs.File, error)
	stat(name string) (fs.FileInfo, error)
	readDir(name string) ([]fs.DirEntry, error)
}

// osFiles is fileAccess over the operating system, with host path semantics.
type osFiles struct{}

func (osFiles) join(dir, child string) string         { return filepath.Join(dir, child) }
func (osFiles) open(name string) (fs.File, error)     { return os.Open(name) }
func (osFiles) stat(name string) (fs.FileInfo, error) { return os.Stat(name) }
func (osFiles) readDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

// fsFiles is fileAccess over an fs.FS, with slash-separated path semantics.
// A name that leaves the root of the fs.FS (a leading "..") is not valid
// for it and is reported as not found.
type fsFiles struct{ fsys fs.FS }

func (fsFiles) join(dir, child string) string { return path.Join(dir, child) }

func (f fsFiles) open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return f.fsys.Open(name)
}

func (f fsFiles) stat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
	}
	return fs.Stat(f.fsys, name)
}

func (f fsFiles) readDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	return fs.ReadDir(f.fsys, name)
}

// source is where the files referenced from a .def are looked up: through
// files, relative to dir (the .def's directory).
type source struct {
	files fileAccess
	dir   string
}

// shown resolves a reference to the path error messages name, normalizing
// Windows-style backslash separators to forward slashes first. Real .def
// files are authored on Windows, where a backslash is unambiguously a
// separator, but on a forward-slash filesystem (Linux, the browser/WASM
// target) it would be a literal filename character and the referenced
// subdirectory never traversed. dir itself is left untouched: it is already
// in the host's convention. See backlog item 049.
func (s source) shown(referenced string) string {
	return s.files.join(s.dir, strings.ReplaceAll(referenced, "\\", "/"))
}

// open opens the file a .def references and, if no file exists under that
// exact casing, falls back to a case-insensitive match against the actual
// directory entries, one path component at a time, before giving up. Real
// .def files are authored on Windows, where paths are case-insensitive, and
// many reference a file under different letter casing than its actual
// name. The exact-case name is always tried first, so a correctly-cased
// reference never pays for a directory listing. When no case-insensitive
// match exists either, the original error against the exact-case name is
// returned unchanged. See backlog items 050 and 084.
func (s source) open(referenced string) (fs.File, error) {
	name := s.shown(referenced)
	f, err := s.files.open(name)
	if err == nil {
		return f, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if resolved, ok := s.resolveCaseInsensitive(strings.ReplaceAll(referenced, "\\", "/")); ok {
		if f, ferr := s.files.open(resolved); ferr == nil {
			return f, nil
		}
	}
	return nil, err
}

// resolveCaseInsensitive walks relPath component by component starting at
// the .def's directory, matching each component against the actual
// directory entries at that level case-insensitively when no exact match
// exists. It returns the fully resolved name and true once every component
// (including the final file) has been matched, or ("", false) as soon as
// any component has no match under either casing.
func (s source) resolveCaseInsensitive(relPath string) (string, bool) {
	current := s.dir
	for _, component := range strings.Split(relPath, "/") {
		if component == "" || component == "." {
			continue
		}
		candidate := s.files.join(current, component)
		if _, err := s.files.stat(candidate); err == nil {
			current = candidate
			continue
		}

		entries, err := s.files.readDir(current)
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
		current = s.files.join(current, match)
	}
	return current, true
}

// loadAnimations opens and parses the .air file the reference points to.
func loadAnimations(src source, referenced string) ([]air.Animation, error) {
	path := src.shown(referenced)
	f, err := src.open(referenced)
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

// loadSprites opens and loads the .sff file the reference points to.
func loadSprites(src source, referenced string) ([]sff.SpriteGroup, error) {
	path := src.shown(referenced)
	f, err := src.open(referenced)
	if err != nil {
		return nil, fmt.Errorf("character: opening sprite file %q: %w", path, err)
	}
	defer f.Close()

	// sff.Load needs random access, which an fs.File does not promise.
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("character: loading sprite file %q: %w", path, err)
	}
	sprites, err := sff.Load(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("character: loading sprite file %q: %w", path, err)
	}
	return sprites, nil
}

// loadStateDefs opens and parses the .cns file the reference points to.
func loadStateDefs(src source, referenced string) ([]cns.StateDef, error) {
	path := src.shown(referenced)
	f, err := src.open(referenced)
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

// loadSounds opens and decodes the .snd file the reference points to.
//
// Unlike loadAnimations/loadSprites/loadStateDefs, referenced == "" is not
// an error: SoundFile is genuinely optional (validateFilesSection has never
// required it, unlike AnimationFile/SpriteFile/ConstantsFile) — many real
// characters have no dedicated sound file, or reference a shared one that
// isn't always present. Once referenced is non-empty, though, a missing or
// undecodable .snd file is a hard error, exactly like the other three
// referenced files. See .vibe/decisions/029.
//
// A resolved v2 external-file-reference entry (Ikemen GO's extension) is
// looked up relative to the .def directory, the same base every other referenced file
// already uses.
func loadSounds(src source, referenced string) ([]SoundGroup, error) {
	if referenced == "" {
		return nil, nil
	}

	path := src.shown(referenced)
	f, err := src.open(referenced)
	if err != nil {
		return nil, fmt.Errorf("character: opening sound file %q: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("character: reading sound file %q: %w", path, err)
	}

	openExternal := func(ref string) ([]byte, error) {
		return readExact(src, ref)
	}
	return decodeSoundGroups(data, path, openExternal)
}

// loadCommands opens and parses the .cmd file the reference points to.
//
// Unlike the other referenced files, a failure is non-fatal: it returns the
// normalized empty CommandFile plus a descriptive message naming the file and
// the failing step, which Load stores in Character.CommandError.
// referenced == "" (no cmd declared) yields the empty value and no message.
// See .vibe/decisions/031.
func loadCommands(src source, referenced string) (cmd.CommandFile, string) {
	empty := normalizeCommandFileForJSON(cmd.CommandFile{})
	if referenced == "" {
		return empty, ""
	}

	path := src.shown(referenced)
	f, err := src.open(referenced)
	if err != nil {
		return empty, fmt.Sprintf("character: opening command file %q: %v", path, err)
	}
	defer f.Close()

	file, err := cmd.Parse(f)
	if err != nil {
		return empty, fmt.Sprintf("character: parsing command file %q: %v", path, err)
	}
	return normalizeCommandFileForJSON(file), ""
}

// readExact reads a file by its exact name (no case-insensitive fallback,
// as for every v2 .snd external reference before LoadFS existed).
func readExact(src source, referenced string) ([]byte, error) {
	f, err := src.files.open(src.shown(referenced))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
