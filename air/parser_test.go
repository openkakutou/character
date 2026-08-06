package air

import (
	"errors"
	"strings"
	"testing"
)

func TestParse_MultiActionSample_ProducesExpectedAnimations(t *testing.T) {
	src := `[Begin Action 200]
Clsn2Default: 1
 Clsn2[0] = -6,-92,6,0
0,0, 0,0, 5
0,1, 5,-2, 4, H
Clsn1: 1
 Clsn1[0] = -3,-73,15,-1
0,2, 0,0, 3, , A
Loopstart
0,3, 0,0, 6

[Begin Action 201]
0,0, 0,0, 10, HV, AS
`

	animations, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(animations) != 2 {
		t.Fatalf("expected 2 animations, got %d", len(animations))
	}

	a0 := animations[0]
	if a0.Number != 200 {
		t.Errorf("expected action number 200, got %d", a0.Number)
	}
	if len(a0.Frames) != 4 {
		t.Fatalf("expected 4 frames in action 200, got %d", len(a0.Frames))
	}
	// Loopstart appeared right before the 4th frame line (index 3).
	if a0.LoopStart != 3 {
		t.Errorf("expected LoopStart 3, got %d", a0.LoopStart)
	}

	defaultClsn2 := []ClsnBox{{Left: -6, Top: -92, Right: 6, Bottom: 0}}

	f0 := a0.Frames[0]
	want0 := Frame{Group: 0, Image: 0, X: 0, Y: 0, Time: 5, Flip: FlipNone, Blend: "", Clsn2: defaultClsn2}
	if f0.Group != want0.Group || f0.Image != want0.Image || f0.X != want0.X || f0.Y != want0.Y || f0.Time != want0.Time {
		t.Errorf("frame 0: expected %+v, got %+v", want0, f0)
	}
	if f0.Flip != FlipNone || f0.Blend != "" {
		t.Errorf("frame 0: expected no flip/blend, got Flip=%q Blend=%q", f0.Flip, f0.Blend)
	}
	if len(f0.Clsn2) != 1 || f0.Clsn2[0] != defaultClsn2[0] {
		t.Errorf("frame 0: expected default Clsn2 %+v, got %+v", defaultClsn2, f0.Clsn2)
	}
	if len(f0.Clsn1) != 0 {
		t.Errorf("frame 0: expected no Clsn1 boxes, got %+v", f0.Clsn1)
	}

	f1 := a0.Frames[1]
	if f1.Group != 0 || f1.Image != 1 || f1.X != 5 || f1.Y != -2 || f1.Time != 4 {
		t.Errorf("frame 1: unexpected geometry, got %+v", f1)
	}
	if f1.Flip != FlipH {
		t.Errorf("frame 1: expected FlipH, got %q", f1.Flip)
	}
	if f1.Blend != "" {
		t.Errorf("frame 1: expected no blend, got %q", f1.Blend)
	}
	if len(f1.Clsn2) != 1 || f1.Clsn2[0] != defaultClsn2[0] {
		t.Errorf("frame 1: expected the Clsn2Default to still apply, got %+v", f1.Clsn2)
	}

	f2 := a0.Frames[2]
	if f2.Group != 0 || f2.Image != 2 || f2.Time != 3 {
		t.Errorf("frame 2: unexpected geometry, got %+v", f2)
	}
	if f2.Flip != FlipNone {
		t.Errorf("frame 2: expected FlipNone (blank flip field), got %q", f2.Flip)
	}
	if f2.Blend != BlendMode("A") {
		t.Errorf("frame 2: expected blend \"A\", got %q", f2.Blend)
	}
	wantClsn1 := ClsnBox{Left: -3, Top: -73, Right: 15, Bottom: -1}
	if len(f2.Clsn1) != 1 || f2.Clsn1[0] != wantClsn1 {
		t.Errorf("frame 2: expected Clsn1 override %+v, got %+v", wantClsn1, f2.Clsn1)
	}
	if len(f2.Clsn2) != 1 || f2.Clsn2[0] != defaultClsn2[0] {
		t.Errorf("frame 2: expected the Clsn2Default to still apply, got %+v", f2.Clsn2)
	}

	f3 := a0.Frames[3]
	if f3.Group != 0 || f3.Image != 3 || f3.Time != 6 {
		t.Errorf("frame 3: unexpected geometry, got %+v", f3)
	}
	if len(f3.Clsn1) != 0 {
		t.Errorf("frame 3: expected the one-shot Clsn1 override from frame 2 not to carry over, got %+v", f3.Clsn1)
	}
	if len(f3.Clsn2) != 1 || f3.Clsn2[0] != defaultClsn2[0] {
		t.Errorf("frame 3: expected the Clsn2Default to still apply, got %+v", f3.Clsn2)
	}

	a1 := animations[1]
	if a1.Number != 201 {
		t.Errorf("expected action number 201, got %d", a1.Number)
	}
	if a1.LoopStart != 0 {
		t.Errorf("expected LoopStart 0 (no Loopstart marker), got %d", a1.LoopStart)
	}
	if len(a1.Frames) != 1 {
		t.Fatalf("expected 1 frame in action 201, got %d", len(a1.Frames))
	}
	g1 := a1.Frames[0]
	if g1.Time != 10 {
		t.Errorf("expected time 10, got %d", g1.Time)
	}
	if g1.Flip != FlipHV {
		t.Errorf("expected FlipHV, got %q", g1.Flip)
	}
	if g1.Blend != BlendMode("AS") {
		t.Errorf("expected blend \"AS\", got %q", g1.Blend)
	}
	if len(g1.Clsn1) != 0 || len(g1.Clsn2) != 0 {
		t.Errorf("expected action 201 to start with no inherited Clsn boxes from action 200, got Clsn1=%+v Clsn2=%+v", g1.Clsn1, g1.Clsn2)
	}
}

