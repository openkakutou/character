package air

import (
	"bytes"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

func editedOutput(t *testing.T, src string, mutate func(a []Animation) []Animation) string {
	t.Helper()
	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	anims := mutate(doc.Animations)
	var buf bytes.Buffer
	if err := doc.SerializeEdited(&buf, anims); err != nil {
		t.Fatalf("SerializeEdited: %v", err)
	}
	return buf.String()
}

func TestSerializeEditedChangingOneFrameTimeRewritesOnlyThatLine(t *testing.T) {
	src := "; stand\n[Begin Action 0]\n0,0, 0,0, 5 ; first\n0,1, 0,0, 5\n\n[Begin Action 1]\n1,0, 0,0, 3\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames[0].Time = 9
		return a
	})
	want := "; stand\n[Begin Action 0]\n0,0, 0,0, 9 ; first\n0,1, 0,0, 5\n\n[Begin Action 1]\n1,0, 0,0, 3\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedNoOpIsByteExact(t *testing.T) {
	src := "; banner\r\n:odd\r\n[Begin Action 0]\r\nInterpolate Offset\r\nClsn2Default: 1\r\n Clsn2 [0] = 1,2,3,4\r\n0,0, 0,0, 5 junk ;c\r\nLoopstart\r\n\r\n[Statedef 5]\r\ntype = S\r\n"
	got := editedOutput(t, src, func(a []Animation) []Animation { return a })
	if got != src {
		t.Errorf("got %q\nwant %q", got, src)
	}
}

func TestSerializeEditedRenumberedActionKeepsItsComment(t *testing.T) {
	src := "; walk\n[begin action 10]\n0,0, 0,0, 5\n\n; run\n[Begin Action 20]\n1,0, 0,0, 5\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Number = 11
		return a
	})
	want := "; walk\n[begin action 11]\n0,0, 0,0, 5\n\n; run\n[Begin Action 20]\n1,0, 0,0, 5\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedInsertsNewActionBetweenTwo(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 5\n\n; two\n[Begin Action 2]\n2,0, 0,0, 5\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		n := Animation{Number: 1, Frames: []Frame{{Group: 1, Time: 4}}}
		return []Animation{a[0], n, a[1]}
	})
	want := "[Begin Action 0]\n0,0, 0,0, 5\n\n[Begin Action 1]\n1,0, 0,0, 4\n\n; two\n[Begin Action 2]\n2,0, 0,0, 5\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedDeletingActionRemovesItsCommentBlock(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 5\n\n; about one\n; more\n[Begin Action 1]\n1,0, 0,0, 5\n\n[Begin Action 2]\n2,0, 0,0, 5\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		return []Animation{a[0], a[2]}
	})
	want := "[Begin Action 0]\n0,0, 0,0, 5\n\n[Begin Action 2]\n2,0, 0,0, 5\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedCRLFFileKeepsLineEndings(t *testing.T) {
	src := "[Begin Action 0]\r\n0,0, 0,0, 5 ;c\r\n\r\n[Begin Action 1]\r\n1,0, 0,0, 5\r\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames[0].Time = 7
		a = append(a, Animation{Number: 2, Frames: []Frame{{Group: 2, Time: 1}}})
		return a
	})
	want := "[Begin Action 0]\r\n0,0, 0,0, 7 ;c\r\n\r\n[Begin Action 1]\r\n1,0, 0,0, 5\r\n\r\n[Begin Action 2]\r\n2,0, 0,0, 1\r\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedClsnChangePatchesOnlyThatActionsBoxLine(t *testing.T) {
	src := "; a\n[Begin Action 0]\nClsn2: 1\n Clsn2[0] = 0,0,1,1\n0,0, 0,0, 5 ;x\n\n; b\n[Begin Action 1]\n1,0, 0,0, 5 ;y\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames[0].Clsn2 = []ClsnBox{{Left: 5, Top: 6, Right: 7, Bottom: 8}}
		return a
	})
	want := "; a\n[Begin Action 0]\nClsn2: 1\n Clsn2[0] = 5,6,7,8\n0,0, 0,0, 5 ;x\n\n; b\n[Begin Action 1]\n1,0, 0,0, 5 ;y\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedLeavesEmbeddedStatedefSectionUntouched(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 5\n\n[Statedef 5]\n; keep\ntype = S\n\n[Begin Action 1]\n1,0, 0,0, 5\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[1].Frames[0].X = 3
		return a
	})
	want := "[Begin Action 0]\n0,0, 0,0, 5\n\n[Statedef 5]\n; keep\ntype = S\n\n[Begin Action 1]\n1,0, 3,0, 5\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedWrapsWriterError(t *testing.T) {
	doc, err := ParseDocument(strings.NewReader("[Begin Action 0]\n0,0, 0,0, 5\n"))
	if err != nil {
		t.Fatal(err)
	}
	sentinel := errBoom
	err = doc.SerializeEdited(failingWriter{}, nil)
	if !errors.Is(err, sentinel) || !strings.HasPrefix(err.Error(), "air: ") {
		t.Errorf("got %v, want wrapped %v with air: prefix", err, sentinel)
	}
}

