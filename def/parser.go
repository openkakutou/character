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
// be wrapped in double quotes, which are removed. StateFiles are ordered
// with "st" first, then by numeric suffix; Palettes by numeric suffix; ties
// keep file order (per CharacterInfo's documented field order). An empty
// input returns a zero-value CharacterInfo and a nil error. A malformed section
// header still returns a descriptive error naming the offending line, as
// does a reader that fails outright. A content line inside a known section
// that isn't a valid key=value pair (missing "=") is ignored rather than
// erroring, the same way an unrecognized key already is — real MUGEN/Ikemen
// engines tolerate it.
func Parse(r io.Reader) (CharacterInfo, error) {
	scanner := bufio.NewScanner(r)

	var info CharacterInfo
	var currentSection string

	type paletteEntry struct {
		number int
		path   string
	}
	var palettes []paletteEntry
	var stateFiles []paletteEntry
	var movelists []MovelistRef

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
			// A content line inside a known section that isn't a valid
			// key=value pair (a decorative separator, a truncated leftover
			// key) is ignored, the same way an unrecognized key already is —
			// real MUGEN/Ikemen engines tolerate it. See backlog item 044.
			continue
		}

		switch currentSection {
		case "info":
			switch strings.ToLower(key) {
			case "name":
				info.Name = value
			case "author":
				info.Author = value
			case "mugenversion":
				info.MugenVersion = value
			case "ikemenversion":
				info.IkemenVersion = value
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
			case strings.EqualFold(key, "stcommon"):
				info.StateCommonFile = value
			case stateFileKeyPattern.MatchString(key):
				// "st" has no number: it sorts before every "stN" (number -1).
				number := -1
				if m := stateFileKeyPattern.FindStringSubmatch(key); m[1] != "" {
					number, _ = strconv.Atoi(m[1])
				}
				stateFiles = append(stateFiles, paletteEntry{number: number, path: value})
			case paletteKeyPattern.MatchString(key):
				m := paletteKeyPattern.FindStringSubmatch(key)
				number, _ := strconv.Atoi(m[1])
				palettes = append(palettes, paletteEntry{number: number, path: value})
			default:
				if ref, ok := movelistRef(key, value); ok {
					movelists = append(movelists, ref)
				}
			}
			// Any other key within [Files] is unrecognized and ignored.
		}
	}

	if err := scanner.Err(); err != nil {
		return CharacterInfo{}, fmt.Errorf("def: reading character definition source: %w", err)
	}

	if len(stateFiles) > 0 {
		sort.SliceStable(stateFiles, func(i, j int) bool { return stateFiles[i].number < stateFiles[j].number })
		info.StateFiles = make([]string, len(stateFiles))
		for i, s := range stateFiles {
			info.StateFiles[i] = s.path
		}
	}

	if len(palettes) > 0 {
		sort.SliceStable(palettes, func(i, j int) bool { return palettes[i].number < palettes[j].number })
		info.Palettes = make([]string, len(palettes))
		for i, p := range palettes {
			info.Palettes[i] = p.path
		}
	}

	sort.SliceStable(movelists, func(i, j int) bool {
		if movelists[i].Index != movelists[j].Index {
			return movelists[i].Index < movelists[j].Index
		}
		return isBareMovelistKey(movelists[i].Key) && !isBareMovelistKey(movelists[j].Key)
	})
	info.Movelists = movelists

	return info, nil
}

// movelistKeyPrefix starts every move list key: "movelist", "movelist0",
// "movelist1", ...
const movelistKeyPrefix = "movelist"

// movelistRef reads a [Files] key/value as a move list declaration the way
// Ikemen GO's loadMovelists does (src/char.go:3425-3452, commit
// 54d4148a728153c2097a0fb899a06812b2d3e193): the key is matched
// case-insensitively, "movelist" alone is index 0, a decimal suffix is the
// index, and an empty value, a negative index or any other suffix declares
// nothing.
func movelistRef(key, value string) (MovelistRef, bool) {
	lower := strings.ToLower(key)
	if !strings.HasPrefix(lower, movelistKeyPrefix) || value == "" {
		return MovelistRef{}, false
	}
	suffix := lower[len(movelistKeyPrefix):]
	if suffix == "" {
		return MovelistRef{Key: key, Index: 0, File: value}, true
	}
	index, err := strconv.Atoi(suffix)
	if err != nil || index < 0 {
		return MovelistRef{}, false
	}
	return MovelistRef{Key: key, Index: index, File: value}, true
}

// isBareMovelistKey reports whether key is "movelist" without an index, which
// Ikemen GO tries before "movelist0" (src/char.go:3445-3446).
func isBareMovelistKey(key string) bool {
	return strings.EqualFold(key, movelistKeyPrefix)
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
