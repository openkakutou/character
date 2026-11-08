package zss

import (
	"bytes"
	"fmt"
	"io"
)

// Document is the write-path counterpart to Parse/Serialize: it exists so a
// .zss file can be round-tripped — parsed, then serialized back out —
// without losing any formatting Parse's Script deliberately does not carry
// (multi-line-wrapped headers, exact spacing), mirroring
// air.Document/def.Document/cns.Document (see
// .vibe/decisions/003-air-round-trip-via-separate-document-type.md).
//
// Document.Script is decoded the same way Parse's return value is, for
// convenient structured access to what was parsed — but Serialize does not
// read it back. As long as Document.Script is left untouched, ParseDocument
// followed by Serialize reproduces the original source byte-for-byte.
// Mutating Script has no effect on Serialize's output; to write an edited
// Script while keeping every untouched comment, blank line and wrapped
// header byte-exact, use SerializeEdited (rules documented there).
type Document struct {
	Script Script

	source   []byte
	baseline Script      // independent copy of the parsed Script, for diffing
	spans    []blockSpan // per-block line ranges in source, parallel to baseline.Blocks
}

// ParseDocument reads Ikemen GO .zss text from r, decoding it the same way
// Parse does while also retaining the exact source bytes needed for a
// faithful round trip through Serialize.
func ParseDocument(r io.Reader) (*Document, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("zss: reading document source: %w", err)
	}

	script, err := Parse(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}

	baseline, err := Parse(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}

	return &Document{Script: script, source: source, baseline: baseline, spans: indexBlocks(splitLines(source))}, nil
}

// Serialize writes the Document's retained source back out to w verbatim,
// reproducing the exact text ParseDocument read.
func (d *Document) Serialize(w io.Writer) error {
	if _, err := w.Write(d.source); err != nil {
		return fmt.Errorf("zss: writing document: %w", err)
	}
	return nil
}
