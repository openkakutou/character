package def

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParseDocument_RealisticFixture_SerializeReproducesSourceByteForByte(t *testing.T) {
	src := `; Kung Fu Man character definition
[Info] ; identifying information
name = "Kung Fu Man"		; display name
author = "Elecbyte"

[Arcade]
1.storyboard = intro.def

[Files]
cmd = kfm.cmd
cns = kfm.cns
sprite = kfm.sff
anim   = kfm.air
sound = kfm.snd
st = kfm.st
st1 = kfm_extra1.st
pal2 = kfm2.act
pal1 = kfm1.act
`

	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}

	if buf.String() != src {
		t.Errorf("expected byte-for-byte reproduction of source.\n--- want ---\n%s\n--- got ---\n%s", src, buf.String())
	}
}

func TestParseDocument_RealisticFixture_ExposesDecodedCharacterInfo(t *testing.T) {
	src := `[Info]
name = "Kung Fu Man"
author = "Elecbyte"

[Files]
sprite = kfm.sff
anim = kfm.air
`

	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Info.Name != "Kung Fu Man" {
		t.Errorf("expected decoded Info.Name %q, got %q", "Kung Fu Man", doc.Info.Name)
	}
	if doc.Info.Author != "Elecbyte" {
		t.Errorf("expected decoded Info.Author %q, got %q", "Elecbyte", doc.Info.Author)
	}
	if doc.Info.SpriteFile != "kfm.sff" {
		t.Errorf("expected decoded Info.SpriteFile %q, got %q", "kfm.sff", doc.Info.SpriteFile)
	}
	if doc.Info.AnimationFile != "kfm.air" {
		t.Errorf("expected decoded Info.AnimationFile %q, got %q", "kfm.air", doc.Info.AnimationFile)
	}
}

func TestParseDocument_EmptyInput_SerializeReproducesEmptySource(t *testing.T) {
	doc, err := ParseDocument(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestParseDocument_MalformedSource_ReturnsError(t *testing.T) {
	src := `[Info
name = "Kung Fu Man"
`
	_, err := ParseDocument(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for malformed source, got nil")
	}
}

func TestParseDocument_ReaderError_ReturnsWrappedError(t *testing.T) {
	_, err := ParseDocument(&failingReader{})
	if err == nil {
		t.Fatal("expected an error from a failing reader, got nil")
	}
	var target *failingReaderError
	if !errors.As(err, &target) {
		t.Errorf("expected wrapped failingReaderError, got: %v", err)
	}
}

func TestDocument_MutatingInfoAfterParse_DoesNotAffectSerializeOutput(t *testing.T) {
	src := `[Info]
name = "Kung Fu Man"

[Files]
sprite = kfm.sff
`
	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc.Info.Name = "Someone Else"

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}
	if buf.String() != src {
		t.Errorf("expected Serialize to ignore mutated Info and reproduce original source.\n--- want ---\n%s\n--- got ---\n%s", src, buf.String())
	}
}
