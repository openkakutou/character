package character

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openkakutou/character/def"
)

const zssFixtureText = "[Statedef 200; type: S;]\nif AnimElem = 1 {\n\tmapAdd{map:\"a\";value:1}\n}\n"

// fakeReader serves files from a map; absent keys report fs.ErrNotExist and
// keys listed in failing report a non-NotExist error. It records every
// entry it was asked for.
type fakeReader struct {
	files   map[string]string
	failing map[string]bool
	asked   []string
}

func (f *fakeReader) read(entry string) ([]byte, error) {
	f.asked = append(f.asked, entry)
	if f.failing[entry] {
		return nil, errors.New("disk exploded")
	}
	s, ok := f.files[entry]
	if !ok {
		return nil, fmt.Errorf("open %s: %w", entry, fs.ErrNotExist)
	}
	return []byte(s), nil
}

func scriptFiles(scripts []ScriptFile) string {
	names := make([]string, len(scripts))
	for i, s := range scripts {
		names[i] = s.File
	}
	return strings.Join(names, ",")
}

func TestResolveScripts_ZssStateFile_LoadsBlocks(t *testing.T) {
	r := &fakeReader{files: map[string]string{"a.zss": zssFixtureText}}
	got := resolveScripts(def.CharacterInfo{StateFiles: []string{"a.zss"}}, r.read)
	if len(got) != 1 || got[0].File != "a.zss" || got[0].Error != "" {
		t.Fatalf("got %+v", got)
	}
	if len(got[0].Script.Blocks) != 1 || got[0].Script.Blocks[0].Number != 200 {
		t.Errorf("blocks = %+v, want one Statedef 200", got[0].Script.Blocks)
	}
}

func TestResolveScripts_MixedCnsAndZss_OnlyZssInScripts(t *testing.T) {
	r := &fakeReader{files: map[string]string{"a.st": "x", "b.zss": zssFixtureText, "c.cns": "y"}}
	got := resolveScripts(def.CharacterInfo{StateFiles: []string{"a.st", "b.zss", "c.cns"}}, r.read)
	if scriptFiles(got) != "b.zss" {
		t.Errorf("files = %q, want only b.zss", scriptFiles(got))
	}
	for _, asked := range r.asked {
		if asked == "b.zss.zss" {
			t.Errorf("a found .zss must not trigger a fallback read")
		}
	}
}

func TestResolveScripts_UppercaseExtensionAndBackslash_Recognized(t *testing.T) {
	r := &fakeReader{files: map[string]string{`sub\A.ZSS`: zssFixtureText}}
	got := resolveScripts(def.CharacterInfo{StateFiles: []string{`sub\A.ZSS`}}, r.read)
	if len(got) != 1 || got[0].Error != "" || len(got[0].Script.Blocks) != 1 {
		t.Errorf("got %+v", got)
	}
}

func TestResolveScripts_ZssxExtension_NotTreatedAsZss(t *testing.T) {
	r := &fakeReader{files: map[string]string{"a.zssx": "whatever"}}
	got := resolveScripts(def.CharacterInfo{StateFiles: []string{"a.zssx"}}, r.read)
	if len(got) != 0 {
		t.Errorf("got %+v, want none", got)
	}
}

func TestResolveScripts_StateFilesThenStcommon_KeepsInfoOrderStcommonLast(t *testing.T) {
	r := &fakeReader{files: map[string]string{"s.zss": zssFixtureText, "s1.zss": zssFixtureText, "s2.zss": zssFixtureText, "c.zss": zssFixtureText}}
	info := def.CharacterInfo{StateFiles: []string{"s.zss", "s1.zss", "s2.zss"}, StateCommonFile: "c.zss"}
	got := resolveScripts(info, r.read)
	if scriptFiles(got) != "s.zss,s1.zss,s2.zss,c.zss" {
		t.Errorf("order = %q", scriptFiles(got))
	}
}

func TestResolveScripts_MissingCnsWithZssSibling_LoadsFallbackNamedAsLoaded(t *testing.T) {
	r := &fakeReader{files: map[string]string{"common1.cns.zss": zssFixtureText}}
	got := resolveScripts(def.CharacterInfo{StateCommonFile: "common1.cns"}, r.read)
	if len(got) != 1 || got[0].File != "common1.cns.zss" || got[0].Error != "" || len(got[0].Script.Blocks) != 1 {
		t.Errorf("got %+v", got)
	}
}

