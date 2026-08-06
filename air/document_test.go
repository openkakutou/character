package air

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestParseDocument_RealisticFixture_SerializeReproducesOriginalBytesExactly(t *testing.T) {
	original, err := os.ReadFile("testdata/sample.air")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	doc, err := ParseDocument(bytes.NewReader(original))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), original) {
		t.Errorf("round trip did not reproduce the original file byte-for-byte.\n--- original ---\n%s\n--- got ---\n%s", original, buf.String())
	}
}

func TestParseDocument_RealisticFixture_PreservesAllComments(t *testing.T) {
	original, err := os.ReadFile("testdata/sample.air")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	doc, err := ParseDocument(bytes.NewReader(original))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}
	out := buf.String()

	wantComments := []string{
		"; OpenKakutou sample animation set",
		"; Action 200: walk cycle",
		";walking",
		"; idle-ish first frame",
		"; Action 201: a single-frame stance",
	}
	for _, c := range wantComments {
		if !strings.Contains(out, c) {
			t.Errorf("expected round-tripped output to still contain comment %q, got:\n%s", c, out)
		}
	}
}

func TestParseDocument_RealisticFixture_LosesNoFramesClsnBoxesOrLoopstart(t *testing.T) {
	original, err := os.ReadFile("testdata/sample.air")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	doc, err := ParseDocument(bytes.NewReader(original))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Animations) != 2 {
		t.Fatalf("expected 2 animations, got %d", len(doc.Animations))
	}

	a200 := doc.Animations[0]
	if a200.Number != 200 {
		t.Errorf("expected action number 200, got %d", a200.Number)
	}
	if len(a200.Frames) != 4 {
		t.Fatalf("expected 4 frames in action 200, got %d", len(a200.Frames))
	}
	if a200.LoopStart != 3 {
		t.Errorf("expected LoopStart 3, got %d", a200.LoopStart)
	}

	wantClsn2Default := ClsnBox{Left: -6, Top: -92, Right: 6, Bottom: 0}
	if len(a200.Frames[0].Clsn2) != 1 || a200.Frames[0].Clsn2[0] != wantClsn2Default {
		t.Errorf("expected frame 0 to carry the Clsn2Default box, got %+v", a200.Frames[0].Clsn2)
	}
	if len(a200.Frames[3].Clsn2) != 1 || a200.Frames[3].Clsn2[0] != wantClsn2Default {
		t.Errorf("expected frame 3 to still carry the Clsn2Default box, got %+v", a200.Frames[3].Clsn2)
	}

	wantClsn1Override := ClsnBox{Left: -3, Top: -73, Right: 15, Bottom: -1}
	if len(a200.Frames[2].Clsn1) != 1 || a200.Frames[2].Clsn1[0] != wantClsn1Override {
		t.Errorf("expected frame 2 to carry the Clsn1 override, got %+v", a200.Frames[2].Clsn1)
	}
	if len(a200.Frames[0].Clsn1) != 0 || len(a200.Frames[3].Clsn1) != 0 {
		t.Errorf("expected the one-shot Clsn1 override not to leak to other frames")
	}

	a201 := doc.Animations[1]
	if a201.Number != 201 {
		t.Errorf("expected action number 201, got %d", a201.Number)
	}
	if len(a201.Frames) != 1 {
		t.Fatalf("expected 1 frame in action 201, got %d", len(a201.Frames))
	}
	f := a201.Frames[0]
	if f.Time != 10 || f.Flip != FlipHV || f.Blend != BlendMode("AS") {
		t.Errorf("expected time=10 Flip=HV Blend=AS, got Time=%d Flip=%q Blend=%q", f.Time, f.Flip, f.Blend)
	}
}

func TestParseDocument_FileWithoutTrailingNewline_RoundTripsExactly(t *testing.T) {
	src := "[Begin Action 0]\n0,0, 0,0, 1 ; last line, no newline after it"

	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != src {
		t.Errorf("expected exact round trip, want %q, got %q", src, buf.String())
	}
	if len(doc.Animations) != 1 || len(doc.Animations[0].Frames) != 1 {
		t.Errorf("expected 1 animation with 1 frame, got %+v", doc.Animations)
	}
}

func TestParseDocument_CommentAndBlankOnlyFile_RoundTripsWithNoAnimations(t *testing.T) {
	src := "; nothing to see here\n\n; still nothing\n"

	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Animations) != 0 {
		t.Errorf("expected no animations, got %d", len(doc.Animations))
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != src {
		t.Errorf("expected exact round trip, want %q, got %q", src, buf.String())
	}
}

func TestParseDocument_CRLFLineEndings_RoundTripExactlyAndDecodeCorrectly(t *testing.T) {
	src := "[Begin Action 7]\r\n0,0, 0,0, 1\r\n0,1, 0,0, 2\r\n"

	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Animations) != 1 || doc.Animations[0].Number != 7 {
		t.Fatalf("expected 1 animation numbered 7, got %+v", doc.Animations)
	}
	if len(doc.Animations[0].Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(doc.Animations[0].Frames))
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != src {
		t.Errorf("expected exact round trip preserving CRLF endings, want %q, got %q", src, buf.String())
	}
}

func TestParseDocument_ReaderFailure_ReturnsErrorNotPanic(t *testing.T) {
	_, err := ParseDocument(errReader{})
	if err == nil {
		t.Fatal("expected an error when the underlying reader fails, got nil")
	}
}

func TestParseDocument_MalformedFrameLine_ReturnsError(t *testing.T) {
	src := "[Begin Action 0]\nabc,0, 0,0, 5\n"

	_, err := ParseDocument(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a malformed frame line, got nil")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected the error to name line 2, got: %v", err)
	}
}

func TestDocumentSerialize_ReturnsError_OnWriterFailure(t *testing.T) {
	doc, err := ParseDocument(strings.NewReader("[Begin Action 0]\n0,0, 0,0, 1\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = doc.Serialize(failingWriter{})
	if err == nil {
		t.Fatal("expected an error when the writer fails, got nil")
	}
}

func TestDocumentSerialize_MutatingAnimationsDoesNotAffectOutput(t *testing.T) {
	// Document guarantees a faithful round trip for *unmodified* content
	// (see .vibe/decisions/003-air-round-trip-via-separate-document-type.md).
	// It does not yet regenerate output from edited Animations, so a
	// mutation must not silently produce different bytes than the
	// retained source.
	src := "[Begin Action 0]\n0,0, 0,0, 1\n"

	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc.Animations[0].Number = 999
	doc.Animations[0].Frames[0].Time = 42

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != src {
		t.Errorf("expected Serialize to still emit the original retained source, want %q, got %q", src, buf.String())
	}
}
