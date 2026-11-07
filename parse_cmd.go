package character

import (
	"bytes"
	"fmt"

	"github.com/openkakutou/character/cmd"
)

// ParseCmd parses .cmd file content into a CommandFile, the read-path
// counterpart to SerializeCmd — used by LoadBytes to fill
// Character.Commands and by cmd/wasm's standalone loadCmd entrypoint (see
// .vibe/decisions/031).
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