func TestResolveScripts_ExistingNonZssEntry_SkippedWithoutFallback(t *testing.T) {
	r := &fakeReader{files: map[string]string{"a.st": "x", "a.st.zss": zssFixtureText}}
	got := resolveScripts(def.CharacterInfo{StateFiles: []string{"a.st"}}, r.read)
	if len(got) != 0 {
		t.Errorf("got %+v, want none: an existing non-zss file is not replaced by its .zss sibling", got)
	}
}

func TestResolveScripts_NonZssStcommonMissingNoFallback_SilentlySkipped(t *testing.T) {
	r := &fakeReader{}
	got := resolveScripts(def.CharacterInfo{StateCommonFile: "common1.cns"}, r.read)
	if len(got) != 0 {
		t.Errorf("got %+v, want none", got)
	}
}

func TestResolveScripts_NonNotExistErrorOnNonZssEntry_DoesNotTriggerFallback(t *testing.T) {
	r := &fakeReader{failing: map[string]bool{"a.cns": true}, files: map[string]string{"a.cns.zss": zssFixtureText}}
	got := resolveScripts(def.CharacterInfo{StateFiles: []string{"a.cns"}}, r.read)
	if len(got) != 0 {
		t.Errorf("got %+v, want none", got)
	}
}

func TestResolveScripts_MissingZss_ErrorNamesFileOthersStillLoaded(t *testing.T) {
	r := &fakeReader{files: map[string]string{"ok.zss": zssFixtureText}}
	got := resolveScripts(def.CharacterInfo{StateFiles: []string{"gone.zss", "ok.zss"}}, r.read)
	if len(got) != 2 {
		t.Fatalf("got %+v", got)
	}
	if !strings.Contains(got[0].Error, "opening") || !strings.Contains(got[0].Error, "gone.zss") {
		t.Errorf("Error = %q, want it to name the opening step and gone.zss", got[0].Error)
	}
	if got[0].File != "gone.zss" || len(got[0].Script.Blocks) != 0 || got[0].Script.Blocks == nil {
		t.Errorf("failed entry = %+v, want File set and a normalized empty script", got[0])
	}
	if got[1].Error != "" || len(got[1].Script.Blocks) != 1 {
		t.Errorf("second entry = %+v, want loaded", got[1])
	}
}

func TestResolveScripts_ZssReadFailsOtherwise_ErrorNotSilentlySkipped(t *testing.T) {
	r := &fakeReader{failing: map[string]bool{"a.zss": true}}
	got := resolveScripts(def.CharacterInfo{StateFiles: []string{"a.zss"}}, r.read)
	if len(got) != 1 || !strings.Contains(got[0].Error, "disk exploded") {
		t.Errorf("got %+v", got)
	}
}

func TestResolveScripts_MalformedZss_ErrorNamesParsingStep(t *testing.T) {
	r := &fakeReader{files: map[string]string{"bad.zss": "[Statedef\nfoo\n"}}
	got := resolveScripts(def.CharacterInfo{StateFiles: []string{"bad.zss"}}, r.read)
	if len(got) != 1 || !strings.Contains(got[0].Error, "parsing") || !strings.Contains(got[0].Error, "bad.zss") {
		t.Errorf("got %+v", got)
	}
}

func TestResolveScripts_NoEntries_ReturnsNonNilEmpty(t *testing.T) {
	got := resolveScripts(def.CharacterInfo{}, (&fakeReader{}).read)
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v, want non-nil empty", got)
	}
}

// writeZssCharacter builds on writeFixtureCharacter: it replaces char.def
// with a [Files] section holding filesSection (no cns unless the caller
// writes one), removes char.cns, and writes extra files into the folder.
func writeZssCharacter(t *testing.T, filesSection string, extra map[string]string) string {
	t.Helper()
	defPath := writeFixtureCharacter(t)
	dir := filepath.Dir(defPath)
	if err := os.Remove(filepath.Join(dir, "char.cns")); err != nil {
		t.Fatalf("test setup: %v", err)
	}
	defText := "[Info]\nname = Z\n\n[Files]\nsprite = char.sff\nanim = char.air\n" + filesSection
	if err := os.WriteFile(defPath, []byte(defText), 0o644); err != nil {
		t.Fatalf("test setup: %v", err)
	}
	for name, content := range extra {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("test setup: %v", err)
		}
	}
	return defPath
}

