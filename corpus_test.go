package character

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/openkakutou/character/air"
	"github.com/openkakutou/character/cmd"
	"github.com/openkakutou/character/cns"
	"github.com/openkakutou/character/def"
	"github.com/openkakutou/character/zss"
)

// TestCorpus_SaveRoundTripAndSingleEdit is opt-in: set CHARACTER_CORPUS_DIR
// to a directory of real characters. For every parseable file it checks
// that an unmodified model saves byte-exact, and that a single-value edit
// changes at most two lines and survives a re-parse.
func TestCorpus_SaveRoundTripAndSingleEdit(t *testing.T) {
	dir := os.Getenv("CHARACTER_CORPUS_DIR")
	if dir == "" {
		t.Skip("CHARACTER_CORPUS_DIR not set")
	}

	type stats struct{ ok, skipped, noEdit, failed int }
	kinds := map[string]int{}
	perExt := map[string]*stats{}
	var failures []string
	fail := func(path, msg string) {
		failures = append(failures, path+": "+msg)
	}

	// kind -> (parse+noop+edit) handler; returns skip, noEditPossible, error message.
	handlers := map[string]func(data []byte) (skip, noEdit bool, msg string){
		"def": corpusDef, "air": corpusAir, "cns": corpusCns,
		"st": corpusCns, "cmd": corpusCmd, "zss": corpusZss,
	}

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
		h, ok := handlers[ext]
		if !ok {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		s := perExt[ext]
		if s == nil {
			s = &stats{}
			perExt[ext] = s
		}
		skip, noEdit, msg := h(data)
		switch {
		case skip:
			s.skipped++
		case msg != "":
			s.failed++
			k := "error"
			for _, p := range []string{"no-op mismatch", "reparse differs", "too many lines"} {
				if strings.HasPrefix(msg, p) {
					k = p
				}
			}
			kinds[ext+" / "+k]++
			fail(path, msg)
		default:
			s.ok++
			if noEdit {
				s.noEdit++
			}
		}
		return nil
	})

	exts := make([]string, 0, len(perExt))
	for e := range perExt {
		exts = append(exts, e)
	}
	sort.Strings(exts)
	for _, e := range exts {
		s := perExt[e]
		t.Logf(".%s: ok=%d (no-op only, nothing editable: %d) parse-skipped=%d failed=%d", e, s.ok, s.noEdit, s.skipped, s.failed)
	}
	ks := make([]string, 0, len(kinds))
	for k := range kinds {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		t.Logf("failure kind %s: %d", k, kinds[k])
	}
	for i, f := range failures {
		if i == 40 {
			t.Logf("... and %d more", len(failures)-40)
			break
		}
		t.Log(f)
	}
	if len(failures) > 0 {
		t.Errorf("%d corpus failures", len(failures))
	}
}

// lineDiff counts lines present in one text but not the other (multiset
// difference, both directions).
func lineDiff(a, b []byte) int {
	counts := map[string]int{}
	for _, l := range strings.Split(string(a), "\n") {
		counts[l]++
	}
	for _, l := range strings.Split(string(b), "\n") {
		counts[l]--
	}
	n := 0
	for _, c := range counts {
		if c < 0 {
			c = -c
		}
		n += c
	}
	return n
}

// lineHint lists up to 6 lines that differ (- original only, + output only).
func lineHint(a, b []byte) string {
	counts := map[string]int{}
	for _, l := range strings.Split(string(a), "\n") {
		counts[l]++
	}
	var out []string
	for _, l := range strings.Split(string(b), "\n") {
		if counts[l] > 0 {
			counts[l]--
		} else {
			out = append(out, fmt.Sprintf("+%q", l))
		}
	}
	for l, c := range counts {
		if c > 0 {
			out = append(out, fmt.Sprintf("-%q", l))
		}
	}
	sort.Strings(out)
	if len(out) > 6 {
		out = out[:6]
	}
	return strings.Join(out, " ")
}

func corpusCheck(original, noop, edited []byte, editedOK bool) string {
	if !bytes.Equal(noop, original) {
		return "no-op mismatch: save is not byte-exact"
	}
	if !editedOK {
		return ""
	}
	if n := lineDiff(original, edited); n > 2 {
		return fmt.Sprintf("too many lines: single edit changed %d lines; %s", n, lineHint(original, edited))
	}
	return ""
}