func TestSerializeEditedClsnBoxValueChangePatchesOnlyThatBoxLine(t *testing.T) {
	src := "[Begin Action 0]\nInterpolate Offset\nClsn2: 2 ;decl\n Clsn2 [0] = 1,2,3,4 ;keep\n Clsn2[1] = 5,6,7,8\n0,0, 0,0, 5 ;f\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames[0].Clsn2[0].Left = 9
		return a
	})
	want := "[Begin Action 0]\nInterpolate Offset\nClsn2: 2 ;decl\n Clsn2 [0] = 9,2,3,4 ;keep\n Clsn2[1] = 5,6,7,8\n0,0, 0,0, 5 ;f\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedClsnBoxCountChangeRewritesOnlyThatBlock(t *testing.T) {
	src := "[Begin Action 0]\nClsn2: 2 ;decl\n Clsn2 [0] = 1,2,3,4 ;keep\n Clsn2[1] = 5,6,7,8\n; note\n0,0, 0,0, 5 ;f\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames[0].Clsn2 = append(a[0].Frames[0].Clsn2, ClsnBox{9, 9, 9, 9})
		return a
	})
	want := "[Begin Action 0]\nClsn2: 3 ;decl\n Clsn2[0] = 1,2,3,4\n Clsn2[1] = 5,6,7,8\n Clsn2[2] = 9,9,9,9\n; note\n0,0, 0,0, 5 ;f\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedChangedDefaultBoxesRegenerateTheAction(t *testing.T) {
	src := "[Begin Action 0]\nClsn2Default: 1\n Clsn2[0] = 1,1,2,2\n0,0, 0,0, 5\n0,1, 0,0, 5\n\n[Begin Action 1]\n1,0, 0,0, 5 ;y\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames[0].Clsn2 = []ClsnBox{{3, 3, 4, 4}}
		return a
	})
	want := "[Begin Action 0]\nClsn2: 1\n Clsn2[0] = 3,3,4,4\n0,0, 0,0, 5\nClsn2: 1\n Clsn2[0] = 1,1,2,2\n0,1, 0,0, 5\n\n[Begin Action 1]\n1,0, 0,0, 5 ;y\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedDeletedFrameRemovesItsLineAndOwnClsnBlock(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 1\n; mid\nClsn1: 1\n Clsn1[0] = 1,2,3,4\n0,1, 0,0, 2 ;two\n0,2, 0,0, 3\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames = []Frame{a[0].Frames[0], a[0].Frames[2]}
		return a
	})
	want := "[Begin Action 0]\n0,0, 0,0, 1\n; mid\n0,2, 0,0, 3\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedInsertedFrameGoesRightAfterItsPredecessorWithItsOwnClsn(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 1 ;one\nInterpolate Offset\n0,2, 0,0, 3\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		f := Frame{Group: 7, Time: 4, Clsn1: []ClsnBox{{1, 2, 3, 4}}}
		a[0].Frames = []Frame{a[0].Frames[0], f, a[0].Frames[1]}
		return a
	})
	want := "[Begin Action 0]\n0,0, 0,0, 1 ;one\nClsn1: 1\n Clsn1[0] = 1,2,3,4\n7,0, 0,0, 4\nInterpolate Offset\n0,2, 0,0, 3\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedInsertedFrameMatchingDefaultGetsNoClsnBlock(t *testing.T) {
	src := "[Begin Action 0]\nClsn2Default: 1\n Clsn2[0] = 1,1,2,2\n0,0, 0,0, 5\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		f := Frame{Group: 0, Image: 1, Time: 6, Clsn2: []ClsnBox{{1, 1, 2, 2}}}
		a[0].Frames = append(a[0].Frames, f)
		return a
	})
	want := "[Begin Action 0]\nClsn2Default: 1\n Clsn2[0] = 1,1,2,2\n0,0, 0,0, 5\n0,1, 0,0, 6\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedLoopStartChangeMovesOnlyTheMarkerLine(t *testing.T) {
	src := "[Begin Action 0]\nInterpolate Offset\n0,0, 0,0, 1\nLoopstart ;lp\n0,1, 0,0, 1\n0,2, 0,0, 1\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].LoopStart = 2
		return a
	})
	want := "[Begin Action 0]\nInterpolate Offset\n0,0, 0,0, 1\n0,1, 0,0, 1\nLoopstart\n0,2, 0,0, 1\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedDeletingFrameBeforeLoopMarkerKeepsTheMarkerInPlace(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 1\nLoopstart\n0,1, 0,0, 2\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames = a[0].Frames[1:]
		a[0].LoopStart = 0
		return a
	})
	want := "[Begin Action 0]\nLoopstart\n0,1, 0,0, 2\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedChangedFrameKeepsIkemenExtraFieldsAndComment(t *testing.T) {
	src := "[Begin Action 200]\n200,0, 0,0, 5, , A, 1.5,1.5, 45 ;spin\r\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames[0].Time = 6
		return a
	})
	want := "[Begin Action 200]\n200,0, 0,0, 6, , A, 1.5,1.5, 45 ;spin\r\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedAddingFlipAndBlendAppendsThemToTheFrameLine(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 5 ;c\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames[0].Flip = FlipH
		a[0].Frames[0].Blend = "A"
		return a
	})
	want := "[Begin Action 0]\n0,0, 0,0, 5, H, A ;c\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func parseWithin(t *testing.T, src string) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic: %v", r)
			}
		}()
		_, _ = ParseDocument(strings.NewReader(src))
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ParseDocument did not return")
	}
}