func TestLoad_ZssOnlyCharacterWithoutCns_LoadsScriptsAndEmptyStateDefs(t *testing.T) {
	path := writeZssCharacter(t, "st = Main.ZSS\n", map[string]string{"main.zss": zssFixtureText})
	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(c.StateDefs) != 0 {
		t.Errorf("StateDefs = %d, want 0", len(c.StateDefs))
	}
	if len(c.Scripts) != 1 || c.Scripts[0].Error != "" || len(c.Scripts[0].Script.Blocks) != 1 {
		t.Fatalf("Scripts = %+v", c.Scripts)
	}
	if c.Scripts[0].File != "Main.ZSS" {
		t.Errorf("File = %q, want the entry as written", c.Scripts[0].File)
	}
}

func TestLoad_ZssStcommonAbsentFromFolder_ReportedNonFatally(t *testing.T) {
	path := writeZssCharacter(t, "st = main.zss\nstcommon = common.zss\n", map[string]string{"main.zss": zssFixtureText})
	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load must succeed, got %v", err)
	}
	if len(c.Scripts) != 2 || c.Scripts[0].Error != "" {
		t.Fatalf("Scripts = %+v", c.Scripts)
	}
	if c.StateCommonFile != "common.zss" {
		t.Errorf("StateCommonFile = %q", c.StateCommonFile)
	}
	if !strings.Contains(c.Scripts[1].Error, "common.zss") || !strings.Contains(c.Scripts[1].Error, "opening") {
		t.Errorf("Error = %q, want opening step naming common.zss", c.Scripts[1].Error)
	}
}

func TestLoad_CnsStcommonNotInFolder_SilentlySkipped(t *testing.T) {
	path := writeZssCharacter(t, "stcommon = common1.cns\n", nil)
	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(c.Scripts) != 0 {
		t.Errorf("Scripts = %+v, want none", c.Scripts)
	}
}

func TestLoad_CnsStcommonWithZssSibling_LoadsFallback(t *testing.T) {
	path := writeZssCharacter(t, "stcommon = common1.cns\n", map[string]string{"common1.cns.zss": zssFixtureText})
	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(c.Scripts) != 1 || c.Scripts[0].File != "common1.cns.zss" || c.Scripts[0].Error != "" {
		t.Errorf("Scripts = %+v", c.Scripts)
	}
}

func TestLoad_UnmodifiedZss_SerializeZssIsByteExact(t *testing.T) {
	content := "; comment kept\n" + zssFixtureText + "\n\n[Function F()]\nlet x = 1;\n"
	path := writeZssCharacter(t, "st = main.zss\n", map[string]string{"main.zss": content})
	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	out, err := SerializeZss([]byte(content), c.Scripts[0].Script)
	if err != nil {
		t.Fatalf("SerializeZss: %v", err)
	}
	if string(out) != content {
		t.Errorf("not byte-exact:\n%q\nvs\n%q", out, content)
	}
}

func TestLoadBytes_ZssSuppliedUnderDifferentCaseAndSeparator_Loaded(t *testing.T) {
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	defBytes := []byte("[Files]\nst = Sub\\Main.ZSS\n")
	c, err := LoadBytes(defBytes, airBytes, sffBytes, cnsBytes, nil, nil, map[string][]byte{"sub/main.zss": []byte(zssFixtureText)})
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if len(c.Scripts) != 1 || c.Scripts[0].Error != "" || len(c.Scripts[0].Script.Blocks) != 1 {
		t.Errorf("Scripts = %+v", c.Scripts)
	}
}

func TestLoadBytes_DeclaredZssNotSupplied_ErrorButLoadSucceeds(t *testing.T) {
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	c, err := LoadBytes([]byte("[Files]\nst = main.zss\n"), airBytes, sffBytes, cnsBytes, nil, nil, nil)
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if len(c.Scripts) != 1 || !strings.Contains(c.Scripts[0].Error, "not supplied") || strings.Count(c.Scripts[0].Error, "main.zss") != 1 {
		t.Errorf("Scripts = %+v", c.Scripts)
	}
}

