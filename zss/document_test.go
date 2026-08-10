package zss

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParseDocument_RealisticFixture_SerializeReproducesOriginalBytesExactly(t *testing.T) {
	original, err := os.ReadFile("testdata/sample.zss")
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

func TestParseDocument_RealisticFixture_ExposesDecodedScript(t *testing.T) {
	original, err := os.ReadFile("testdata/sample.zss")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	doc, err := ParseDocument(bytes.NewReader(original))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Script.Blocks) != 4 {
		t.Fatalf("expected 4 decoded blocks, got %d", len(doc.Script.Blocks))
	}
	if doc.Script.Blocks[0].Kind != BlockKindStatedef || doc.Script.Blocks[0].Number != 200 {
		t.Errorf("expected first decoded block to be Statedef 200, got %+v", doc.Script.Blocks[0])
	}
	if !strings.Contains(doc.Script.Preamble, "Sample .zss fixture") {
		t.Errorf("expected decoded Preamble to contain the fixture's banner comment, got %q", doc.Script.Preamble)
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

// ParseDocument delegates to Parse (see its own tests for the malformed
// header behavior); this only needs to confirm a genuinely malformed source
// still surfaces as an error here too.
func TestParseDocument_MalformedSource_ReturnsError(t *testing.T) {
	src := `[NotAKnownHeader something]
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

func TestDocument_MutatingScriptAfterParse_DoesNotAffectSerializeOutput(t *testing.T) {
	src := `[Statedef 0]
noop{}
`
	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc.Script.Blocks[0].Number = 999
	doc.Script.Blocks = nil

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}
	if buf.String() != src {
		t.Errorf("expected Serialize to ignore mutated Script and reproduce original source.\n--- want ---\n%s\n--- got ---\n%s", src, buf.String())
	}
}
