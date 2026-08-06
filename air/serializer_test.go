package air

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

// failingWriter returns errBoom on the very first Write call, simulating a
// downstream I/O failure (e.g. a full disk).
type failingWriter struct{}

var errBoom = errors.New("boom: simulated write failure")

func (failingWriter) Write([]byte) (int, error) {
	return 0, errBoom
}

func TestSerialize_MultiActionAnimation_ReparsesToEquivalentStructure(t *testing.T) {
	animations := []Animation{
		{
			Number: 200,
			Frames: []Frame{
				{Group: 0, Image: 0, X: 0, Y: 0, Time: 5},
				{Group: 0, Image: 1, X: 5, Y: -2, Time: 4, Flip: FlipH},
				{Group: 0, Image: 2, X: 0, Y: 0, Time: 3, Blend: "A"},
				{Group: 0, Image: 3, X: 0, Y: 0, Time: 6},
			},
			LoopStart: 3,
		},
		{
			Number: 201,
			Frames: []Frame{
				{Group: 1, Image: 0, X: 10, Y: 20, Time: 10, Flip: FlipHV, Blend: "AS"},
			},
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, animations); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := Parse(&buf)
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, buf.String())
	}

	if !reflect.DeepEqual(got, animations) {
		t.Errorf("round trip mismatch:\n want %+v\n got  %+v", animations, got)
	}
}

func TestSerialize_EmitsClsnBoxesPerFrame(t *testing.T) {
	animations := []Animation{
		{
			Number: 0,
			Frames: []Frame{
				{
					Group: 0, Image: 0, X: 0, Y: 0, Time: 1,
					Clsn1: []ClsnBox{{Left: -3, Top: -73, Right: 15, Bottom: -1}},
					Clsn2: []ClsnBox{{Left: -6, Top: -92, Right: 6, Bottom: 0}, {Left: -1, Top: -2, Right: 3, Bottom: 4}},
				},
				{Group: 0, Image: 1, X: 0, Y: 0, Time: 2},
			},
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, animations); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := Parse(&buf)
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, buf.String())
	}
	if !reflect.DeepEqual(got, animations) {
		t.Errorf("round trip mismatch:\n want %+v\n got  %+v", animations, got)
	}

	// Frame 1 declared no collision boxes at all; the output must not
	// fabricate any (e.g. via a leftover default from frame 0).
	if len(got[0].Frames[1].Clsn1) != 0 || len(got[0].Frames[1].Clsn2) != 0 {
		t.Errorf("expected frame 1 to have no Clsn boxes, got Clsn1=%+v Clsn2=%+v", got[0].Frames[1].Clsn1, got[0].Frames[1].Clsn2)
	}
}

func TestSerialize_EmitsLoopstartMarker_AtMidAnimation(t *testing.T) {
	animations := []Animation{
		{
			Number: 5,
			Frames: []Frame{
				{Group: 0, Image: 0, X: 0, Y: 0, Time: 1},
				{Group: 0, Image: 1, X: 0, Y: 0, Time: 1},
				{Group: 0, Image: 2, X: 0, Y: 0, Time: 1},
			},
			LoopStart: 2,
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, animations); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := Parse(&buf)
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if got[0].LoopStart != 2 {
		t.Errorf("expected LoopStart 2, got %d", got[0].LoopStart)
	}
}

func TestSerialize_LoopstartPastLastFrame_RoundTrips(t *testing.T) {
	// LoopStart equal to len(Frames) is a valid edge case: the marker sits
	// after the very last frame line, meaning the animation never loops.
	animations := []Animation{
		{
			Number: 5,
			Frames: []Frame{
				{Group: 0, Image: 0, X: 0, Y: 0, Time: 1},
				{Group: 0, Image: 1, X: 0, Y: 0, Time: 1},
			},
			LoopStart: 2,
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, animations); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := Parse(&buf)
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if got[0].LoopStart != 2 {
		t.Errorf("expected LoopStart 2, got %d", got[0].LoopStart)
	}
}

func TestSerialize_OmitsOptionalFieldsWhenUnset(t *testing.T) {
	animations := []Animation{
		{
			Number: 0,
			Frames: []Frame{
				{Group: 0, Image: 0, X: 0, Y: 0, Time: 1},
			},
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, animations); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, ",  ,") || strings.Contains(out, ", ,") {
		t.Errorf("expected no dangling empty Flip/Blend fields in output, got:\n%s", out)
	}

	got, err := Parse(strings.NewReader(out))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if got[0].Frames[0].Flip != FlipNone || got[0].Frames[0].Blend != "" {
		t.Errorf("expected zero-value Flip/Blend, got Flip=%q Blend=%q", got[0].Frames[0].Flip, got[0].Frames[0].Blend)
	}
}

func TestSerialize_EmptyAnimationsSlice_ProducesEmptyOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := Serialize(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for an empty animations slice, got %q", buf.String())
	}
}

func TestSerialize_ReturnsError_OnWriterFailure(t *testing.T) {
	animations := []Animation{
		{
			Number: 0,
			Frames: []Frame{
				{Group: 0, Image: 0, X: 0, Y: 0, Time: 1},
			},
		},
	}

	err := Serialize(failingWriter{}, animations)
	if err == nil {
		t.Fatal("expected an error when the writer fails, got nil")
	}
	if !errors.Is(err, errBoom) {
		t.Errorf("expected error to wrap the underlying writer failure, got: %v", err)
	}
}
