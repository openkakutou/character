package zss

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Serialize writes script to w as Ikemen GO .zss text, emitting the
// Preamble verbatim followed by one block header per Block in order — a
// "[Statedef N; ...]" line for BlockKindStatedef, or a
// "[Function Name(...) ...]" line for BlockKindFunction — then that Block's
// Body verbatim.
//
// This is a first-pass write path: it does not attempt a byte-exact
// round-trip of any original file's formatting (a separate, format-
// preserving concern — see Document) — it only guarantees valid output that
// Parse reads back into an equivalent Script, mirroring cns.Serialize. A
// Statedef's HeaderParams (an unordered map) are written sorted by key for
// deterministic, reviewable output; map order carries no meaning Parse
// relies on. A Body already carries its own trailing newline (see Parse),
// so nothing is added after it.
func Serialize(w io.Writer, script Script) error {
	if _, err := io.WriteString(w, script.Preamble); err != nil {
		return fmt.Errorf("zss: writing preamble: %w", err)
	}

	for i, b := range script.Blocks {
		if err := writeBlockHeader(w, b); err != nil {
			return fmt.Errorf("zss: writing block %d header: %w", i, err)
		}
		if _, err := io.WriteString(w, b.Body); err != nil {
			return fmt.Errorf("zss: writing block %d body: %w", i, err)
		}
	}
	return nil
}

// writeBlockHeader writes a single block's header line for b, dispatching
// on Kind.
func writeBlockHeader(w io.Writer, b Block) error {
	switch b.Kind {
	case BlockKindFunction:
		return writeFunctionHeader(w, b)
	default:
		// BlockKindStatedef, and any future/unrecognized kind: written as a
		// Statedef header, since Number is the only identifying field a
		// caller could otherwise have set.
		return writeStatedefHeader(w, b)
	}
}

// writeStatedefHeader writes "[Statedef N]" (no HeaderParams) or
// "[Statedef N; key1: value1; key2: value2]" (HeaderParams sorted by key).
func writeStatedefHeader(w io.Writer, b Block) error {
	if len(b.HeaderParams) == 0 {
		_, err := fmt.Fprintf(w, "[Statedef %d]\n", b.Number)
		return err
	}

	keys := make([]string, 0, len(b.HeaderParams))
	for k := range b.HeaderParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, len(keys))
	for i, k := range keys {
		pairs[i] = fmt.Sprintf("%s: %s", k, b.HeaderParams[k])
	}

	_, err := fmt.Fprintf(w, "[Statedef %d; %s]\n", b.Number, strings.Join(pairs, "; "))
	return err
}

// writeFunctionHeader writes "[Function Name(p1, p2) r1, r2]", omitting the
// return-variable segment entirely when Ret is empty.
func writeFunctionHeader(w io.Writer, b Block) error {
	if len(b.Ret) == 0 {
		_, err := fmt.Fprintf(w, "[Function %s(%s)]\n", b.Name, strings.Join(b.Params, ", "))
		return err
	}
	_, err := fmt.Fprintf(w, "[Function %s(%s) %s]\n", b.Name, strings.Join(b.Params, ", "), strings.Join(b.Ret, ", "))
	return err
}