func TestParseDocumentSurvivesHostileClsnCounts(t *testing.T) {
	for name, src := range map[string]string{
		"negative in statedef": "[Statedef 1]\nClsn1: -5\n[Begin Action 0]\n0,0, 0,0, 1\n",
		"huge in statedef":     "[Statedef 1]\nClsn1: 9223372036854775807\n[Begin Action 0]\n0,0, 0,0, 1\n",
		"negative in action":   "[Begin Action 0]\nClsn1: -5\n0,0, 0,0, 1\n",
		"huge in action":       "[Begin Action 0]\nClsn1: 9223372036854775807\n0,0, 0,0, 1\n",
		"huge default":         "[Begin Action 0]\nClsn1Default: 9223372036854775807\n0,0, 0,0, 1\n",
	} {
		t.Run(name, func(t *testing.T) { parseWithin(t, src) })
	}
}

func TestSerializeEditedInsertsActionBeforeFirstActionWhenNoPredecessor(t *testing.T) {
	src := "; banner\n\n; one\n[Begin Action 5]\n0,0, 0,0, 1\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		return append([]Animation{{Number: 1, Frames: []Frame{{Time: 2}}}}, a...)
	})
	want := "[Begin Action 1]\n0,0, 0,0, 2\n\n; banner\n\n; one\n[Begin Action 5]\n0,0, 0,0, 1\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedInsertsActionIntoFileWithoutActions(t *testing.T) {
	src := "; only a comment\n[Statedef 1]\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		return []Animation{{Number: 3, Frames: []Frame{{Time: 4}}}}
	})
	want := "[Begin Action 3]\n0,0, 0,0, 4\n\n; only a comment\n[Statedef 1]\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedInsertedFrameBeforeFirstFrameGoesAboveIt(t *testing.T) {
	src := "[Begin Action 0]\n; c\n0,1, 0,0, 5\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames = append([]Frame{{Image: 0, Time: 2}}, a[0].Frames...)
		return a
	})
	want := "[Begin Action 0]\n; c\n0,0, 0,0, 2\n0,1, 0,0, 5\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedBoxAddedToFrameWithoutDefaultGetsOwnBlockAboveIt(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 5\n0,1, 0,0, 5\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames[1].Clsn1 = []ClsnBox{{1, 2, 3, 4}}
		return a
	})
	want := "[Begin Action 0]\n0,0, 0,0, 5\nClsn1: 1\n Clsn1[0] = 1,2,3,4\n0,1, 0,0, 5\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedLoopStartMovedToTheEndPlacesMarkerAfterLastFrame(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 1\n0,1, 0,0, 1\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].LoopStart = 2
		return a
	})
	want := "[Begin Action 0]\n0,0, 0,0, 1\n0,1, 0,0, 1\nLoopstart\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedAmbiguousActionWithTwoLoopstartsIsRegenerated(t *testing.T) {
	src := "[Begin Action 0] ;h\nLoopstart\n0,0, 0,0, 1 ;keep?\nLoopstart\n0,1, 0,0, 1\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames[0].Time = 9
		return a
	})
	want := "[Begin Action 0]\n0,0, 0,0, 9\nLoopstart\n0,1, 0,0, 1\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedNoOpOnSampleFileIsByteExact(t *testing.T) {
	src, err := os.ReadFile("testdata/sample.air")
	if err != nil {
		t.Fatal(err)
	}
	if got := editedOutput(t, string(src), func(a []Animation) []Animation { return a }); got != string(src) {
		t.Errorf("sample.air changed by a no-op edit:\n%s", got)
	}
}

