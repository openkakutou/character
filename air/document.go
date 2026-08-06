package air

import (
	"bytes"
	"fmt"
	"io"
)

// Document is the write-path counterpart to Parse/Serialize: it exists so a
// .air file can be round-tripped — parsed, then serialized back out —
// without losing the comments, blank lines, and exact formatting that the
// pure-data Animation/Frame model deliberately does not carry (see
// .vibe/decisions/002-air-comments-stripped-not-preserved-by-parse.md).
//
// Document.Animations is decoded the same way Parse's return value is, for
// convenient structured access to what was parsed — but Serialize does not
// read it back. As long as Document.Animations is left untouched,
// ParseDocument followed by Serialize reproduces the original source
// byte-for-byte, comments included. Mutating Animations has no effect on
// Serialize's output: regenerating text from an edited Animations slice
// while still preserving unrelated comments/ordering around the edit is a
// heavier per-line reconciliation this type does not attempt (see
// .vibe/decisions/003-air-round-trip-via-separate-document-type.md).
type Document struct {
	Animations []Animation

	source []byte
}

// ParseDocument reads MUGEN/Ikemen GO .air animation text from r, decoding
// it the same way Parse does while also retaining the exact source bytes
// needed for a faithful round trip through Serialize.
func ParseDocument(r io.Reader) (*Document, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("air: reading document source: %w", err)
	}

	animations, err := Parse(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}

	return &Document{Animations: animations, source: source}, nil
}

// Serialize writes the Document's retained source back out to w verbatim,
// reproducing the exact text ParseDocument read — including comments,
// blank lines, and original line endings.
func (d *Document) Serialize(w io.Writer) error {
	if _, err := w.Write(d.source); err != nil {
		return fmt.Errorf("air: writing document: %w", err)
	}
	return nil
}
