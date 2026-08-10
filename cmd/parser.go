// Parse turns MUGEN/Ikemen GO .cmd text into a CommandFile — the read-path
// entry point for this package.
package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/openkakutou/character/cns"
)

// commandSectionAttemptPattern recognizes any bracket line that starts with
// one of this package's own three known section keywords ("remap",
// "defaults", "command"), whether or not it goes on to be a validly closed
// header — used to tell a malformed attempt at one of them apart from a
// genuinely unrelated section (including a "[Statedef ...]"/"[State ...]"
// block, which belongs to the cns package instead). Mirrors
// cns.statedefAttemptPattern/stateAttemptPattern (see
// .vibe/decisions/012-cns-parse-header-detection-strategy.md in this
// repo's history for the pattern this follows).
var commandSectionAttemptPattern = regexp.MustCompile(`(?i)^\[\s*(remap|defaults|command)(\s|\])`)

// statedefHeaderAttemptPattern and stateHeaderAttemptPattern recognize an
// attempted "[Statedef ...]"/"[State ...]" header the same way cns
// package's own unexported statedefAttemptPattern/stateAttemptPattern do
// (matching even without a closing "]", so a real-world missing-bracket
// typo is still recognized) — duplicated here rather than exported from
// cns, since this is the minimum needed to detect whether a .cmd file
// declares its own "[Statedef -1]" header at all; see
// ensureImplicitStatedef.
var statedefHeaderAttemptPattern = regexp.MustCompile(`(?i)^\[\s*statedef(\s|\])`)
var stateHeaderAttemptPattern = regexp.MustCompile(`(?i)^\[\s*state(\s|\])`)

// Parse reads MUGEN/Ikemen GO .cmd input-command text from r and returns the
// CommandFile it describes.
//
// The "[Remap]", "[Defaults]", and "[Command]" sections (matched
// case-insensitively) populate CommandFile.Remap/Defaults/Commands; any
// other section, including "[Statedef -1]" and its "[State ...]"
// controllers, is skipped by this pass without validation — those are
// parsed separately via cns.Parse (run against the same source) into
// CommandFile.States, since they share .cns's syntax byte-for-byte. See the
// package doc comment and
// .vibe/decisions/025-cmd-package-reuses-cns-for-state-triggering-block.md.
//
// Within "[Defaults]"/"[Command]", "command.time"/"time" and
// "command.buffer.time"/"buffer.time" (matched case-insensitively, with or
// without the "command." prefix) set the numeric Time/BufferTime fields; a
// value that doesn't parse as a literal integer is ignored, leaving the
// field at zero, since this package has no expression grammar for these
// fields (unlike cns.StateDef's HeaderExprs escape hatch — real .cmd files
// don't appear to need one for these two fields). "name"/"command" set a
// Command's Name/Input; Input is stored verbatim and unevaluated, including
// any MUGEN/Ikemen input-sequence modifiers ("~", "$", "/", "+"). Every
// "[Remap]" key/value pair becomes a lowercase-keyed Remap entry.
//
// A bracket line missing its closing "]" that looks like an attempt at this
// package's own "[Remap]"/"[Defaults]"/"[Command]" header returns a
// descriptive, line-numbered error; any other bracket line missing "]"
// (including a "[State ...]" header — a real-world .cmd authoring typo,
// see backlog item 042's .cns equivalent) is left for cns.Parse's own
// recovery and is not treated as an error here. A content line inside a
// known section that isn't a valid "key=value" pair is ignored rather than
// erroring, the same tolerance def.Parse/cns.Parse already apply. Comment
// lines (';', whole-line or trailing) are ignored. An empty input returns a
// zero-value CommandFile and a nil error.
func Parse(r io.Reader) (CommandFile, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return CommandFile{}, fmt.Errorf("cmd: reading command source: %w", err)
	}

	file, err := parseSections(bytes.NewReader(source))
	if err != nil {
		return CommandFile{}, err
	}

	states, err := cns.Parse(bytes.NewReader(ensureImplicitStatedef(source)))
	if err != nil {
		return CommandFile{}, fmt.Errorf("cmd: parsing linked state: %w", err)
	}
	file.States = states

	return file, nil
}

// ensureImplicitStatedef prepends a synthetic "[Statedef -1]" header before
// a .cmd file's "[State ...]" controllers when the file never declares one
// explicitly. Real MUGEN/Ikemen .cmd files sometimes omit it entirely,
// since -1 is the only Statedef number a .cmd "always" section can ever
// use — found via a real-character corpus scan (Marvel's "Jean Grey" and
// "Nova"). cns.Parse alone has no such implicit-numbering convention (a
// bare .cns file always requires an explicit Statedef header before its
// State controllers), so this compensates before delegating to it. A file
// that already declares "[Statedef ...]" anywhere before its first
// "[State ...]" header is returned unmodified.
func ensureImplicitStatedef(source []byte) []byte {
	scanner := bufio.NewScanner(bytes.NewReader(source))

	injectBeforeLine := 0
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := stripComment(scanner.Text())
		if line == "" {
			continue
		}
		if statedefHeaderAttemptPattern.MatchString(line) {
			return source
		}
		if stateHeaderAttemptPattern.MatchString(line) {
			injectBeforeLine = lineNumber
			break
		}
	}

	if injectBeforeLine == 0 {
		return source
	}

	lines := strings.SplitAfter(string(source), "\n")
	var out strings.Builder
	for i, l := range lines {
		if i+1 == injectBeforeLine {
			out.WriteString("[Statedef -1]\n")
		}
		out.WriteString(l)
	}
	return []byte(out.String())
}

