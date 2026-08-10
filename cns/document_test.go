package cns

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParseDocument_RealisticFixture_SerializeReproducesOriginalBytesExactly(t *testing.T) {
	original, err := os.ReadFile("testdata/sample.cns")
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

func TestParseDocument_RealisticFixture_PreservesCommentsAndUnrelatedSections(t *testing.T) {
	original, err := os.ReadFile("testdata/sample.cns")
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

	wantSubstrings := []string{
		"; Sample .cns fixture used by cns.Parse tests.",
		"; An unrelated section that appears before the first Statedef; its content",
		"[Data]",
		"life = 1000",
		"[Statedef 0, Standing]",
		"[State 0, ChangeState]",
	}
	for _, s := range wantSubstrings {
		if !strings.Contains(out, s) {
			t.Errorf("expected round-tripped output to still contain %q, got:\n%s", s, out)
		}
	}
}

func TestParseDocument_RealisticFixture_ExposesDecodedStateDefs(t *testing.T) {
	original, err := os.ReadFile("testdata/sample.cns")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	doc, err := ParseDocument(bytes.NewReader(original))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.StateDefs) != 3 {
		t.Fatalf("expected 3 decoded StateDefs, got %d", len(doc.StateDefs))
	}
	if doc.StateDefs[0].Number != 0 {
		t.Errorf("expected first decoded StateDef Number 0, got %d", doc.StateDefs[0].Number)
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

// ParseDocument delegates to Parse (see its own tests for the recovery
// behavior on a Statedef/State header missing its closing "]" — backlog
// item 042); this only needs to confirm genuinely malformed source, unrelated
// to either header shape, still surfaces as an error here too.
func TestParseDocument_MalformedSource_ReturnsError(t *testing.T) {
	src := `[Clsn1Default
Clsn1: 1
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

func TestDocument_MutatingStateDefsAfterParse_DoesNotAffectSerializeOutput(t *testing.T) {
	src := `[Statedef 0]
type = S

[State 0]
type = VelSet
x = 0
`
	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc.StateDefs[0].Type = StateTypeAir
	doc.StateDefs[0].Controllers = nil

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}
	if buf.String() != src {
		t.Errorf("expected Serialize to ignore mutated StateDefs and reproduce original source.\n--- want ---\n%s\n--- got ---\n%s", src, buf.String())
	}
}
