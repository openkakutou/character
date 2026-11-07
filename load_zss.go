package character

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/openkakutou/character/def"
	"github.com/openkakutou/character/zss"
)

// ScriptFile is one .zss state script referenced from a .def [Files]
// section (st, stN or stcommon), see .vibe/decisions/032.
type ScriptFile struct {
	// File is the name the script was actually loaded under: the .def entry
	// as written, or "<entry>.zss" when the entry was not found and Ikemen
	// GO's fallback applied.
	File string `json:"file"`
	// Script is the parsed script, the normalized empty value when Error is
	// set.
	Script zss.Script `json:"script"`
	// Error is non-empty when the .zss could not be read or parsed. It is
	// non-fatal: the load still succeeds, same shape as CommandError
	// (.vibe/decisions/031).
	Error string `json:"error,omitempty"`
}

// resolveScripts applies Ikemen GO's .zss file selection to info: each
// StateFiles entry in order, then StateCommonFile. read returns the content
// of an entry, and must report a missing file with an error matching
// fs.ErrNotExist (only that triggers the fallback below).
//
//   - An entry whose extension is ".zss" (case-insensitive, exact: Ikemen
//     GO's check is an unanchored regex, so ".zssx" would match there) is
//     loaded; a read or parse failure is reported in ScriptFile.Error.
//   - Any other entry is read only to test that it exists (.cns-format
//     state files are out of scope, their content is discarded). Only when
//     it is not found is "<entry>.zss" tried, and loaded if present
//     (File = "<entry>.zss"); otherwise the entry is skipped silently, e.g.
//     the engine's common1.cns, which is not in the character folder.
//
// The result is never nil.
func resolveScripts(info def.CharacterInfo, read func(entry string) ([]byte, error)) []ScriptFile {
	scripts := []ScriptFile{}
	entries := info.StateFiles
	if info.StateCommonFile != "" {
		entries = append(append([]string{}, entries...), info.StateCommonFile)
	}
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		if isZssName(entry) {
			data, err := read(entry)
			scripts = append(scripts, parseScript(entry, data, err))
			continue
		}
		// shortcut: the entry is fully read just to test existence, upgrade
		// to a stat-style probe if .cns state files get large.
		if _, err := read(entry); !errors.Is(err, fs.ErrNotExist) {
			continue // exists (a .cns, out of scope) or unreadable: not ours
		}
		fallback := entry + ".zss"
		data, err := read(fallback)
		if !errors.Is(err, fs.ErrNotExist) {
			scripts = append(scripts, parseScript(fallback, data, err))
		}
	}
	return scripts
}

// isZssName reports whether name has the extension ".zss", ignoring case.
func isZssName(name string) bool {
	return strings.EqualFold(path.Ext(strings.ReplaceAll(name, "\\", "/")), ".zss")
}

// parseScript parses the already-read content of one .zss, turning a read
// error (readErr) or a parse failure into the entry's Error.
func parseScript(file string, data []byte, readErr error) ScriptFile {
	empty := ScriptFile{File: file, Script: normalizeScriptForJSON(zss.Script{})}
	if readErr != nil {
		empty.Error = fmt.Sprintf("character: opening state script file %q: %v", file, readErr)
		return empty
	}
	script, err := zss.Parse(bytes.NewReader(data))
	if err != nil {
		empty.Error = fmt.Sprintf("character: parsing state script file %q: %v", file, err)
		return empty
	}
	return ScriptFile{File: file, Script: normalizeScriptForJSON(script)}
}