func corpusDef(data []byte) (bool, bool, string) {
	doc, err := def.ParseDocument(bytes.NewReader(data))
	if err != nil {
		return true, false, ""
	}
	noop, err := SerializeDef(data, doc.Info)
	if err != nil {
		return false, false, "no-op: " + err.Error()
	}
	info := doc.Info
	switch {
	case info.Author != "":
		info.Author += "x"
	case info.Name != "":
		info.Name += "x"
	default:
		return false, true, corpusCheck(data, noop, nil, false)
	}
	out, err := SerializeDef(data, info)
	if err != nil {
		return false, false, "edit: " + err.Error()
	}
	if msg := corpusCheck(data, noop, out, true); msg != "" {
		return false, false, msg
	}
	again, err := def.ParseDocument(bytes.NewReader(out))
	if err != nil {
		return false, false, "re-parse: " + err.Error()
	}
	a, b := again.Info, info
	normalizeCharacterInfoForJSON(&a)
	normalizeCharacterInfoForJSON(&b)
	if !reflect.DeepEqual(a, b) {
		return false, false, reparseDiff(a, b)
	}
	return false, false, ""
}

func corpusAir(data []byte) (bool, bool, string) {
	doc, err := air.ParseDocument(bytes.NewReader(data))
	if err != nil {
		return true, false, ""
	}
	noop, err := SerializeAir(data, doc.Animations)
	if err != nil {
		return false, false, "no-op: " + err.Error()
	}
	if msg := corpusCheck(data, noop, nil, false); msg != "" {
		return false, false, msg
	}
	// Deep-copy through a fresh parse so the edit does not alias doc.
	fresh, _ := air.ParseDocument(bytes.NewReader(data))
	for i := range fresh.Animations {
		if len(fresh.Animations[i].Frames) == 0 {
			continue
		}
		fresh.Animations[i].Frames[0].Time++
		out, err := SerializeAir(data, fresh.Animations)
		if err != nil {
			return false, false, "edit: " + err.Error()
		}
		if msg := corpusCheck(data, noop, out, true); msg != "" {
			return false, false, msg
		}
		again, err := air.ParseDocument(bytes.NewReader(out))
		if err != nil {
			return false, false, "re-parse: " + err.Error()
		}
		if !reflect.DeepEqual(normalizeAnimationsForJSON(again.Animations), normalizeAnimationsForJSON(fresh.Animations)) {
			return false, false, reparseDiff(normalizeAnimationsForJSON(again.Animations), normalizeAnimationsForJSON(fresh.Animations))
		}
		return false, false, ""
	}
	return false, true, ""
}

// editStates applies one valid edit: Anim+1 on the first StateDef without an
// "anim" header expression, else "+0" on the first controller parameter
// (sorted first key). Reports false when nothing is editable.
func editStates(states []cns.StateDef) bool {
	for i := range states {
		if _, expr := states[i].HeaderExprs["anim"]; !expr {
			states[i].Anim++
			return true
		}
	}
	for i := range states {
		for j := range states[i].Controllers {
			p := states[i].Controllers[j].Parameters
			if len(p) == 0 {
				continue
			}
			keys := make([]string, 0, len(p))
			for k := range p {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			p[keys[0]] += "+0"
			return true
		}
	}
	return false
}

// firstDiff walks two values and names the first differing path.
func firstDiff(path string, a, b reflect.Value) string {
	if a.Kind() != b.Kind() {
		return path + ": kind"
	}
	switch a.Kind() {
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: len %d vs %d", path, a.Len(), b.Len())
		}
		for i := 0; i < a.Len(); i++ {
			if d := firstDiff(fmt.Sprintf("%s[%d]", path, i), a.Index(i), b.Index(i)); d != "" {
				return d
			}
		}
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if d := firstDiff(path+"."+a.Type().Field(i).Name, a.Field(i), b.Field(i)); d != "" {
				return d
			}
		}
	case reflect.Map:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: map len %d vs %d", path, a.Len(), b.Len())
		}
		for _, k := range a.MapKeys() {
			bv := b.MapIndex(k)
			if !bv.IsValid() || fmt.Sprint(bv.Interface()) != fmt.Sprint(a.MapIndex(k).Interface()) {
				return fmt.Sprintf("%s[%v]: %v vs %v", path, k, a.MapIndex(k), bv)
			}
		}
	default:
		if !reflect.DeepEqual(a.Interface(), b.Interface()) {
			return fmt.Sprintf("%s: %v vs %v", path, a.Interface(), b.Interface())
		}
	}
	return ""
}

