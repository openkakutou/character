package cmd

import (
	"fmt"
	"io"
	"sort"

	"github.com/openkakutou/character/cns"
)

// Serialize writes file to w as MUGEN/Ikemen GO .cmd text: a "[Remap]"
// section (if file.Remap is non-empty), a "[Defaults]" section (if
// file.Defaults is not the zero value), one "[Command]" block per
// file.Commands entry in order, and finally the linked "always" state
// (file.States) via cns.Serialize — see
// .vibe/decisions/025-cmd-package-reuses-cns-for-state-triggering-block.md.
//
// This is a first-pass write path: it does not attempt a byte-exact
// round-trip of any original file's formatting, comments, or unrecognized
// sections (a separate, format-preserving concern — see Document) — it only
// guarantees valid, readable output that Parse reads back into an
// equivalent CommandFile. Remap is written sorted by key for deterministic
// output, since Go map order carries no meaning Parse relies on. A
// Command's Time/BufferTime is only written when non-zero: zero means "not
// set, use CommandFile.Defaults" (see Command's own doc comment), so
// writing an explicit zero would round-trip into a real override instead of
// staying unset.
func Serialize(w io.Writer, file CommandFile) error {
	if len(file.Remap) > 0 {
		if err := writeRemapSection(w, file.Remap); err != nil {
			return err
		}
	}
	if file.Defaults != (CommandDefaults{}) {
		if err := writeDefaultsSection(w, file.Defaults); err != nil {
			return err
		}
	}
	for _, c := range file.Commands {
		if err := writeCommandSection(w, c); err != nil {
			return err
		}
	}
	if len(file.States) > 0 {
		if err := cns.Serialize(w, file.States); err != nil {
			return fmt.Errorf("cmd: writing linked state: %w", err)
		}
	}
	return nil
}

func writeRemapSection(w io.Writer, remap map[string]string) error {
	if _, err := fmt.Fprintln(w, "[Remap]"); err != nil {
		return fmt.Errorf("cmd: writing [Remap] header: %w", err)
	}

	keys := make([]string, 0, len(remap))
	for key := range remap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, err := fmt.Fprintf(w, "%s = %s\n", key, remap[key]); err != nil {
			return fmt.Errorf("cmd: writing Remap %s: %w", key, err)
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("cmd: writing Remap section separator: %w", err)
	}
	return nil
}

func writeDefaultsSection(w io.Writer, d CommandDefaults) error {
	if _, err := fmt.Fprintln(w, "[Defaults]"); err != nil {
		return fmt.Errorf("cmd: writing [Defaults] header: %w", err)
	}
	if _, err := fmt.Fprintf(w, "command.time = %d\n", d.Time); err != nil {
		return fmt.Errorf("cmd: writing command.time: %w", err)
	}
	if _, err := fmt.Fprintf(w, "command.buffer.time = %d\n", d.BufferTime); err != nil {
		return fmt.Errorf("cmd: writing command.buffer.time: %w", err)
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("cmd: writing Defaults section separator: %w", err)
	}
	return nil
}

func writeCommandSection(w io.Writer, c Command) error {
	if _, err := fmt.Fprintln(w, "[Command]"); err != nil {
		return fmt.Errorf("cmd: writing [Command] header: %w", err)
	}
	if _, err := fmt.Fprintf(w, "name = %q\n", c.Name); err != nil {
		return fmt.Errorf("cmd: writing Command %q name: %w", c.Name, err)
	}
	if _, err := fmt.Fprintf(w, "command = %s\n", c.Input); err != nil {
		return fmt.Errorf("cmd: writing Command %q input: %w", c.Name, err)
	}
	if c.Time != 0 {
		if _, err := fmt.Fprintf(w, "time = %d\n", c.Time); err != nil {
			return fmt.Errorf("cmd: writing Command %q time: %w", c.Name, err)
		}
	}
	if c.BufferTime != 0 {
		if _, err := fmt.Fprintf(w, "buffer.time = %d\n", c.BufferTime); err != nil {
			return fmt.Errorf("cmd: writing Command %q buffer.time: %w", c.Name, err)
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("cmd: writing Command %q separator: %w", c.Name, err)
	}
	return nil
}
