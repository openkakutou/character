package def

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func editedOutput(t *testing.T, src string, edit func(*CharacterInfo)) string {
	t.Helper()
	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	info := doc.Info
	info.StateFiles = append([]string(nil), info.StateFiles...)
	info.Palettes = append([]string(nil), info.Palettes...)
	edit(&info)
	var buf bytes.Buffer
	if err := doc.SerializeEdited(&buf, info); err != nil {
		t.Fatalf("SerializeEdited: %v", err)
	}
	return buf.String()
}

func assertEdited(t *testing.T, src, want string, edit func(*CharacterInfo)) {
	t.Helper()
	if got := editedOutput(t, src, edit); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestSerializeEdited_ChangedAuthor_ChangesOnlyThatLineKeepingComment(t *testing.T) {
	src := "; header\n[Info]\nName = \"Kfm\"\nauthor = \"Elecbyte\"  ; who\n\n[Files]\ncns = kfm.cns\n"
	want := "; header\n[Info]\nName = \"Kfm\"\nauthor = \"Me\"  ; who\n\n[Files]\ncns = kfm.cns\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.Author = "Me" })
}

func TestSerializeEdited_NoChange_ReproducesSourceByteExact(t *testing.T) {
	src := "[Info]\r\nName = \"A\" ;c\r\n\r\n[Weird]\r\nx=1\r\n[Files]\r\nst = a.st\r\nst1=b.st\r\npal1 = p.act"
	assertEdited(t, src, src, func(*CharacterInfo) {})
}

func TestSerializeEdited_CRLFFile_KeepsCRLFOnPatchedAndInsertedLines(t *testing.T) {
	src := "[Info]\r\nname = \"A\"\r\n[Files]\r\ncns = a.cns\r\n"
	want := "[Info]\r\nname = \"B\"\r\n[Files]\r\ncns = a.cns\r\nsprite = a.sff\r\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.Name = "B"; i.SpriteFile = "a.sff" })
}

func TestSerializeEdited_QuotedAndUnquotedValues_KeepTheirStyle(t *testing.T) {
	src := "[Info]\nname = \"A\"\n[Files]\ncns =   a.cns\nsprite = \"a.sff\"\n"
	want := "[Info]\nname = \"B\"\n[Files]\ncns =   b.cns\nsprite = \"b.sff\"\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.Name = "B"; i.ConstantsFile = "b.cns"; i.SpriteFile = "b.sff" })
}

func TestSerializeEdited_UnknownSectionsKeysAndComments_Untouched(t *testing.T) {
	src := "; top\n[Info]\nNAME = A\nfoo = bar ; keep\n\n[Arcade]\nintro.storyboard = x.def\n[Files]\n; note\nCNS = a.cns\nweird = 1\n"
	want := "; top\n[Info]\nNAME = Z\nfoo = bar ; keep\n\n[Arcade]\nintro.storyboard = x.def\n[Files]\n; note\nCNS = a.cns\nweird = 1\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.Name = "Z" })
}

func TestSerializeEdited_ClearedField_RemovesOnlyItsLine(t *testing.T) {
	src := "[Info]\nname = \"A\"\nauthor = \"B\" ; x\n\n[Files]\ncns = a.cns\n"
	want := "[Info]\nname = \"A\"\n\n[Files]\ncns = a.cns\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.Author = "" })
}

func TestSerializeEdited_AddedPalette_InsertsAfterLastKeyLineOfFiles(t *testing.T) {
	src := "[Files]\npal1 = a.act\npal2 = b.act ; two\n\n[Arcade]\nx = 1\n"
	want := "[Files]\npal1 = a.act\npal2 = b.act ; two\npal3 = c.act\n\n[Arcade]\nx = 1\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.Palettes = append(i.Palettes, "c.act") })
}

func TestSerializeEdited_RemovedPaletteAndChangedStateFile_AlignByIndex(t *testing.T) {
	src := "[Files]\nst = a.st\nst1 = b.st\npal2 = q.act\npal1 = p.act\n"
	want := "[Files]\nst = a.st\nst1 = c.st\npal1 = p.act\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.StateFiles[1] = "c.st"; i.Palettes = i.Palettes[:1] })
}

func TestSerializeEdited_FieldAddedToMissingSection_AppendsSectionAtEnd(t *testing.T) {
	src := "[Info]\nname = \"A\"\n; end"
	want := "[Info]\nname = \"A\"\n; end\n\n[Files]\nsprite = a.sff\ncns = a.cns\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.SpriteFile = "a.sff"; i.ConstantsFile = "a.cns" })
}

func TestSerializeEdited_NewNameInExistingInfo_IsQuotedAndInsertedAfterLastKey(t *testing.T) {
	src := "[Info]\nauthor = \"B\"\n; tail\n[Files]\n"
	want := "[Info]\nauthor = \"B\"\nname = \"N\"\n; tail\n[Files]\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.Name = "N" })
}

func TestSerializeEdited_RepeatedKey_EditsTheLineParseUses(t *testing.T) {
	src := "[Info]\nname = first\nname = second\n"
	want := "[Info]\nname = first\nname = third\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.Name = "third" })
}

type failWriter struct{ err error }

func (f failWriter) Write([]byte) (int, error) { return 0, f.err }

func TestSerializeEdited_WriterFailure_ReturnsWrappedError(t *testing.T) {
	doc, _ := ParseDocument(strings.NewReader("[Info]\nname = A\n"))
	boom := errors.New("boom")
	err := doc.SerializeEdited(failWriter{boom}, doc.Info)
	if !errors.Is(err, boom) || !strings.HasPrefix(err.Error(), "def: ") {
		t.Fatalf("want wrapped boom, got %v", err)
	}
}

func reparse(t *testing.T, src string) CharacterInfo {
	t.Helper()
	info, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}
	return info
}

