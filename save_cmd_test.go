package character

import (
	"strings"
	"testing"

	"github.com/openkakutou/character/cmd"
)

func TestSerializeCmd_NoEdits_ProducesByteIdenticalOutput(t *testing.T) {
	original := []byte(`[Defaults]
command.time = 15
command.buffer.time = 1

[Command]
name = "a"
command = a
time = 1

[Statedef -1]

[State -1, HoldBack Detect]
type = VarSet
trigger1 = 1
value = 1
`)

	file, err := cmd.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: cmd.Parse failed: %v", err)
	}

	out, err := SerializeCmd(original, file)
	if err != nil {
		t.Fatalf("SerializeCmd returned an error: %v", err)
	}

	if string(out) != string(original) {
		t.Fatalf("expected byte-identical output on no edits\n--- got ---\n%s\n--- want ---\n%s", out, original)
	}
}

func TestSerializeCmd_WithEdits_ReflectsEditedValues(t *testing.T) {
	original := []byte(`[Defaults]
command.time = 15
command.buffer.time = 1

[Command]
name = "a"
command = a
time = 1
`)

	file, err := cmd.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: cmd.Parse failed: %v", err)
	}
	file.Commands[0].Input = "~D, DF, F, a"

	out, err := SerializeCmd(original, file)
	if err != nil {
		t.Fatalf("SerializeCmd returned an error: %v", err)
	}
	if string(out) == string(original) {
		t.Fatalf("expected output to differ from original once edited, got identical bytes")
	}

	roundTripped, err := cmd.Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if roundTripped.Commands[0].Input != "~D, DF, F, a" {
		t.Fatalf("expected edited Input to survive round trip, got %q", roundTripped.Commands[0].Input)
	}
}

func TestSerializeCmd_EmptyOriginal_SerializesFreshForNewCommandFile(t *testing.T) {
	file := cmd.CommandFile{
		Commands: []cmd.Command{{Name: "QCF_a", Input: "~D, DF, F, a"}},
	}

	out, err := SerializeCmd(nil, file)
	if err != nil {
		t.Fatalf("SerializeCmd returned an error: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("expected non-empty output for a brand new command file")
	}

	roundTripped, err := cmd.Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if len(roundTripped.Commands) != 1 || roundTripped.Commands[0].Name != "QCF_a" {
		t.Fatalf("expected the new command to survive round trip, got %+v", roundTripped.Commands)
	}
}

func TestSerializeCmd_MalformedOriginal_ReturnsDescriptiveError(t *testing.T) {
	malformed := []byte("[Defaults\ncommand.time = 15\n")

	_, err := SerializeCmd(malformed, cmd.CommandFile{})
	if err == nil {
		t.Fatalf("expected an error for a malformed original .cmd, got nil")
	}
	if !strings.Contains(err.Error(), "cmd:") {
		t.Fatalf("expected error to mention the cmd package's own diagnostic, got: %v", err)
	}
}