// parseSections reads the "[Remap]"/"[Defaults]"/"[Command]" sections of a
// .cmd file's text, ignoring every other section (left for cns.Parse).
func parseSections(r io.Reader) (CommandFile, error) {
	scanner := bufio.NewScanner(r)

	var file CommandFile
	var currentSection string
	var currentCommand *Command

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := stripComment(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				if commandSectionAttemptPattern.MatchString(line) {
					return CommandFile{}, fmt.Errorf("cmd: line %d: malformed section header %q", lineNumber, line)
				}
				// Not an attempt at one of this package's own headers
				// (e.g. a "[State ...]" header missing its closing "]") —
				// left for cns.Parse to recover or error on separately.
				currentSection = ""
				currentCommand = nil
				continue
			}

			switch strings.ToLower(strings.TrimSpace(line[1 : len(line)-1])) {
			case "remap":
				currentSection = "remap"
				currentCommand = nil
			case "defaults":
				currentSection = "defaults"
				currentCommand = nil
			case "command":
				currentSection = "command"
				file.Commands = append(file.Commands, Command{})
				currentCommand = &file.Commands[len(file.Commands)-1]
			default:
				// A "[Statedef ...]"/"[State ...]" block (handled by
				// cns.Parse) or any other unrecognized section: its
				// content has no place in this pass.
				currentSection = ""
				currentCommand = nil
			}
			continue
		}

		key, value, ok := parseKeyValueLine(line)
		if !ok {
			// A content line inside a known section that isn't a valid
			// key=value pair is ignored, the same tolerance def.Parse/
			// cns.Parse already apply elsewhere in this repo.
			continue
		}

		switch currentSection {
		case "remap":
			if file.Remap == nil {
				file.Remap = make(map[string]string)
			}
			file.Remap[strings.ToLower(key)] = value
		case "defaults":
			applyDefaultsField(&file.Defaults, key, value)
		case "command":
			applyCommandField(currentCommand, key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return CommandFile{}, fmt.Errorf("cmd: reading command source: %w", err)
	}

	return file, nil
}

// applyDefaultsField applies a single key=value line to a CommandDefaults
// being built. An unrecognized key is ignored, as is a recognized key whose
// value doesn't parse as a literal integer.
func applyDefaultsField(d *CommandDefaults, key, value string) {
	switch normalizeCommandKey(key) {
	case "command.time":
		if n, err := strconv.Atoi(value); err == nil {
			d.Time = n
		}
	case "command.buffer.time":
		if n, err := strconv.Atoi(value); err == nil {
			d.BufferTime = n
		}
	}
}

// applyCommandField applies a single key=value line to a Command being
// built. An unrecognized key is ignored, as is a recognized numeric key
// whose value doesn't parse as a literal integer.
func applyCommandField(c *Command, key, value string) {
	switch normalizeCommandKey(key) {
	case "name":
		c.Name = value
	case "command":
		c.Input = value
	case "time", "command.time":
		if n, err := strconv.Atoi(value); err == nil {
			c.Time = n
		}
	case "buffer.time", "command.buffer.time":
		if n, err := strconv.Atoi(value); err == nil {
			c.BufferTime = n
		}
	}
}

// normalizeCommandKey lowercases key for case-insensitive matching — real
// .cmd files mix key case freely (e.g. "Command.Time", "Buffer.Time").
func normalizeCommandKey(key string) string {
	return strings.ToLower(key)
}

// stripComment removes a ".cmd" comment from line — everything from the
// first ';' to the end of the line, whether the comment stands on its own
// line or trails after real content — and trims surrounding whitespace from
// what remains.
func stripComment(line string) string {
	if idx := strings.IndexByte(line, ';'); idx != -1 {
		line = line[:idx]
	}
	return strings.TrimSpace(line)
}

// parseKeyValueLine splits a "key = value" line on its first '=', trimming
// whitespace from both sides and removing a matching pair of surrounding
// double quotes from the value. ok is false if line has no '=' or an empty
// key.
func parseKeyValueLine(line string) (key, value string, ok bool) {
	idx := strings.IndexByte(line, '=')
	if idx == -1 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false
	}
	value = unquote(strings.TrimSpace(line[idx+1:]))
	return key, value, true
}

// unquote removes a matching pair of surrounding double quotes from s, if
// present.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