func TestSerializeEdited_AddedPaletteToSparseNumbering_UsesMaxPlusOne(t *testing.T) {
	src := "[Files]\npal1 = a.act\npal4 = d.act\npal5 = e.act\n"
	want := src + "pal6 = f.act\n"
	got := editedOutput(t, src, func(i *CharacterInfo) { i.Palettes = append(i.Palettes, "f.act") })
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	if p := reparse(t, got).Palettes; strings.Join(p, ",") != "a.act,d.act,e.act,f.act" {
		t.Errorf("reparsed order = %v", p)
	}
}

func TestSerializeEdited_AddedStateFileToSparseNumbering_UsesMaxPlusOne(t *testing.T) {
	src := "[Files]\nst = a.st\nst2 = c.st\n"
	want := src + "st3 = d.st\n"
	got := editedOutput(t, src, func(i *CharacterInfo) { i.StateFiles = append(i.StateFiles, "d.st") })
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	if s := reparse(t, got).StateFiles; strings.Join(s, ",") != "a.st,c.st,d.st" {
		t.Errorf("reparsed order = %v", s)
	}
}

func TestSerializeEdited_StateFilesAddedToFilesWithoutAny_NumbersStThenSt1(t *testing.T) {
	src := "[Files]\ncns = a.cns\n"
	want := src + "st = a.st\nst1 = b.st\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.StateFiles = []string{"a.st", "b.st"} })
}

func TestSerializeEdited_ClearedFieldWithDuplicateKey_RemovesEveryLine(t *testing.T) {
	src := "[Info]\nname = first\nauthor = x\nname = second\n"
	want := "[Info]\nauthor = x\n"
	got := editedOutput(t, src, func(i *CharacterInfo) { i.Name = "" })
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	if n := reparse(t, got).Name; n != "" {
		t.Errorf("reparsed name = %q", n)
	}
}

func TestSerializeEdited_FieldAddedAfterLastLineWithoutNewline_TerminatesItFirst(t *testing.T) {
	src := "[Info]\nname = A"
	want := "[Info]\nname = A\nauthor = \"B\"\n"
	assertEdited(t, src, want, func(i *CharacterInfo) { i.Author = "B" })
}