func TestParse_ActionWithNoFrames_ProducesEmptyAnimation(t *testing.T) {
	src := "[Begin Action 5]\n"

	animations, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(animations) != 1 {
		t.Fatalf("expected 1 animation, got %d", len(animations))
	}
	if animations[0].Number != 5 {
		t.Errorf("expected action number 5, got %d", animations[0].Number)
	}
	if len(animations[0].Frames) != 0 {
		t.Errorf("expected no frames, got %d", len(animations[0].Frames))
	}
	if animations[0].LoopStart != 0 {
		t.Errorf("expected LoopStart 0, got %d", animations[0].LoopStart)
	}
}

func TestParse_MultipleDefaultClsnBoxes_PreservesDeclarationOrder(t *testing.T) {
	src := `[Begin Action 0]
Clsn2Default: 2
 Clsn2[0] = -1,-2,3,4
 Clsn2[1] = -5,-6,7,8
0,0, 0,0, 1
`

	animations, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(animations) != 1 {
		t.Fatalf("expected 1 animation, got %d", len(animations))
	}
	frames := animations[0].Frames
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}
	got := frames[0].Clsn2
	want := []ClsnBox{
		{Left: -1, Top: -2, Right: 3, Bottom: 4},
		{Left: -5, Top: -6, Right: 7, Bottom: 8},
	}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("expected Clsn2 boxes %+v in declaration order, got %+v", want, got)
	}
}

func TestParse_FrameLineWithOnlyBlendField_LeavesFlipAsNone(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 1, , S\n"

	animations, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := animations[0].Frames[0]
	if f.Flip != FlipNone {
		t.Errorf("expected FlipNone when the flip field is blank, got %q", f.Flip)
	}
	if f.Blend != BlendMode("S") {
		t.Errorf("expected blend \"S\", got %q", f.Blend)
	}
}

func TestParse_FrameLineWithoutOptionalFields_DefaultsFlipAndBlend(t *testing.T) {
	src := "[Begin Action 0]\n1,2, 10,-20, 7\n"

	animations, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := animations[0].Frames[0]
	if f.Group != 1 || f.Image != 2 || f.X != 10 || f.Y != -20 || f.Time != 7 {
		t.Errorf("unexpected frame geometry: %+v", f)
	}
	if f.Flip != FlipNone {
		t.Errorf("expected FlipNone by default, got %q", f.Flip)
	}
	if f.Blend != "" {
		t.Errorf("expected empty Blend by default, got %q", f.Blend)
	}
}

