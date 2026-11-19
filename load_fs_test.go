package character

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

// mapFSFromDir copies every regular file under dir into a MapFS keyed by its
// slash-separated path relative to dir.
func mapFSFromDir(t *testing.T, dir string) fstest.MapFS {
	t.Helper()
	m := fstest.MapFS{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		m[filepath.ToSlash(rel)] = &fstest.MapFile{Data: data}
		return nil
	})
	if err != nil {
		t.Fatalf("test setup: copying fixture to MapFS: %v", err)
	}
	return m
}

func TestLoadFS_MapFSOfRealCharacter_EqualsLoadOnDisk(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	want, err := Load(defPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(want.Animations) != 1 || len(want.StateDefs) != 1 || len(want.Sprites) == 0 {
		t.Fatalf("fixture too thin to compare: %d anims, %d states, %d sprite groups", len(want.Animations), len(want.StateDefs), len(want.Sprites))
	}

	got, err := LoadFS(mapFSFromDir(t, filepath.Dir(defPath)), filepath.Base(defPath))
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LoadFS result differs from Load:\n got: %+v\nwant: %+v", got, want)
	}
}

func TestLoadFS_DefInSubdirectory_ResolvesReferencesFromDefDirectory(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	files := mapFSFromDir(t, filepath.Dir(defPath))
	nested := fstest.MapFS{}
	for name, f := range files {
		nested["chars/kfm/"+name] = f
	}

	c, err := LoadFS(nested, "chars/kfm/char.def")
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	if c.Name != "Test Character" || len(c.Animations) != 1 {
		t.Errorf("unexpected character: name %q, %d animations", c.Name, len(c.Animations))
	}
}

func TestLoadFS_CaseMismatchAndBackslashReferences_Resolve(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	files := mapFSFromDir(t, filepath.Dir(defPath))
	def := strings.Replace(string(files["char.def"].Data), "anim = char.air", `anim = Data\SUB\Char.AIR`, 1)
	files["char.def"] = &fstest.MapFile{Data: []byte(def)}
	files["data/sub/char.air"] = files["char.air"]
	delete(files, "char.air")

	c, err := LoadFS(files, "char.def")
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	if len(c.Animations) != 1 || c.Animations[0].Number != 200 {
		t.Errorf("expected action 200 loaded through backslash+case fallback, got %+v", c.Animations)
	}
}

func TestLoadFS_ReferenceInsideRootWithDotDot_Resolves(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	files := mapFSFromDir(t, filepath.Dir(defPath))
	def := strings.Replace(string(files["char.def"].Data), "anim = char.air", "anim = ../shared.air", 1)
	nested := fstest.MapFS{
		"a/b/char.def": &fstest.MapFile{Data: []byte(def)},
		"a/b/char.sff": files["char.sff"],
		"a/b/char.cns": files["char.cns"],
		"a/shared.air": files["char.air"],
	}

	c, err := LoadFS(nested, "a/b/char.def")
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	if len(c.Animations) != 1 {
		t.Errorf("expected the ../shared.air animation, got %d", len(c.Animations))
	}
}

func TestLoadFS_ReferenceEscapingRoot_IsMissingFileNotPanic(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	files := mapFSFromDir(t, filepath.Dir(defPath))
	def := strings.Replace(string(files["char.def"].Data), "anim = char.air", "anim = ../x.air", 1)
	files["char.def"] = &fstest.MapFile{Data: []byte(def)}
	files["x.air"] = files["char.air"] // exists, but outside the root of a sub-FS
	sub, err := fs.Sub(files, ".")
	if err != nil {
		t.Fatal(err)
	}

	for _, fsys := range []fs.FS{files, sub} {
		_, err = LoadFS(fsys, "char.def")
		if err == nil {
			t.Fatal("expected an error for a reference leaving the root, got nil")
		}
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("expected a not-exist error, got: %v", err)
		}
		if !strings.Contains(err.Error(), "opening animation file") {
			t.Errorf("expected the opening-animation-file message, got: %v", err)
		}
	}
}

func TestLoadFS_MissingDef_ReturnsError(t *testing.T) {
	_, err := LoadFS(fstest.MapFS{}, "nope/char.def")
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected a not-exist error for a missing .def, got: %v", err)
	}
}

func TestLoad_ReferenceWithDotDot_StillLoads(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	dir := filepath.Dir(defPath)
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"char.def", "char.sff", "char.cns"} {
		data, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			t.Fatal(err)
		}
		if n == "char.def" {
			data = []byte(strings.Replace(string(data), "anim = char.air", `anim = ..\char.air`, 1))
		}
		if err := os.WriteFile(filepath.Join(sub, n), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	for name, path := range map[string]string{"absolute": filepath.Join(sub, "char.def"), "relative": "char.def"} {
		t.Run(name, func(t *testing.T) {
			if name == "relative" {
				t.Chdir(sub)
			}
			c, err := Load(path)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if len(c.Animations) != 1 {
				t.Errorf("expected the ../char.air animation, got %d", len(c.Animations))
			}
		})
	}
}

func TestLoad_RelativePathFromSymlinkedCwd_ResolvesDotDotAsTheOSDoes(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	files := mapFSFromDir(t, filepath.Dir(defPath))
	base := t.TempDir()
	realCwd := filepath.Join(base, "real", "deep", "cwd")
	links := filepath.Join(base, "links")
	for _, d := range []string{realCwd, links} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(links, "link")
	if err := os.Symlink(realCwd, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	// shared.air sits next to the real cwd's parent only: climbing through the
	// symlink's own parent ("links") would not find it.
	def := strings.Replace(string(files["char.def"].Data), "anim = char.air", "anim = ../shared.air", 1)
	writes := map[string][]byte{
		filepath.Join(realCwd, "char.def"):                []byte(def),
		filepath.Join(realCwd, "char.sff"):                files["char.sff"].Data,
		filepath.Join(realCwd, "char.cns"):                files["char.cns"].Data,
		filepath.Join(base, "real", "deep", "shared.air"): files["char.air"].Data,
	}
	for p, data := range writes {
		if err := os.WriteFile(p, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Chdir(link)
	c, err := Load("char.def")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(c.Animations) != 1 {
		t.Errorf("expected the ../shared.air animation, got %d", len(c.Animations))
	}
}

func TestLoad_MissingCommandFile_ErrorKeepsCallerSpelledPath(t *testing.T) {
	defPath := writeFixtureCharacter(t)
	dir := filepath.Dir(defPath)
	data, err := os.ReadFile(defPath)
	if err != nil {
		t.Fatal(err)
	}
	def := strings.Replace(string(data), "cns = char.cns", "cns = char.cns\ncmd = missing.cmd", 1)
	if err := os.WriteFile(defPath, []byte(def), 0o644); err != nil {
		t.Fatal(err)
	}

	c, err := Load(defPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	p := dir + "/missing.cmd"
	want := `character: opening command file "` + p + `": open ` + p + `: no such file or directory`
	if c.CommandError != want {
		t.Errorf("CommandError = %q, want %q", c.CommandError, want)
	}
}