func TestSerializeEditedFrameInsertedBeforeFirstFrameStaysAboveLoopMarker(t *testing.T) {
	src := "[Begin Action 0]\nLoopstart\n0,0,0,0,5\n"
	var reparsed []Animation
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames = []Frame{{Group: 7, Time: 4}, a[0].Frames[0]}
		a[0].LoopStart = 1
		return a
	})
	want := "[Begin Action 0]\n7,0, 0,0, 4\nLoopstart\n0,0,0,0,5\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	reparsed, err := Parse(strings.NewReader(got))
	if err != nil || len(reparsed) != 1 || reparsed[0].LoopStart != 1 {
		t.Errorf("reparse: LoopStart/err = %+v, %v; want LoopStart 1", reparsed, err)
	}
}

func TestSerializeEditedFrameInsertedBeforeFirstFrameLandsAboveItsClsnBlock(t *testing.T) {
	src := "[Begin Action 0]\nClsn1: 1\n Clsn1[0] = 1,2,3,4\n0,0, 0,0, 5\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames = []Frame{{Group: 7, Time: 4}, a[0].Frames[0]}
		return a
	})
	want := "[Begin Action 0]\n7,0, 0,0, 4\nClsn1: 1\n Clsn1[0] = 1,2,3,4\n0,0, 0,0, 5\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedLoopMarkerMovingWithFrameCountChangeRegeneratesTheAction(t *testing.T) {
	src := "[Begin Action 0] ;h\n0,0, 0,0, 1 ;keep?\n0,1, 0,0, 1\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		a[0].Frames = append(a[0].Frames, Frame{Group: 7, Time: 4})
		a[0].LoopStart = 1
		return a
	})
	want := "[Begin Action 0]\n0,0, 0,0, 1\nLoopstart\n0,1, 0,0, 1\n7,0, 0,0, 4\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestSerializeEditedFrameInsertedAboveLoopMarkerBeforeClsnDefaultKeepsItsBoxes(t *testing.T) {
	src := "[Begin Action 0]\nLoopstart\nClsn2Default: 1\n Clsn2[0] = 1,2,3,4\n0,0, 0,0, 5\n"
	got := editedOutput(t, src, func(a []Animation) []Animation {
		n := a[0].Frames[0]
		n.Group = 7
		a[0].Frames = []Frame{n, a[0].Frames[0]}
		a[0].LoopStart = 1
		return a
	})
	reparsed, err := Parse(strings.NewReader(got))
	if err != nil || len(reparsed) != 1 || len(reparsed[0].Frames) != 2 || reparsed[0].LoopStart != 1 ||
		!slices.Equal(reparsed[0].Frames[0].Clsn2, []ClsnBox{{1, 2, 3, 4}}) ||
		!slices.Equal(reparsed[0].Frames[1].Clsn2, []ClsnBox{{1, 2, 3, 4}}) {
		t.Errorf("reparse of %q = %+v, %v; want 2 frames with Clsn2 {1,2,3,4} and LoopStart 1", got, reparsed, err)
	}
}
