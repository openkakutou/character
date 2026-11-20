package character

import (
	"bytes"
	"fmt"

	"github.com/openkakutou/character/def"
)

// Movelist is one move list declared by a .def [Files] key ("movelist",
// "movelist0", "movelist1", ...), see def.MovelistRef for what the key and
// index mean.
//
// To pick the text for a selected index the way Ikemen GO does (src/char.go:
// 3425-3452 loadMovelists, src/script.go:3996-4014 getMovelist, commit
// 54d4148a728153c2097a0fb899a06812b2d3e193): take the first entry of that
// Index whose Error is empty, and when the index has none, the first such
// entry of index 0.
type Movelist struct {
	// Key is the .def key as written.
	Key string `json:"key"`
	// Index is the move list number the key stands for.
	Index int `json:"index"`
	// File is the path as written in the .def.
	File string `json:"file"`
	// Text is the file's content exactly as written, line endings and glyph
	// tokens included, only a leading UTF-8 byte order mark removed
	// (src/common.go:503-519 LoadText). Empty when Error is set.
	Text string `json:"text"`
	// Error is non-empty when the file could not be read. It is non-fatal:
	// the load still succeeds, same shape as CommandError.
	Error string `json:"error,omitempty"`
}

// utf8BOM is the byte order mark Ikemen GO strips from a move list file.
var utf8BOM = []byte{0xef, 0xbb, 0xbf}

// resolveMovelists reads the file of each declaration in refs, in order,
// through read. A read failure is reported in that entry's Error. The result
// is never nil.
func resolveMovelists(refs []def.MovelistRef, read func(file string) ([]byte, error)) []Movelist {
	movelists := []Movelist{}
	for _, ref := range refs {
		entry := Movelist{Key: ref.Key, Index: ref.Index, File: ref.File}
		data, err := read(ref.File)
		if err != nil {
			entry.Error = fmt.Sprintf("character: reading move list file %q: %v", ref.File, err)
		} else {
			entry.Text = string(bytes.TrimPrefix(data, utf8BOM))
		}
		movelists = append(movelists, entry)
	}
	return movelists
}
