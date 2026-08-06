package def

import (
	"errors"
	"strings"
	"testing"
)

func TestSerialize_FullCharacterInfo_RoundTripsThroughParseWithEquivalentStructure(t *testing.T) {
	info := CharacterInfo{
		Name:          "Kung Fu Man",
		Author:        "Elecbyte",
		SpriteFile:    "kfm.sff",
		AnimationFile: "kfm.air",
		SoundFile:     "kfm.snd",
		CommandFile:   "kfm.cmd",
		ConstantsFile: "kfm.cns",
		StateFiles:    []string{"kfm.st", "kfm_extra1.st", "kfm_extra2.st"},
		Palettes:      []string{"kfm1.act", "kfm2.act", "kfm3.act"},
	}

	var buf strings.Builder
	if err := Serialize(&buf, info); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, buf.String())
	}

	if got.Name != info.Name {
		t.Errorf("Name: expected %q, got %q", info.Name, got.Name)
	}
	if got.Author != info.Author {
		t.Errorf("Author: expected %q, got %q", info.Author, got.Author)
	}
	if got.SpriteFile != info.SpriteFile {
		t.Errorf("SpriteFile: expected %q, got %q", info.SpriteFile, got.SpriteFile)
	}
	if got.AnimationFile != info.AnimationFile {
		t.Errorf("AnimationFile: expected %q, got %q", info.AnimationFile, got.AnimationFile)
	}
	if got.SoundFile != info.SoundFile {
		t.Errorf("SoundFile: expected %q, got %q", info.SoundFile, got.SoundFile)
	}
	if got.CommandFile != info.CommandFile {
		t.Errorf("CommandFile: expected %q, got %q", info.CommandFile, got.CommandFile)
	}
	if got.ConstantsFile != info.ConstantsFile {
		t.Errorf("ConstantsFile: expected %q, got %q", info.ConstantsFile, got.ConstantsFile)
	}
	if len(got.StateFiles) != len(info.StateFiles) {
		t.Fatalf("StateFiles: expected %d entries, got %d: %v", len(info.StateFiles), len(got.StateFiles), got.StateFiles)
	}
	for i, want := range info.StateFiles {
		if got.StateFiles[i] != want {
			t.Errorf("StateFiles[%d]: expected %q, got %q (order should be preserved)", i, want, got.StateFiles[i])
		}
	}
	if len(got.Palettes) != len(info.Palettes) {
		t.Fatalf("Palettes: expected %d entries, got %d: %v", len(info.Palettes), len(got.Palettes), got.Palettes)
	}
	for i, want := range info.Palettes {
		if got.Palettes[i] != want {
			t.Errorf("Palettes[%d]: expected %q, got %q (number order should be preserved)", i, want, got.Palettes[i])
		}
	}
}

func TestSerialize_MinimalCharacterInfo_OmitsEmptyOptionalFieldsAndRoundTrips(t *testing.T) {
	info := CharacterInfo{
		Name:       "Minimal",
		SpriteFile: "minimal.sff",
	}

	var buf strings.Builder
	if err := Serialize(&buf, info); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()

	// Unset optional fields must not appear as keys in the output.
	for _, unwanted := range []string{"author", "anim", "sound", "cmd =", "cns ="} {
		if strings.Contains(strings.ToLower(output), unwanted) {
			t.Errorf("expected output to omit unset field %q, got:\n%s", unwanted, output)
		}
	}

	got, err := Parse(strings.NewReader(output))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, output)
	}
	if got.Name != "Minimal" || got.SpriteFile != "minimal.sff" {
		t.Errorf("expected Name %q and SpriteFile %q, got Name %q SpriteFile %q", "Minimal", "minimal.sff", got.Name, got.SpriteFile)
	}
	if got.Author != "" || got.AnimationFile != "" || got.SoundFile != "" || got.CommandFile != "" || got.ConstantsFile != "" {
		t.Errorf("expected all other fields to stay empty, got %+v", got)
	}
	if len(got.StateFiles) != 0 || len(got.Palettes) != 0 {
		t.Errorf("expected no StateFiles/Palettes, got %+v", got)
	}
}

func TestSerialize_EmptyCharacterInfo_ProducesValidTextThatReparsesToZeroValue(t *testing.T) {
	var buf strings.Builder
	if err := Serialize(&buf, CharacterInfo{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, buf.String())
	}

	var zero CharacterInfo
	if got.Name != zero.Name || got.Author != zero.Author || got.SpriteFile != zero.SpriteFile ||
		got.AnimationFile != zero.AnimationFile || got.SoundFile != zero.SoundFile ||
		got.CommandFile != zero.CommandFile || got.ConstantsFile != zero.ConstantsFile {
		t.Errorf("expected zero-value CharacterInfo after round trip, got %+v", got)
	}
	if len(got.StateFiles) != 0 || len(got.Palettes) != 0 {
		t.Errorf("expected no StateFiles/Palettes, got %+v", got)
	}
}

func TestSerialize_WriterError_ReturnsWrappedError(t *testing.T) {
	info := CharacterInfo{Name: "Kung Fu Man", SpriteFile: "kfm.sff"}

	err := Serialize(&failingWriter{}, info)
	if err == nil {
		t.Fatal("expected an error from a failing writer, got nil")
	}
	var target *failingWriterError
	if !errors.As(err, &target) {
		t.Errorf("expected wrapped failingWriterError, got: %v", err)
	}
}

// failingWriterError is a sentinel error type used to verify Serialize wraps
// (rather than discards) the underlying writer's error.
type failingWriterError struct{}

func (e *failingWriterError) Error() string { return "simulated write failure" }

// failingWriter always fails on Write.
type failingWriter struct{}

func (w *failingWriter) Write(p []byte) (int, error) {
	return 0, &failingWriterError{}
}
