package character

import (
	"bytes"
	"fmt"

	"github.com/openkakutou/character/cmd"
)

// ParseCmd parses .cmd file content into a CommandFile, the read-path
// counterpart to SerializeCmd — used by cmd/wasm's loadCmd entrypoint,
// since .cmd isn't wired into Character/LoadBytes the way .def/.air/.cns
// are (a character uses .cmd independently of the rest of its files, and
// .cmd's own CommandFile shape has no place on Character yet — see
// .vibe/decisions/028's rejected-alternatives section for the same
// reasoning already applied to the write path).
//
// The returned CommandFile is JSON-marshal-ready for the WASM/JS boundary:
// every slice/map reachable from it (Remap, Commands, States) is guaranteed
// non-nil, the same normalizeCommandFileForJSON guarantee SerializeCmd
// already applies on the write path and LoadBytes applies for Character.
//
// A malformed or truncated cmdBytes returns a descriptive error from
// cmd.Parse rather than panicking.
func ParseCmd(cmdBytes []byte) (cmd.CommandFile, error) {
	file, err := cmd.Parse(bytes.NewReader(cmdBytes))
	if err != nil {
		return cmd.CommandFile{}, fmt.Errorf("character: parsing command bytes: %w", err)
	}
	return normalizeCommandFileForJSON(file), nil
}