func TestLoadBytes_ZssFallbackSupplied_Loaded(t *testing.T) {
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	c, err := LoadBytes([]byte("[Files]\nstcommon = common1.cns\n"), airBytes, sffBytes, cnsBytes, nil, nil,
		map[string][]byte{"common1.cns.zss": []byte(zssFixtureText)})
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if len(c.Scripts) != 1 || c.Scripts[0].File != "common1.cns.zss" {
		t.Errorf("Scripts = %+v", c.Scripts)
	}
}

func TestLoadBytes_NoZss_ScriptsJSONIsEmptyArray(t *testing.T) {
	defBytes, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	c, err := LoadBytes(defBytes, airBytes, sffBytes, cnsBytes, nil, nil, nil)
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"scripts":[]`) {
		t.Errorf("JSON lacks \"scripts\":[]")
	}
}

func TestLoadBytes_EmptyCnsBytes_YieldsEmptyStateDefs(t *testing.T) {
	defBytes, airBytes, sffBytes, _ := fixtureCharacterBytes(t)
	c, err := LoadBytes(defBytes, airBytes, sffBytes, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if len(c.StateDefs) != 0 {
		t.Errorf("StateDefs = %d", len(c.StateDefs))
	}
}

func TestResolveScripts_EmptyStEntry_Skipped(t *testing.T) {
	r := &fakeReader{files: map[string]string{"a.zss": zssFixtureText}}
	got := resolveScripts(def.CharacterInfo{StateFiles: []string{"", "a.zss"}}, r.read)
	if scriptFiles(got) != "a.zss" {
		t.Errorf("files = %q, want only a.zss", scriptFiles(got))
	}
	if len(r.asked) != 1 || r.asked[0] != "a.zss" {
		t.Errorf("asked = %v, want only a.zss", r.asked)
	}
}

func TestResolveScripts_FallbackExistsButReadFails_ReportedAsError(t *testing.T) {
	r := &fakeReader{failing: map[string]bool{"x.cns.zss": true}}
	got := resolveScripts(def.CharacterInfo{StateCommonFile: "x.cns"}, r.read)
	if len(got) != 1 || got[0].File != "x.cns.zss" || !strings.Contains(got[0].Error, "disk exploded") {
		t.Errorf("got %+v, want one x.cns.zss entry carrying the read error", got)
	}
}

func TestResolveScripts_FallbackReadsEachFileOnce(t *testing.T) {
	r := &fakeReader{files: map[string]string{"c.cns.zss": zssFixtureText}}
	resolveScripts(def.CharacterInfo{StateCommonFile: "c.cns"}, r.read)
	if strings.Join(r.asked, ",") != "c.cns,c.cns.zss" {
		t.Errorf("asked = %v, want each of c.cns and c.cns.zss read once", r.asked)
	}
}

func TestLoadBytes_ErroredScriptEntry_JSONKeepsEmptyBlocks(t *testing.T) {
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	c, err := LoadBytes([]byte("[Files]\nst = main.zss\n"), airBytes, sffBytes, cnsBytes, nil, nil, nil)
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	data, _ := json.Marshal(c.Scripts)
	if !strings.Contains(string(data), `"blocks":[]`) {
		t.Errorf("JSON = %s, want \"blocks\":[]", data)
	}
}

func TestLoadBytes_ScriptBlockJSON_KeepsZeroNumberAndEmptyCollections(t *testing.T) {
	_, airBytes, sffBytes, cnsBytes := fixtureCharacterBytes(t)
	text := "[Statedef 0; type: S;]\nx\n[Function f()]\ny\n"
	c, err := LoadBytes([]byte("[Files]\nst = a.zss\n"), airBytes, sffBytes, cnsBytes, nil, nil, map[string][]byte{"a.zss": []byte(text)})
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	data, _ := json.Marshal(c.Scripts[0].Script.Blocks)
	for _, want := range []string{`"number":0`, `"params":[]`, `"ret":[]`, `"headerParams":{}`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("JSON = %s, want it to contain %s", data, want)
		}
	}
}
