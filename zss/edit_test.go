package zss

import (
	"errors"
	"os"
	"strings"
	"testing"
)

const editSrc = "# banner\n" +
	"\n" +
	"# state 100 comment\n" +
	"[Statedef 100; type: S;\n" +
	"  physics: N]\n" +
	"if 1 {\n" +
	"  a := 1\n" +
	"}\n" +
	"\n" +
	"# helper comment\n" +
	"[Function Foo(a, b) r]\n" +
	"r := a + b\n"

func parseEdit(t *testing.T, src string) *Document {
	t.Helper()
	d, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func edited(t *testing.T, src string, mutate func(*Script)) string {
	t.Helper()
	d := parseEdit(t, src)
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	mutate(&s)
	var sb strings.Builder
	if err := d.SerializeEdited(&sb, s); err != nil {
		t.Fatal(err)
	}
	return sb.String()
}

func TestSerializeEditedNoOpIsByteExact(t *testing.T) {
	src := strings.ReplaceAll(editSrc, "\n", "\r\n") + "tail without newline"
	got := edited(t, src, func(*Script) {})
	if got != src {
		t.Fatalf("got %q", got)
	}
}

func TestSerializeEditedOneBodyLineChangesOnlyThatLine(t *testing.T) {
	got := edited(t, editSrc, func(s *Script) {
		s.Blocks[0].Body = strings.Replace(s.Blocks[0].Body, "a := 1", "a := 2", 1)
	})
	want := strings.Replace(editSrc, "  a := 1\n", "  a := 2\n", 1)
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedHeaderParamChangeRewritesHeaderOnly(t *testing.T) {
	got := edited(t, editSrc, func(s *Script) { s.Blocks[0].HeaderParams["physics"] = "A" })
	want := "# banner\n\n# state 100 comment\n" +
		"[Statedef 100; physics: A; type: S]\n" +
		"if 1 {\n  a := 1\n}\n\n# helper comment\n[Function Foo(a, b) r]\nr := a + b\n"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedKeepsWrappedHeaderWhenOnlyBodyChanges(t *testing.T) {
	got := edited(t, editSrc, func(s *Script) { s.Blocks[1].Body = "r := a * b\n" })
	want := strings.Replace(editSrc, "r := a + b", "r := a * b", 1)
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestSerializeEditedInsertsNewBlockAfterPredecessorBeforeNextComment(t *testing.T) {
	got := edited(t, editSrc, func(s *Script) {
		nb := Block{Kind: BlockKindStatedef, Number: 200, Body: "x := 1\n"}
		s.Blocks = []Block{s.Blocks[0], nb, s.Blocks[1]}
	})
	want := "# banner\n\n# state 100 comment\n" +
		"[Statedef 100; type: S;\n  physics: N]\nif 1 {\n  a := 1\n}\n" +
		"[Statedef 200]\nx := 1\n" +
		"\n# helper comment\n[Function Foo(a, b) r]\nr := a + b\n"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedInsertsFirstBlockAfterPreamble(t *testing.T) {
	got := edited(t, "# only banner\n", func(s *Script) {
		s.Blocks = []Block{{Kind: BlockKindFunction, Name: "F", Body: "y := 1\n"}}
	})
	want := "# only banner\n[Function F()]\ny := 1\n"
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestSerializeEditedDeletesBlockWithItsLeadingComment(t *testing.T) {
	got := edited(t, editSrc, func(s *Script) { s.Blocks = s.Blocks[:1] })
	// "\n# helper comment\n" trivia of Foo goes; the blank line after "}" is trivia too.
	want := "# banner\n\n# state 100 comment\n" +
		"[Statedef 100; type: S;\n  physics: N]\nif 1 {\n  a := 1\n}\n"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedDeletingFirstBlockKeepsNextBlockComment(t *testing.T) {
	got := edited(t, editSrc, func(s *Script) { s.Blocks = s.Blocks[1:] })
	// the banner is contiguous with the deleted block's comment, so it is that block's trivia
	want := "\n# helper comment\n[Function Foo(a, b) r]\nr := a + b\n"
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestSerializeEditedRenumberKeepsLeadingComment(t *testing.T) {
	got := edited(t, editSrc, func(s *Script) { s.Blocks[0].Number = 101 })
	want := "# banner\n\n# state 100 comment\n" +
		"[Statedef 101; physics: N; type: S]\n" +
		"if 1 {\n  a := 1\n}\n\n# helper comment\n[Function Foo(a, b) r]\nr := a + b\n"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedGeneratedLinesUseCRLF(t *testing.T) {
	src := "[Statedef 1]\r\na := 1\r\nb := 2\r\n"
	got := edited(t, src, func(s *Script) {
		s.Blocks[0].Body = "a := 1\nc := 3\nb := 2\n"
		s.Blocks = append(s.Blocks, Block{Kind: BlockKindStatedef, Number: 2, Body: "z := 0\n"})
	})
	want := "[Statedef 1]\r\na := 1\r\nc := 3\r\nb := 2\r\n[Statedef 2]\r\nz := 0\r\n"
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestSerializeEditedChangedPreambleLeavesBlocksUntouched(t *testing.T) {
	got := edited(t, editSrc, func(s *Script) {
		s.Preamble = "# new banner\n\n# state 100 comment\n"
	})
	want := strings.Replace(editSrc, "# banner\n\n", "# new banner\n\n", 1)
	if got != want {
		t.Fatalf("got %q", got)
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func TestSerializeEditedReturnsWriterError(t *testing.T) {
	d := parseEdit(t, editSrc)
	s := d.Script
	s.Blocks = s.Blocks[:1]
	err := d.SerializeEdited(failWriter{}, s)
	if err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("err = %v", err)
	}
}

func TestSerializeEditedBodyEditedInTwoPlacesKeepsMiddleLinesVerbatim(t *testing.T) {
	src := "[Statedef 1; type: S;]\r\n" +
		"  a := 1\r\n" +
		"  keep1 := 0\n" +
		"\tkeep2 := 0\r\n" +
		"  b := 2\r\n"
	got := edited(t, src, func(s *Script) {
		s.Blocks[0].Body = "  a := 10\n  keep1 := 0\n\tkeep2 := 0\n  b := 20\n"
	})
	want := "[Statedef 1; type: S;]\r\n" +
		"  a := 10\r\n" +
		"  keep1 := 0\n" +
		"\tkeep2 := 0\r\n" +
		"  b := 20\r\n"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

const threeSrc = "# c1\n" +
	"[Statedef 1; type: S;]\n" +
	"a := 1\n" +
	"# c2\n" +
	"\n" +
	"[Statedef 2; type: S;]\n" +
	"b := 2\n" +
	"# c3\n" +
	"[Statedef 3; type: S;]\n" +
	"c := 3\n"

func TestSerializeEditedDeletingMiddleBlockRemovesItsCommentsOnly(t *testing.T) {
	got := edited(t, threeSrc, func(s *Script) {
		s.Blocks = []Block{s.Blocks[0], s.Blocks[2]}
	})
	want := "# c1\n" +
		"[Statedef 1; type: S;]\n" +
		"a := 1\n" +
		"# c3\n" +
		"[Statedef 3; type: S;]\n" +
		"c := 3\n"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedDeletingMiddleBlockWhileEditingPreviousBody(t *testing.T) {
	got := edited(t, threeSrc, func(s *Script) {
		s.Blocks[0].Body = "a := 5\n"
		s.Blocks = []Block{s.Blocks[0], s.Blocks[2]}
	})
	want := "# c1\n" +
		"[Statedef 1; type: S;]\n" +
		"a := 5\n" +
		"# c3\n" +
		"[Statedef 3; type: S;]\n" +
		"c := 3\n"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedNoOpOnSampleFixtureIsByteExact(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample.zss")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if got := edited(t, src, func(*Script) {}); got != src {
		t.Fatalf("got %q", got)
	}
	crlf := strings.ReplaceAll(src, "\n", "\r\n")
	if got := edited(t, crlf, func(*Script) {}); got != crlf {
		t.Fatalf("crlf got %q", got)
	}
}

const symSrc = "# dizzy\n[StateDef const(StateDizzy);\n  type: S]\nx := 1\n"

func TestSerializeEditedSymbolicBodyEditKeepsWrappedHeaderVerbatim(t *testing.T) {
	got := edited(t, symSrc, func(s *Script) { s.Blocks[0].Body = "x := 2\n" })
	want := strings.Replace(symSrc, "x := 1", "x := 2", 1)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSerializeEditedSymbolicNumberRawChangeRewritesHeader(t *testing.T) {
	got := edited(t, symSrc, func(s *Script) { s.Blocks[0].NumberRaw = "const(StateOther)" })
	want := "# dizzy\n[Statedef const(StateOther); type: S]\nx := 1\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSerializeEditedSymbolicToPlainNumberRewritesHeader(t *testing.T) {
	got := edited(t, "[Statedef +1]\nx\n", func(s *Script) { s.Blocks[0].NumberRaw = ""; s.Blocks[0].Number = 5 })
	if got != "[Statedef 5]\nx\n" {
		t.Fatalf("got %q", got)
	}
}
