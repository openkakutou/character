package cns

import (
	"bytes"
	"fmt"
	"io"
)

// Document is the write-path counterpart to Parse/Serialize: it exists so a
// .cns file can be round-tripped — parsed, then serialized back out —
// without losing the comments, block ordering, and unrecognized sections
// that the pure-data StateDef/Controller model deliberately does not carry
// (see .vibe/decisions/012-cns-parse-header-detection-strategy.md), mirroring
// air.Document/def.Document (see
// .vibe/decisions/003-air-round-trip-via-separate-document-type.md).
//
// Document.StateDefs is decoded the same way Parse's return value is, for
// convenient structured access to what was parsed — but Serialize does not
// read it back. As long as Document.StateDefs is left untouched,
// ParseDocument followed by Serialize reproduces the original source
// byte-for-byte, comments and all. Mutating StateDefs has no effect on
// Serialize's output: regenerating text from an edited StateDefs slice
// while still preserving unrelated comments/sections/ordering around the
// edit is a heavier per-line reconciliation this type does not attempt.
type Document struct {
	StateDefs []StateDef

	source []byte
}

// ParseDocument reads MUGEN/Ikemen GO .cns combat logic text from r,
// decoding it the same way Parse does while also retaining the exact
// source bytes needed for a faithful round trip through Serialize.
func ParseDocument(r io.Reader) (*Document, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cns: reading document source: %w", err)
	}

	states, err := Parse(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}

	return &Document{StateDefs: states, source: source}, nil
}

// Serialize writes the Document's retained source back out to w verbatim,
// reproducing the exact text ParseDocument read — including comments,
// block ordering, unrecognized sections, and original line endings.
func (d *Document) Serialize(w io.Writer) error {
	if _, err := w.Write(d.source); err != nil {
		return fmt.Errorf("cns: writing document: %w", err)
	}
	return nil
}
