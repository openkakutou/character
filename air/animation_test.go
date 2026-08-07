package air

import "testing"

func TestAnimation_ZeroValue_HasNoFramesAndLoopsFromStart(t *testing.T) {
	var a Animation

	if a.Frames != nil {
		t.Errorf("expected zero-value Animation to have nil Frames, got %v", a.Frames)
	}
	if len(a.Frames) != 0 {
		t.Errorf("expected zero-value Animation to have 0 frames, got %d", len(a.Frames))
	}
	if a.LoopStart != 0 {
		t.Errorf("expected zero-value Animation to have LoopStart 0 (loop from the first frame, matching .air's default when no Loopstart marker is present), got %d", a.LoopStart)
	}
	if a.Number != 0 {
		t.Errorf("expected zero-value Animation to have Number 0, got %d", a.Number)
	}

	// Ranging over a nil Frames slice must not panic.
	for range a.Frames {
		t.Fatal("expected no iterations over a nil Frames slice")
	}
}

func TestFrame_ZeroValue_HasNoClsnBoxes(t *testing.T) {
	var f Frame

	if f.Clsn1 != nil {
		t.Errorf("expected zero-value Frame to have nil Clsn1, got %v", f.Clsn1)
	}
	if f.Clsn2 != nil {
		t.Errorf("expected zero-value Frame to have nil Clsn2, got %v", f.Clsn2)
	}
	if f.Flip != FlipNone {
		t.Errorf("expected zero-value Frame to have Flip FlipNone, got %q", f.Flip)
	}
	if f.Blend != "" {
		t.Errorf("expected zero-value Frame to have empty Blend, got %q", f.Blend)
	}

	// Ranging over nil Clsn1/Clsn2 slices must not panic.
	for range f.Clsn1 {
		t.Fatal("expected no iterations over a nil Clsn1 slice")
	}
	for range f.Clsn2 {
		t.Fatal("expected no iterations over a nil Clsn2 slice")
	}
}

func TestClsnBox_ZeroValue_AllCoordinatesZero(t *testing.T) {
	var box ClsnBox

	if box.Left != 0 || box.Top != 0 || box.Right != 0 || box.Bottom != 0 {
		t.Errorf("expected zero-value ClsnBox to have all coordinates at 0, got %+v", box)
	}
}

func TestAnimation_WithFrames_PreservesAssignedFieldValues(t *testing.T) {
	a := Animation{
		Number: 200,
		Frames: []Frame{
			{
				Group: 5,
				Image: 12,
				X:     -10,
				Y:     3,
				Time:  4,
				Flip:  FlipH,
				Blend: BlendMode("A"),
				Clsn1: []ClsnBox{{Left: -3, Top: -73, Right: 15, Bottom: -1}},
				Clsn2: []ClsnBox{{Left: -2, Top: -70, Right: 12, Bottom: 0}},
			},
		},
		LoopStart: 1,
	}

	if a.Number != 200 {
		t.Errorf("expected Number 200, got %d", a.Number)
	}
	if len(a.Frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(a.Frames))
	}

	f := a.Frames[0]
	if f.Group != 5 || f.Image != 12 {
		t.Errorf("expected Group 5 and Image 12, got Group %d Image %d", f.Group, f.Image)
	}
	if f.X != -10 || f.Y != 3 {
		t.Errorf("expected X -10 and Y 3, got X %d Y %d", f.X, f.Y)
	}
	if f.Time != 4 {
		t.Errorf("expected Time 4, got %d", f.Time)
	}
	if f.Flip != FlipH {
		t.Errorf("expected Flip FlipH, got %q", f.Flip)
	}
	if f.Blend != BlendMode("A") {
		t.Errorf("expected Blend \"A\", got %q", f.Blend)
	}

	if len(f.Clsn1) != 1 || f.Clsn1[0] != (ClsnBox{Left: -3, Top: -73, Right: 15, Bottom: -1}) {
		t.Errorf("expected Clsn1 box {-3,-73,15,-1}, got %+v", f.Clsn1)
	}
	if len(f.Clsn2) != 1 || f.Clsn2[0] != (ClsnBox{Left: -2, Top: -70, Right: 12, Bottom: 0}) {
		t.Errorf("expected Clsn2 box {-2,-70,12,0}, got %+v", f.Clsn2)
	}

	if a.LoopStart != 1 {
		t.Errorf("expected LoopStart 1, got %d", a.LoopStart)
	}
}

func TestFrame_IsBlank_TrueWhenGroupAndImageAreBothMinusOne(t *testing.T) {
	f := Frame{Group: -1, Image: -1, X: 0, Y: 0, Time: 5}

	if !f.IsBlank() {
		t.Errorf("expected IsBlank true for the -1,-1 sentinel, got false")
	}
}

func TestFrame_IsBlank_TrueWhenOnlyOneFieldIsMinusOne(t *testing.T) {
	cases := []struct {
		name  string
		frame Frame
	}{
		{"group only", Frame{Group: -1, Image: 0}},
		{"image only", Frame{Group: 0, Image: -1}},
	}

	for _, c := range cases {
		if !c.frame.IsBlank() {
			t.Errorf("%s: expected IsBlank true, got false", c.name)
		}
	}
}

func TestFrame_IsBlank_FalseForOrdinarySpriteReference(t *testing.T) {
	f := Frame{Group: 0, Image: 0}

	if f.IsBlank() {
		t.Errorf("expected IsBlank false for an ordinary (0, 0) reference, got true")
	}
}

func TestFlip_Constants_MatchAirFormatTokens(t *testing.T) {
	cases := []struct {
		name string
		flip Flip
		want string
	}{
		{"none", FlipNone, ""},
		{"horizontal", FlipH, "H"},
		{"vertical", FlipV, "V"},
		{"both", FlipHV, "HV"},
	}

	for _, c := range cases {
		if string(c.flip) != c.want {
			t.Errorf("%s: expected token %q, got %q", c.name, c.want, string(c.flip))
		}
	}
}