// errReader is an io.Reader that always fails, used to exercise Parse's
// error path when the underlying source cannot be read at all.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("simulated read failure")
}

func TestParse_ReaderFailure_ReturnsErrorNotPanic(t *testing.T) {
	_, err := Parse(errReader{})
	if err == nil {
		t.Fatal("expected an error when the underlying reader fails, got nil")
	}
}

func TestParse_EmptyFile_ReturnsEmptyResultWithoutError(t *testing.T) {
	animations, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("expected no error for an empty file, got: %v", err)
	}
	if len(animations) != 0 {
		t.Errorf("expected an empty result for an empty file, got %d animations", len(animations))
	}
}

func TestParse_CommentLines_AreIgnoredAroundValidData(t *testing.T) {
	src := `; this whole action is a walk cycle
[Begin Action 0]
; the first frame holds the idle pose
0,0, 0,0, 5 ; trailing comment after a frame line
0,1, 3,-2, 4
`

	animations, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(animations) != 1 {
		t.Fatalf("expected 1 animation, got %d", len(animations))
	}
	frames := animations[0].Frames
	if len(frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames))
	}
	if frames[0].Group != 0 || frames[0].Image != 0 || frames[0].X != 0 || frames[0].Y != 0 || frames[0].Time != 5 {
		t.Errorf("frame 0: expected trailing comment to be stripped, got %+v", frames[0])
	}
	if frames[1].Group != 0 || frames[1].Image != 1 || frames[1].X != 3 || frames[1].Y != -2 || frames[1].Time != 4 {
		t.Errorf("frame 1: unexpected geometry, got %+v", frames[1])
	}
}

func TestParse_CommentOnlyFile_ReturnsEmptyResultWithoutError(t *testing.T) {
	src := "; nothing but comments here\n; still nothing\n"

	animations, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("expected no error for a comment-only file, got: %v", err)
	}
	if len(animations) != 0 {
		t.Errorf("expected an empty result, got %d animations", len(animations))
	}
}

func TestParse_MalformedActionHeaderAsFirstLine_ReturnsErrorNamingLine(t *testing.T) {
	src := "[Begin Action]\n0,0, 0,0, 1\n"

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a malformed action header, got nil")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("expected the error to name line 1, got: %v", err)
	}
}

func TestParse_MalformedActionHeaderMidFile_ReturnsErrorNamingLine(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 1\n[Begin Foo]\n"

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a malformed action header, got nil")
	}
	if !strings.Contains(err.Error(), "line 3") {
		t.Errorf("expected the error to name line 3, got: %v", err)
	}
}

func TestParse_FrameLineWithNonNumericField_ReturnsErrorNamingLine(t *testing.T) {
	src := "[Begin Action 0]\nabc,0, 0,0, 5\n"

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a non-numeric frame field, got nil")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected the error to name line 2, got: %v", err)
	}
}

func TestParse_FrameLineMissingFields_ReturnsError(t *testing.T) {
	src := "[Begin Action 0]\n0,0\n"

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a frame line missing required fields, got nil")
	}
}

func TestParse_NegativeGroupIndex_ReturnsDescriptiveError(t *testing.T) {
	src := "[Begin Action 0]\n-1,0, 0,0, 5\n"

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a negative group index, got nil")
	}
	if !strings.Contains(err.Error(), "group") {
		t.Errorf("expected the error to mention the group index, got: %v", err)
	}
}

func TestParse_NegativeImageIndex_ReturnsDescriptiveError(t *testing.T) {
	src := "[Begin Action 0]\n0,-1, 0,0, 5\n"

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a negative image index, got nil")
	}
	if !strings.Contains(err.Error(), "image") {
		t.Errorf("expected the error to mention the image index, got: %v", err)
	}
}

func TestParse_NegativeXYAndTime_AreStillAccepted(t *testing.T) {
	src := "[Begin Action 0]\n0,0, -5,-10, -1\n"

	animations, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := animations[0].Frames[0]
	if f.X != -5 || f.Y != -10 || f.Time != -1 {
		t.Errorf("expected negative X/Y/Time to be accepted, got %+v", f)
	}
}
