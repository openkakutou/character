// Parse turns MUGEN/Ikemen GO .def text into a CharacterInfo — the read-path
// entry point for this package.
package def

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// stateFileKeyPattern matches the "st", "st1", "st2", ... keys that
// contribute to CharacterInfo.StateFiles.
var stateFileKeyPattern = regexp.MustCompile(`(?i)^st(\d*)$`)

// paletteKeyPattern matches the "pal1", "pal2", ... keys that contribute to
// CharacterInfo.Palettes, capturing the palette number.
var paletteKeyPattern = regexp.MustCompile(`(?i)^pal(\d+)$`)

// Parse reads MUGEN/Ikemen GO .def character definition text from r and
// returns the CharacterInfo it describes.
//
// Only the "[Info]" and "[Files]" sections (matched case-insensitively) are
// recognized; any other section's lines are skipped without validation, and
// parsing continues into the next known section rather than aborting — see
// .vibe/decisions/009-def-parse-ignores-unknown-sections.md. Within a known
// section, an unrecognized key is likewise ignored. Comment lines (';',
// whole-line or trailing) are stripped before parsing. Values may optionally
// be wrapped in double quotes, which are removed. StateFiles are collected
// in file-appearance order; Palettes are collected and then sorted by their
// numeric suffix, per CharacterInfo's documented field order. An empty input
// returns a zero-value CharacterInfo and a nil error. A malformed section
// header or a key=value line missing "=" inside a known section returns a
// descriptive error naming the offending line, as does a reader that fails
// outright.
func Parse(r io.Reader) (CharacterInfo, error) {
	scanner := bufio.NewScanner(r)

	var info CharacterInfo
	var currentSection string

	type paletteEntry struct {
		number int
		path   string
	}
	var palettes []paletteEntry

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := stripComment(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return CharacterInfo{}, fmt.Errorf("def: line %d: malformed section header %q", lineNumber, line)
			}
			currentSection = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}

		if currentSection != "info" && currentSection != "files" {
			// Content of an unrecognized section has no place in
			// CharacterInfo; skip it without validation.
			continue
		}

		key, value, ok := parseKeyValueLine(line)
		if !ok {
			return CharacterInfo{}, fmt.Errorf("def: line %d: malformed key=value line %q", lineNumber, line)
		}

		switch currentSection {
		case "info":
			switch strings.ToLower(key) {
			case "name":
				info.Name = value
			case "author":
				info.Author = value
			}
		case "files":
			switch {
			case strings.EqualFold(key, "sprite"):
				info.SpriteFile = value
			case strings.EqualFold(key, "anim"):
				info.AnimationFile = value
			case strings.EqualFold(key, "sound"):
				info.SoundFile = value
			case strings.EqualFold(key, "cmd"):
				info.CommandFile = value
			case strings.EqualFold(key, "cns"):
				info.ConstantsFile = value
			case stateFileKeyPattern.MatchString(key):
				info.StateFiles = append(info.StateFiles, value)
			case paletteKeyPattern.MatchString(key):
				m := paletteKeyPattern.FindStringSubmatch(key)
				number, _ := strconv.Atoi(m[1])
				palettes = append(palettes, paletteEntry{number: number, path: value})
			}
			// Any other key within [Files] is unrecognized and ignored.
		}
	}

	if err := scanner.Err(); err != nil {
		return CharacterInfo{}, fmt.Errorf("def: reading character definition source: %w", err)
	}

	if len(palettes) > 0 {
		sort.Slice(palettes, func(i, j int) bool { return palettes[i].number < palettes[j].number })
		info.Palettes = make([]string, len(palettes))
		for i, p := range palettes {
			info.Palettes[i] = p.path
		}
	}

	return info, nil
}

// stripComment removes a ".def" comment from line — everything from the
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
