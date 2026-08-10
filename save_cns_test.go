package character

import (
	"strings"
	"testing"

	"github.com/openkakutou/character/cns"
)

func TestSerializeCns_NoEdits_ProducesByteIdenticalOutput(t *testing.T) {
	original := []byte(`[Statedef 200]
type = S
movetype = A
physics = S
ctrl = 1

[State 200, 0]
type = ChangeState
trigger1 = Time = 0
value = 0
`)

	states, err := cns.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: cns.Parse failed: %v", err)
	}

	out, err := SerializeCns(original, states)
	if err != nil {
		t.Fatalf("SerializeCns returned an error: %v", err)
	}

	if string(out) != string(original) {
		t.Fatalf("expected byte-identical output on no edits\n--- got ---\n%s\n--- want ---\n%s", out, original)
	}
}

func TestSerializeCns_WithEdits_ReflectsEditedValues(t *testing.T) {
	original := []byte(`[Statedef 200]
type = S
movetype = A
physics = S
ctrl = 1

[State 200, 0]
type = ChangeState
trigger1 = Time = 0
value = 0
`)

	states, err := cns.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: cns.Parse failed: %v", err)
	}
	states[0].Ctrl = false

	out, err := SerializeCns(original, states)
	if err != nil {
		t.Fatalf("SerializeCns returned an error: %v", err)
	}
	if string(out) == string(original) {
		t.Fatalf("expected output to differ from original once edited, got identical bytes")
	}

	roundTripped, err := cns.Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if roundTripped[0].Ctrl != false {
		t.Fatalf("expected edited Ctrl to survive round trip, got %v", roundTripped[0].Ctrl)
	}
}

func TestSerializeCns_EmptyOriginal_SerializesFreshForNewStates(t *testing.T) {
	states := []cns.StateDef{
		{Number: 5000, Type: cns.StateType("S")},
	}

	out, err := SerializeCns(nil, states)
	if err != nil {
		t.Fatalf("SerializeCns returned an error: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("expected non-empty output for brand new states")
	}

	roundTripped, err := cns.Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if len(roundTripped) != 1 || roundTripped[0].Number != 5000 {
		t.Fatalf("expected the new state to survive round trip, got %+v", roundTripped)
	}
}

func TestSerializeCns_MalformedOriginal_ReturnsDescriptiveError(t *testing.T) {
	malformed := []byte("[Statedef abc]\ntype = S\n")

	_, err := SerializeCns(malformed, nil)
	if err == nil {
		t.Fatalf("expected an error for a malformed original .cns, got nil")
	}
	if !strings.Contains(err.Error(), "cns:") {
		t.Fatalf("expected error to mention the cns package's own diagnostic, got: %v", err)
	}
}
