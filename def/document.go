package def

import (
	"bytes"
	"fmt"
	"io"
)

// Document is the write-path counterpart to Parse/Serialize: it exists so a
// .def file can be round-tripped — parsed, then serialized back out —
// without losing the comments, section ordering, and unrecognized sections
// that the pure-data CharacterInfo model deliberately does not carry (see
// .vibe/decisions/009-def-parse-ignores-unknown-sections.md).
//
// Document.Info is decoded the same way Parse's return value is, for
// convenient structured access to what was parsed. Serialize ignores it and
// writes the retained source verbatim, so ParseDocument followed by Serialize
// reproduces the original source byte-for-byte, comments and all.
// SerializeEdited is the format-preserving write path for an edited
// CharacterInfo: it rewrites only what changed (see
// .vibe/decisions/033-format-preserving-edits-per-format-alignment.md).
type Document struct {
	Info CharacterInfo

	source []byte

	// layout and baseline are what SerializeEdited needs: where each
	// recognized key sits in source, and the CharacterInfo as parsed (kept
	// apart from the caller-visible Info so mutating Info cannot hide an edit).
	layout   sourceLayout
	baseline CharacterInfo
}

// ParseDocument reads MUGEN/Ikemen GO .def character definition text from
// r, decoding it the same way Parse does while also retaining the exact
// source bytes needed for a faithful round trip through Serialize.
func ParseDocument(r io.Reader) (*Document, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("def: reading document source: %w", err)
	}

	info, err := Parse(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}

	return &Document{
		Info:     info,
		source:   source,
		layout:   indexSource(source),
		baseline: cloneInfo(info),
	}, nil
}

// Serialize writes the Document's retained source back out to w verbatim,
// reproducing the exact text ParseDocument read — including comments,
// section ordering, unrecognized sections, and original line endings.
func (d *Document) Serialize(w io.Writer) error {
	if _, err := w.Write(d.source); err != nil {
		return fmt.Errorf("def: writing document: %w", err)
	}
	return nil
}
