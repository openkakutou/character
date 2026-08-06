package def

import (
	"errors"
	"strings"
	"testing"
)

func TestParse_MultiSectionSample_ProducesExpectedCharacterInfo(t *testing.T) {
	src := `[Info]
name = "Kung Fu Man"
author = "Elecbyte"

[Files]
cmd = kfm.cmd
cns = kfm.cns
sprite = kfm.sff
anim = kfm.air
sound = kfm.snd
st = kfm.st
st1 = kfm_extra1.st
st2 = kfm_extra2.st
pal2 = kfm2.act
pal1 = kfm1.act
pal3 = kfm3.act
`

	info, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Name != "Kung Fu Man" {
		t.Errorf("expected Name %q, got %q", "Kung Fu Man", info.Name)
	}
	if info.Author != "Elecbyte" {
		t.Errorf("expected Author %q, got %q", "Elecbyte", info.Author)
	}
	if info.SpriteFile != "kfm.sff" {
		t.Errorf("expected SpriteFile %q, got %q", "kfm.sff", info.SpriteFile)
	}
	if info.AnimationFile != "kfm.air" {
		t.Errorf("expected AnimationFile %q, got %q", "kfm.air", info.AnimationFile)
	}
	if info.SoundFile != "kfm.snd" {
		t.Errorf("expected SoundFile %q, got %q", "kfm.snd", info.SoundFile)
	}
	if info.CommandFile != "kfm.cmd" {
		t.Errorf("expected CommandFile %q, got %q", "kfm.cmd", info.CommandFile)
	}
	if info.ConstantsFile != "kfm.cns" {
		t.Errorf("expected ConstantsFile %q, got %q", "kfm.cns", info.ConstantsFile)
	}

	wantStateFiles := []string{"kfm.st", "kfm_extra1.st", "kfm_extra2.st"}
	if len(info.StateFiles) != len(wantStateFiles) {
		t.Fatalf("expected %d StateFiles, got %d: %v", len(wantStateFiles), len(info.StateFiles), info.StateFiles)
	}
	for i, want := range wantStateFiles {
		if info.StateFiles[i] != want {
			t.Errorf("StateFiles[%d]: expected %q, got %q (file-appearance order expected)", i, want, info.StateFiles[i])
		}
	}

	// pal1/pal2/pal3 appear out of order in the source ("pal2, pal1, pal3");
	// CharacterInfo.Palettes is documented as being in palette number order.
	wantPalettes := []string{"kfm1.act", "kfm2.act", "kfm3.act"}
	if len(info.Palettes) != len(wantPalettes) {
		t.Fatalf("expected %d Palettes, got %d: %v", len(wantPalettes), len(info.Palettes), info.Palettes)
	}
	for i, want := range wantPalettes {
		if info.Palettes[i] != want {
			t.Errorf("Palettes[%d]: expected %q, got %q (palette number order expected)", i, want, info.Palettes[i])
		}
	}
}

func TestParse_UnrecognizedSection_IsSkippedWithoutAbortingKnownSections(t *testing.T) {
	src := `[Info]
name = "Kung Fu Man"

[Arcade]
1.storyboard = intro.def
easy.difficulty = 1,2

[Files]
sprite = kfm.sff
`

	info, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "Kung Fu Man" {
		t.Errorf("expected Name %q, got %q", "Kung Fu Man", info.Name)
	}
	if info.SpriteFile != "kfm.sff" {
		t.Errorf("expected SpriteFile %q parsed from the section after the unrecognized one, got %q", "kfm.sff", info.SpriteFile)
	}
}

func TestParse_CaseInsensitiveSectionsAndKeys_UnquotedAndQuotedValues(t *testing.T) {
	src := `[INFO]
NAME = Unquoted Name
Author = "Quoted Author"

[FILES]
Sprite = kfm.sff
`

	info, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "Unquoted Name" {
		t.Errorf("expected Name %q, got %q", "Unquoted Name", info.Name)
	}
	if info.Author != "Quoted Author" {
		t.Errorf("expected Author %q, got %q", "Quoted Author", info.Author)
	}
	if info.SpriteFile != "kfm.sff" {
		t.Errorf("expected SpriteFile %q, got %q", "kfm.sff", info.SpriteFile)
	}
}

func TestParse_CommentsAndBlankLines_AreIgnored(t *testing.T) {
	src := `; this is a full character definition
[Info] ; identifying information
name = "Kung Fu Man" ; display name

[Files]

sprite = kfm.sff ; sprite sheet
`

	info, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "Kung Fu Man" {
		t.Errorf("expected Name %q, got %q", "Kung Fu Man", info.Name)
	}
	if info.SpriteFile != "kfm.sff" {
		t.Errorf("expected SpriteFile %q, got %q", "kfm.sff", info.SpriteFile)
	}
}

func TestParse_EmptyInput_ReturnsZeroValueCharacterInfoWithoutError(t *testing.T) {
	info, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "" || info.Author != "" || info.SpriteFile != "" || info.AnimationFile != "" ||
		info.SoundFile != "" || info.CommandFile != "" || info.ConstantsFile != "" {
		t.Errorf("expected all string fields empty, got %+v", info)
	}
	if len(info.StateFiles) != 0 || len(info.Palettes) != 0 {
		t.Errorf("expected no StateFiles/Palettes, got %+v", info)
	}
}

func TestParse_MalformedKeyValueLine_ReturnsErrorNamingTheLine(t *testing.T) {
	src := `[Info]
name = "Kung Fu Man"
author World
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for the malformed line, got nil")
	}
	if !strings.Contains(err.Error(), "line 3") {
		t.Errorf("expected error to identify line 3, got: %v", err)
	}
	if !strings.Contains(err.Error(), "author World") {
		t.Errorf("expected error to quote the offending line, got: %v", err)
	}
}

func TestParse_MalformedSectionHeader_ReturnsErrorNamingTheLine(t *testing.T) {
	src := `[Info
name = "Kung Fu Man"
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for the malformed section header, got nil")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("expected error to identify line 1, got: %v", err)
	}
}

func TestParse_UnrecognizedKeyInKnownSection_IsIgnoredWithoutError(t *testing.T) {
	src := `[Info]
name = "Kung Fu Man"
displayname = "Kung Fu Man"

[Files]
sprite = kfm.sff
stcommon = common1.cns
`

	info, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "Kung Fu Man" {
		t.Errorf("expected Name %q, got %q", "Kung Fu Man", info.Name)
	}
	if info.SpriteFile != "kfm.sff" {
		t.Errorf("expected SpriteFile %q, got %q", "kfm.sff", info.SpriteFile)
	}
	if len(info.StateFiles) != 0 {
		t.Errorf("expected stcommon to be ignored (not an 'st'/'stN' key), got StateFiles %v", info.StateFiles)
	}
}

func TestParse_ReaderError_ReturnsWrappedError(t *testing.T) {
	_, err := Parse(&failingReader{})
	if err == nil {
		t.Fatal("expected an error from a failing reader, got nil")
	}
	var target *failingReaderError
	if !errors.As(err, &target) {
		t.Errorf("expected wrapped failingReaderError, got: %v", err)
	}
}

// failingReaderError is a sentinel error type used to verify Parse wraps
// (rather than discards) the underlying reader's error.
type failingReaderError struct{}

func (e *failingReaderError) Error() string { return "simulated read failure" }

// failingReader always fails on Read.
type failingReader struct{}

func (r *failingReader) Read(p []byte) (int, error) {
	return 0, &failingReaderError{}
}