func reparseDiff(again, want any) string {
	return "reparse differs: " + firstDiff("", reflect.ValueOf(again), reflect.ValueOf(want))
}

func corpusCns(data []byte) (bool, bool, string) {
	doc, err := cns.ParseDocument(bytes.NewReader(data))
	if err != nil {
		return true, false, ""
	}
	noop, err := SerializeCns(data, doc.StateDefs)
	if err != nil {
		return false, false, "no-op: " + err.Error()
	}
	if msg := corpusCheck(data, noop, nil, false); msg != "" {
		return false, false, msg
	}
	fresh, _ := cns.ParseDocument(bytes.NewReader(data))
	if !editStates(fresh.StateDefs) {
		return false, true, ""
	}
	out, err := SerializeCns(data, fresh.StateDefs)
	if err != nil {
		return false, false, "edit: " + err.Error()
	}
	if msg := corpusCheck(data, noop, out, true); msg != "" {
		return false, false, msg
	}
	again, err := cns.ParseDocument(bytes.NewReader(out))
	if err != nil {
		return false, false, "re-parse: " + err.Error()
	}
	if !reflect.DeepEqual(normalizeStateDefsForJSON(again.StateDefs), normalizeStateDefsForJSON(fresh.StateDefs)) {
		return false, false, reparseDiff(normalizeStateDefsForJSON(again.StateDefs), normalizeStateDefsForJSON(fresh.StateDefs))
	}
	return false, false, ""
}

func corpusCmd(data []byte) (bool, bool, string) {
	doc, err := cmd.ParseDocument(bytes.NewReader(data))
	if err != nil {
		return true, false, ""
	}
	noop, err := SerializeCmd(data, doc.File)
	if err != nil {
		return false, false, "no-op: " + err.Error()
	}
	if msg := corpusCheck(data, noop, nil, false); msg != "" {
		return false, false, msg
	}
	fresh, _ := cmd.ParseDocument(bytes.NewReader(data))
	if !editStates(fresh.File.States) {
		if len(fresh.File.Commands) == 0 {
			return false, true, ""
		}
		fresh.File.Commands[0].Time++
	}
	out, err := SerializeCmd(data, fresh.File)
	if err != nil {
		return false, false, "edit: " + err.Error()
	}
	if msg := corpusCheck(data, noop, out, true); msg != "" {
		return false, false, msg
	}
	again, err := cmd.ParseDocument(bytes.NewReader(out))
	if err != nil {
		return false, false, "re-parse: " + err.Error()
	}
	if !reflect.DeepEqual(normalizeCommandFileForJSON(again.File), normalizeCommandFileForJSON(fresh.File)) {
		return false, false, reparseDiff(normalizeCommandFileForJSON(again.File), normalizeCommandFileForJSON(fresh.File))
	}
	return false, false, ""
}

func corpusZss(data []byte) (bool, bool, string) {
	doc, err := zss.ParseDocument(bytes.NewReader(data))
	if err != nil {
		return true, false, ""
	}
	noop, err := SerializeZss(data, doc.Script)
	if err != nil {
		return false, false, "no-op: " + err.Error()
	}
	if msg := corpusCheck(data, noop, nil, false); msg != "" {
		return false, false, msg
	}
	fresh, _ := zss.ParseDocument(bytes.NewReader(data))
	if len(fresh.Script.Blocks) == 0 {
		return false, true, ""
	}
	b := &fresh.Script.Blocks[0]
	if b.Body != "" && !strings.HasSuffix(b.Body, "\n") {
		b.Body += "\n"
	}
	b.Body += "# corpus-edit\n"
	out, err := SerializeZss(data, fresh.Script)
	if err != nil {
		return false, false, "edit: " + err.Error()
	}
	if msg := corpusCheck(data, noop, out, true); msg != "" {
		return false, false, msg
	}
	again, err := zss.ParseDocument(bytes.NewReader(out))
	if err != nil {
		return false, false, "re-parse: " + err.Error()
	}
	if !reflect.DeepEqual(normalizeScriptForJSON(again.Script), normalizeScriptForJSON(fresh.Script)) {
		return false, false, reparseDiff(normalizeScriptForJSON(again.Script), normalizeScriptForJSON(fresh.Script))
	}
	return false, false, ""
}
