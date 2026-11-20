package character

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// movelistFixtureText is testdata/movelist/movelist.txt without its byte
// order mark: Ikemen GO strips a UTF-8 BOM and keeps everything else as is
// (src/common.go:503 LoadText), CRLF and glyph tokens included.
const movelistFixtureText = "Special Moves\r\nFireball\t_QCF^A\r\n\t^A^B\r\n"

// writeMovelistCharacter writes the fixture character plus extraDef appended
// to its [Files] section, and the testdata move list as files/<name> for
// each name in present. It returns the .def path.
func writeMovelistCharacter(t *testing.T, extraDef string, present ...string) string {
	t.Helper()
	defPath := writeFixtureCharacter(t)
	dir := filepath.Dir(defPath)
	f, err := os.OpenFile(defPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}
	if _, err := f.WriteString(extraDef); err != nil {
		t.Fatalf("test setup: %v", err)
	}
	f.Close()

	data, err := os.ReadFile(filepath.Join("testdata", "movelist", "movelist.txt"))
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}
	for _, name := range present {
		target := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("test setup: %v", err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			t.Fatalf("test setup: %v", err)
		}
	}
	return defPath
}

func TestLoad_Movelists_ExposeTextPerDeclaredKey(t *testing.T) {
	defPath := writeMovelistCharacter(t, "movelist = help\\en.txt\nmovelist1 = help/fr.txt\n", "help/en.txt", "help/fr.txt")

	c, err := Load(defPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	want := []Movelist{
		{Key: "movelist", Index: 0, File: "help\\en.txt", Text: movelistFixtureText},
		{Key: "movelist1", Index: 1, File: "help/fr.txt", Text: movelistFixtureText},
	}
	if !reflect.DeepEqual(c.Movelists, want) {
		t.Fatalf("Movelists:\n got %#v\nwant %#v", c.Movelists, want)
	}
}

func TestLoad_MissingMovelistFile_ReportedOnCharacterWithoutFailingLoad(t *testing.T) {
	defPath := writeMovelistCharacter(t, "movelist = gone.txt\nmovelist1 = here.txt\n", "here.txt")

	c, err := Load(defPath)
	if err != nil {
		t.Fatalf("Load must not fail on a missing move list, got: %v", err)
	}
	if len(c.Movelists) != 2 {
		t.Fatalf("expected 2 Movelists, got %#v", c.Movelists)
	}
	missing := c.Movelists[0]
	if missing.Text != "" || !strings.Contains(missing.Error, "gone.txt") {
		t.Errorf("missing entry: want empty Text and an Error naming gone.txt, got %#v", missing)
	}
	if c.Movelists[1].Error != "" || c.Movelists[1].Text != movelistFixtureText {
		t.Errorf("present entry must be unaffected, got %#v", c.Movelists[1])
	}
}

func TestLoad_UnreadableMovelistFile_ReportedOnCharacter(t *testing.T) {
	defPath := writeMovelistCharacter(t, "movelist = dir.txt\n")
	if err := os.Mkdir(filepath.Join(filepath.Dir(defPath), "dir.txt"), 0o755); err != nil {
		t.Fatalf("test setup: %v", err)
	}

	c, err := Load(defPath)
	if err != nil {
		t.Fatalf("Load must not fail on an unreadable move list, got: %v", err)
	}
	if len(c.Movelists) != 1 || c.Movelists[0].Error == "" || c.Movelists[0].Text != "" {
		t.Fatalf("expected one entry with an Error and no Text, got %#v", c.Movelists)
	}
}

func TestLoad_CharacterWithoutMovelistKey_HasNoMovelists(t *testing.T) {
	c, err := Load(writeFixtureCharacter(t))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if c.Movelists == nil || len(c.Movelists) != 0 {
		t.Fatalf("expected an empty, non-nil Movelists, got %#v", c.Movelists)
	}
}

func TestLoadBytes_Movelists_TextFromSuppliedFiles_MissingReported(t *testing.T) {
	defBytes := []byte("[Files]\nmovelist = Help\\En.txt\nmovelist1 = fr.txt\n")
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)

	c, err := LoadBytes(defBytes, airBytes, sffBytes, cnsBytes, nil, nil, nil,
		map[string][]byte{"help/en.txt": []byte("\xef\xbb\xbf" + movelistFixtureText)})
	if err != nil {
		t.Fatalf("LoadBytes returned error: %v", err)
	}
	if len(c.Movelists) != 2 {
		t.Fatalf("expected 2 Movelists, got %#v", c.Movelists)
	}
	if c.Movelists[0].Text != movelistFixtureText || c.Movelists[0].Error != "" {
		t.Errorf("supplied entry: got %#v", c.Movelists[0])
	}
	if c.Movelists[1].Text != "" || !strings.Contains(c.Movelists[1].Error, "fr.txt") {
		t.Errorf("unsupplied entry: want an Error naming fr.txt, got %#v", c.Movelists[1])
	}
}

func TestLoadBytes_NoMovelistKey_JSONHasEmptyMovelistsArray(t *testing.T) {
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	c, err := LoadBytes([]byte("[Info]\nname = x\n"), airBytes, sffBytes, cnsBytes, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("LoadBytes returned error: %v", err)
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"movelists":[]`) {
		t.Fatalf("expected \"movelists\":[] in JSON, got %s", data)
	}
}

func TestLoadBytes_Movelists_JSONShape(t *testing.T) {
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	c, err := LoadBytes([]byte("[Files]\nmovelist = a.txt\nmovelist2 = b.txt\n"), airBytes, sffBytes, cnsBytes, nil, nil, nil,
		map[string][]byte{"a.txt": []byte("hi")})
	if err != nil {
		t.Fatalf("LoadBytes returned error: %v", err)
	}
	data, err := json.Marshal(c.Movelists)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got []map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got[0]["key"] != "movelist" || got[0]["index"] != float64(0) || got[0]["file"] != "a.txt" || got[0]["text"] != "hi" {
		t.Errorf("first entry shape: %v", got[0])
	}
	if _, has := got[0]["error"]; has {
		t.Errorf("a loaded entry must omit \"error\": %v", got[0])
	}
	if got[1]["key"] != "movelist2" || got[1]["text"] != "" || got[1]["error"] == "" {
		t.Errorf("second entry shape: %v", got[1])
	}
}
