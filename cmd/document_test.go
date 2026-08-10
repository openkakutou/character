package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParseDocument_RealisticFixture_SerializeReproducesSourceByteForByte(t *testing.T) {
	src := `; Sample .cmd fixture
[Remap] ; button remapping
a = a
b = b

[Defaults]
command.time = 15
command.buffer.time   = 1

[Command]
name = "a"
command = a
time = 1

[Statedef -1]

[State -1, QCF Special]
type = ChangeState
value = 1000
trigger1 = command = "QCF_a"
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

func TestParseDocument_RealisticFixture_ExposesDecodedCommandFile(t *testing.T) {
	src := `[Command]
name = "a"
command = a
`

	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.File.Commands) != 1 {
		t.Fatalf("expected 1 decoded Command, got %d", len(doc.File.Commands))
	}
	if doc.File.Commands[0].Name != "a" || doc.File.Commands[0].Input != "a" {
		t.Errorf("expected decoded Command {Name: a, Input: a}, got %+v", doc.File.Commands[0])
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
	src := `[Command
name = "a"
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

func TestDocument_MutatingFileAfterParse_DoesNotAffectSerializeOutput(t *testing.T) {
	src := `[Command]
name = "a"
command = a
`
	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc.File.Commands[0].Name = "mutated"

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}
	if buf.String() != src {
		t.Errorf("expected Serialize to ignore mutated File and reproduce original source.\n--- want ---\n%s\n--- got ---\n%s", src, buf.String())
	}
}

// failingReaderError is a sentinel error type used to verify ParseDocument
// wraps (rather than discards) the underlying reader's error.
type failingReaderError struct{}

func (e *failingReaderError) Error() string { return "simulated read failure" }

// failingReader always fails on Read.
type failingReader struct{}

func (r *failingReader) Read(p []byte) (int, error) {
	return 0, &failingReaderError{}
}
