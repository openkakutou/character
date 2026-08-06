package def

import (
	"fmt"
	"io"
)

// Serialize writes info to w as MUGEN/Ikemen GO .def text, emitting an
// "[Info]" section followed by a "[Files]" section.
//
// This is a first-pass write path: it does not attempt a byte-exact
// round-trip of any original file's formatting, comments, or unrecognized
// sections (a separate, format-preserving concern — see Document) — it only
// guarantees valid, readable output that Parse reads back into an
// equivalent CharacterInfo. Name and Author are always quoted, matching
// common .def convention; file references are written unquoted, matching
// how Parse's own test fixtures write them. A field left at its zero value
// (an empty string, or a nil/empty StateFiles/Palettes) is omitted from the
// output entirely rather than written as an empty key, keeping the output
// clean for a minimal CharacterInfo. StateFiles are written as "st",
// "st1", "st2", ... in slice order; Palettes are written as "pal1",
// "pal2", ... in slice order — Parse reconstructs both lists from
// file-appearance order (StateFiles) or numeric suffix (Palettes), so this
// numbering always round-trips regardless of the exact key chosen.
func Serialize(w io.Writer, info CharacterInfo) error {
	if err := writeInfoSection(w, info); err != nil {
		return err
	}
	if err := writeFilesSection(w, info); err != nil {
		return err
	}
	return nil
}

func writeInfoSection(w io.Writer, info CharacterInfo) error {
	if _, err := fmt.Fprintln(w, "[Info]"); err != nil {
		return fmt.Errorf("def: writing [Info] header: %w", err)
	}
	if info.Name != "" {
		if _, err := fmt.Fprintf(w, "name = %q\n", info.Name); err != nil {
			return fmt.Errorf("def: writing name: %w", err)
		}
	}
	if info.Author != "" {
		if _, err := fmt.Fprintf(w, "author = %q\n", info.Author); err != nil {
			return fmt.Errorf("def: writing author: %w", err)
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("def: writing section separator: %w", err)
	}
	return nil
}

func writeFilesSection(w io.Writer, info CharacterInfo) error {
	if _, err := fmt.Fprintln(w, "[Files]"); err != nil {
		return fmt.Errorf("def: writing [Files] header: %w", err)
	}

	fields := []struct {
		key   string
		value string
	}{
		{"sprite", info.SpriteFile},
		{"anim", info.AnimationFile},
		{"sound", info.SoundFile},
		{"cmd", info.CommandFile},
		{"cns", info.ConstantsFile},
	}
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		if _, err := fmt.Fprintf(w, "%s = %s\n", f.key, f.value); err != nil {
			return fmt.Errorf("def: writing %s: %w", f.key, err)
		}
	}

	for i, path := range info.StateFiles {
		key := "st"
		if i > 0 {
			key = fmt.Sprintf("st%d", i)
		}
		if _, err := fmt.Fprintf(w, "%s = %s\n", key, path); err != nil {
			return fmt.Errorf("def: writing %s: %w", key, err)
		}
	}

	for i, path := range info.Palettes {
		key := fmt.Sprintf("pal%d", i+1)
		if _, err := fmt.Fprintf(w, "%s = %s\n", key, path); err != nil {
			return fmt.Errorf("def: writing %s: %w", key, err)
		}
	}

	return nil
}
