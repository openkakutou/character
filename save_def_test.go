package character

import (
	"strings"
	"testing"

	"github.com/openkakutou/character/def"
)

func TestSerializeDef_NoEdits_ProducesByteIdenticalOutput(t *testing.T) {
	original := []byte(`[Info]
name = Test Character
author = Test Author

[Files]
sprite = test.sff
anim = test.air
cns = test.cns
`)

	info, err := def.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: def.Parse failed: %v", err)
	}

	out, err := SerializeDef(original, info)
	if err != nil {
		t.Fatalf("SerializeDef returned an error: %v", err)
	}

	if string(out) != string(original) {
		t.Fatalf("expected byte-identical output on no edits\n--- got ---\n%s\n--- want ---\n%s", out, original)
	}
}

func TestSerializeDef_WithEdits_ReflectsEditedValues(t *testing.T) {
	original := []byte(`[Info]
name = Test Character
author = Test Author

[Files]
sprite = test.sff
anim = test.air
cns = test.cns
`)

	info, err := def.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: def.Parse failed: %v", err)
	}
	info.Name = "Renamed Character"

	out, err := SerializeDef(original, info)
	if err != nil {
		t.Fatalf("SerializeDef returned an error: %v", err)
	}

	if string(out) == string(original) {
		t.Fatalf("expected output to differ from original once edited, got identical bytes")
	}

	roundTripped, err := def.Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if roundTripped.Name != "Renamed Character" {
		t.Fatalf("expected edited Name to survive round trip, got %q", roundTripped.Name)
	}
	if roundTripped.Author != "Test Author" {
		t.Fatalf("expected untouched Author to survive round trip, got %q", roundTripped.Author)
	}
}

func TestSerializeDef_EmptyOriginal_SerializesFreshForNewCharacter(t *testing.T) {
	info := def.CharacterInfo{
		Name:          "Brand New Character",
		Author:        "New Author",
		SpriteFile:    "new.sff",
		AnimationFile: "new.air",
		ConstantsFile: "new.cns",
	}

	out, err := SerializeDef(nil, info)
	if err != nil {
		t.Fatalf("SerializeDef returned an error: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("expected non-empty output for a brand new character")
	}

	roundTripped, err := def.Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if roundTripped.Name != info.Name {
		t.Fatalf("expected Name %q, got %q", info.Name, roundTripped.Name)
	}
}

func TestSerializeDef_MalformedOriginal_ReturnsDescriptiveError(t *testing.T) {
	malformed := []byte("[Info\nname = Test\n")

	_, err := SerializeDef(malformed, def.CharacterInfo{Name: "Test"})
	if err == nil {
		t.Fatalf("expected an error for a malformed original .def, got nil")
	}
	if !strings.Contains(err.Error(), "def:") {
		t.Fatalf("expected error to mention the def package's own diagnostic, got: %v", err)
	}
}

// TestSerializeDef_NoEdits_NilVsEmptySlicesDoNotCountAsEdited verifies that
// a caller round-tripping the baseline through JSON — where an absent/empty
// list becomes a non-nil empty slice instead of staying nil — is still
// treated as "no edits", matching how LoadBytes already normalizes its own
// JSON contract (see normalizeForJSON).
func TestSerializeDef_NoEdits_NilVsEmptySlicesDoNotCountAsEdited(t *testing.T) {
	original := []byte(`[Info]
name = Test Character

[Files]
sprite = test.sff
`)

	info, err := def.Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: def.Parse failed: %v", err)
	}
	// info.StateFiles/Palettes are nil here (nothing in the source);
	// simulate what a JSON round trip through a JS caller would produce.
	info.StateFiles = []string{}
	info.Palettes = []string{}

	out, err := SerializeDef(original, info)
	if err != nil {
		t.Fatalf("SerializeDef returned an error: %v", err)
	}
	if string(out) != string(original) {
		t.Fatalf("expected byte-identical output when only nil-vs-empty-slice differs\n--- got ---\n%s\n--- want ---\n%s", out, original)
	}
}
