package character

import (
	"strings"
	"testing"

	"github.com/openkakutou/character/air"
)

func TestSerializeAir_NoEdits_ProducesByteIdenticalOutput(t *testing.T) {
	original := []byte(`[Begin Action 200]
0,0, 0,0, 5
0,1, 10,10, 5
`)

	animations, err := air.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: air.Parse failed: %v", err)
	}

	out, err := SerializeAir(original, animations)
	if err != nil {
		t.Fatalf("SerializeAir returned an error: %v", err)
	}

	if string(out) != string(original) {
		t.Fatalf("expected byte-identical output on no edits\n--- got ---\n%s\n--- want ---\n%s", out, original)
	}
}

func TestSerializeAir_WithEdits_ReflectsEditedValues(t *testing.T) {
	original := []byte(`[Begin Action 200]
0,0, 0,0, 5
0,1, 10,10, 5
`)

	animations, err := air.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: air.Parse failed: %v", err)
	}
	animations[0].Frames[0].Time = 99

	out, err := SerializeAir(original, animations)
	if err != nil {
		t.Fatalf("SerializeAir returned an error: %v", err)
	}
	if string(out) == string(original) {
		t.Fatalf("expected output to differ from original once edited, got identical bytes")
	}

	roundTripped, err := air.Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if roundTripped[0].Frames[0].Time != 99 {
		t.Fatalf("expected edited Time to survive round trip, got %d", roundTripped[0].Frames[0].Time)
	}
}

func TestSerializeAir_EmptyOriginal_SerializesFreshForNewAnimations(t *testing.T) {
	animations := []air.Animation{
		{Number: 0, Frames: []air.Frame{{Group: 0, Image: 0, Time: 3}}},
	}

	out, err := SerializeAir(nil, animations)
	if err != nil {
		t.Fatalf("SerializeAir returned an error: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("expected non-empty output for brand new animations")
	}

	roundTripped, err := air.Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if len(roundTripped) != 1 || roundTripped[0].Number != 0 {
		t.Fatalf("expected the new animation to survive round trip, got %+v", roundTripped)
	}
}

func TestSerializeAir_MalformedOriginal_ReturnsDescriptiveError(t *testing.T) {
	malformed := []byte("[Begin Action abc]\n0,0, 0,0, 5\n")

	_, err := SerializeAir(malformed, nil)
	if err == nil {
		t.Fatalf("expected an error for a malformed original .air, got nil")
	}
	if !strings.Contains(err.Error(), "air:") {
		t.Fatalf("expected error to mention the air package's own diagnostic, got: %v", err)
	}
}
