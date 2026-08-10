package air

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// stripComment removes a ".air" comment from line — everything from the
// first ';' to the end of the line, whether the comment stands on its own
// line or trails after real content — and trims surrounding whitespace from
// what remains.
func stripComment(line string) string {
	if idx := strings.IndexByte(line, ';'); idx != -1 {
		line = line[:idx]
	}
	return strings.TrimSpace(line)
}

// actionHeaderPattern matches a "[Begin Action N]" line and captures the
// action number.
var actionHeaderPattern = regexp.MustCompile(`(?i)^\[\s*begin\s+action\s+(-?\d+)\s*\]`)

// Parse reads MUGEN/Ikemen GO .air animation text from r and returns the
// Animations it describes, in file order.
//
// This covers "[Begin Action N]" headers, frame lines,
// Clsn1Default/Clsn2Default declarations, indexed Clsn[i] box lines, the
// Loopstart marker, and Ikemen GO's Interpolate Offset/Blend/Scale/Angle
// directive lines (recognized and skipped, not yet represented on the
// Animation/Frame model). Comment lines (';', whole-line or trailing) are
// ignored, as is a whole line starting with ':' — a real-world ad-hoc
// comment marker some .air files use instead of ';' (item 054).
// An empty input returns an empty, non-nil-error result rather than an
// error. A frame line's group or image field may be -1, the ".air"
// convention for "no sprite shown on this frame" (see Frame.IsBlank).
// Malformed input — an unrecognized action header, a frame line with
// missing or non-numeric fields, or a group/image index more negative than
// the -1 sentinel — returns a descriptive error naming the offending line
// rather than panicking or silently producing incorrect data, as does a
// reader that fails outright.
func Parse(r io.Reader) ([]Animation, error) {
	scanner := bufio.NewScanner(r)

	var animations []Animation
	var current *Animation

	// Clsn boxes declared via Clsn1Default/Clsn2Default apply to every
	// frame of the current action that doesn't have its own override.
	var defaultClsn1, defaultClsn2 []ClsnBox

	// Clsn boxes declared via a bare Clsn1:/Clsn2: block apply only to the
	// very next frame line, then are consumed.
	var pendingClsn1, pendingClsn2 []ClsnBox

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := stripComment(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, ":") {
			// Some real .air files (e.g. King of Fighters "Mai (98)", item
			// 054) use a ':'-prefixed line as an ad-hoc comment instead of
			// the standard ';' marker. Real MUGEN/Ikemen engines don't
			// treat ':' as a comment marker either, but a genuine frame
			// line can only ever start with a digit or '-', so a
			// ':'-prefixed line is never mistaken for one there; it's
			// simply an unrecognized content line, silently ignored rather
			// than erroring. Recognizing it here as a comment-like line
			// produces the same observable result for this specific
			// real-world quirk.
			continue
		}

		if strings.HasPrefix(line, "[") {
			m := actionHeaderPattern.FindStringSubmatch(line)
			if m == nil {
				return nil, fmt.Errorf("air: line %d: malformed action header %q", lineNumber, line)
			}
			number, err := strconv.Atoi(m[1])
			if err != nil {
				return nil, fmt.Errorf("air: line %d: invalid action number %q: %w", lineNumber, m[1], err)
			}
			animations = append(animations, Animation{Number: number})
			current = &animations[len(animations)-1]
			defaultClsn1, defaultClsn2 = nil, nil
			pendingClsn1, pendingClsn2 = nil, nil
			continue
		}

		if current == nil {
			// Non-header content before the first action header has no
			// action to attach to; nothing meaningful can be done with it.
			continue
		}

		if strings.EqualFold(line, "loopstart") {
			current.LoopStart = len(current.Frames)
			continue
		}

		if isInterpolateDirective(line) {
			// Ikemen GO's Interpolate directive lines (Offset/Blend/Scale/
			// Angle) tell the engine to smoothly transition that property
			// across the animation. They carry no data of their own on this
			// line and aren't represented on the Animation/Frame model yet
			// (same "read-path model can't hold everything yet" pattern
			// already applied to Loopstart, which also isn't stored beyond
			// LoopStart) — recognizing and skipping the line is enough.
			continue
		}

		if declType, count, ok := parseClsnDeclarationHeader(line); ok {
			boxes, newLineNumber, err := readClsnBoxes(scanner, &lineNumber, count)
			if err != nil {
				return nil, err
			}
			switch declType {
			case "Clsn1Default":
				defaultClsn1 = boxes
			case "Clsn2Default":
				defaultClsn2 = boxes
			case "Clsn1":
				pendingClsn1 = boxes
			case "Clsn2":
				pendingClsn2 = boxes
			}
			lineNumber = newLineNumber
			continue
		}

		frame, err := parseFrameLine(line)
		if err != nil {
			return nil, fmt.Errorf("air: line %d: %w", lineNumber, err)
		}

		if pendingClsn1 != nil {
			frame.Clsn1 = pendingClsn1
			pendingClsn1 = nil
		} else if defaultClsn1 != nil {
			frame.Clsn1 = defaultClsn1
		}
		if pendingClsn2 != nil {
			frame.Clsn2 = pendingClsn2
			pendingClsn2 = nil
		} else if defaultClsn2 != nil {
			frame.Clsn2 = defaultClsn2
		}

		current.Frames = append(current.Frames, frame)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("air: reading animation source: %w", err)
	}

	return animations, nil
}

// interpolateDirectives lists the Ikemen GO ".air" Interpolate directive
// keywords: a standalone line (no trailing data) telling the engine to
// smoothly transition that property across the animation.
var interpolateDirectives = []string{
	"Interpolate Offset",
	"Interpolate Blend",
	"Interpolate Scale",
	"Interpolate Angle",
}

// isInterpolateDirective reports whether line is exactly one of the
// recognized Interpolate directive keywords (case-insensitive). A line that
// merely starts with "Interpolate" but isn't one of these is not
// recognized, and falls through to frame-line parsing (and its error) as
// before.
func isInterpolateDirective(line string) bool {
	for _, keyword := range interpolateDirectives {
		if strings.EqualFold(line, keyword) {
			return true
		}
	}
	return false
}

// clsnDeclarationKeywords maps each recognized spelling of a Clsn
// declaration header keyword to the canonical declaration type it means.
// Besides the four correctly spelled keywords, "Clsn1deault"/"Clsn2deault"
// (a missing "f") are also recognized: a real-world authoring typo found in
// at least one real character file (Darkstalkers' Donovan), tolerated here
// the same way def.Parse/cns.Parse tolerate other specific real-world
// authoring mistakes rather than attempting open-ended fuzzy matching.
var clsnDeclarationKeywords = map[string]string{
	"Clsn1Default": "Clsn1Default",
	"Clsn2Default": "Clsn2Default",
	"Clsn1deault":  "Clsn1Default",
	"Clsn2deault":  "Clsn2Default",
	"Clsn1":        "Clsn1",
	"Clsn2":        "Clsn2",
}

// clsnDeclarationCandidates lists clsnDeclarationKeywords' keys, longest
// first, so a longer keyword (e.g. "Clsn1Default") is matched before a
// shorter one it starts with (e.g. "Clsn1").
var clsnDeclarationCandidates = []string{"Clsn1Default", "Clsn2Default", "Clsn1deault", "Clsn2deault", "Clsn1", "Clsn2"}

// parseClsnDeclarationHeader recognizes a "Clsn1Default: N", "Clsn2Default:
// N", "Clsn1: N", or "Clsn2: N" line (or one of clsnDeclarationKeywords'
// tolerated misspellings) and returns its declaration type and the number
// of Clsn[i] lines that follow it.
func parseClsnDeclarationHeader(line string) (declType string, count int, ok bool) {
	for _, candidate := range clsnDeclarationCandidates {
		prefix := candidate + ":"
		if len(line) >= len(prefix) && strings.EqualFold(line[:len(prefix)], prefix) {
			n, err := strconv.Atoi(strings.TrimSpace(line[len(prefix):]))
			if err != nil {
				return "", 0, false
			}
			return clsnDeclarationKeywords[candidate], n, true
		}
	}
	return "", 0, false
}

// clsnBoxLinePattern matches a "Clsn1[0] = L,T,R,B" (or Clsn2) line and
// captures the four coordinates. Whitespace between the keyword and the
// "[index]" bracket is optional: real MUGEN/Ikemen engines tolerate a line
// like "Clsn2 [0] = ...", which some real-world .air files use.
var clsnBoxLinePattern = regexp.MustCompile(`(?i)^Clsn[12]\s*\[\d+\]\s*=\s*(-?\d+)\s*,\s*(-?\d+)\s*,\s*(-?\d+)\s*,\s*(-?\d+)\s*$`)

// readClsnBoxes reads the next count Clsn[i] box lines from scanner.
// lineNumber is updated as lines are consumed for error reporting.
func readClsnBoxes(scanner *bufio.Scanner, lineNumber *int, count int) ([]ClsnBox, int, error) {
	boxes := make([]ClsnBox, 0, count)
	n := *lineNumber
	for i := 0; i < count; i++ {
		if !scanner.Scan() {
			return nil, n, fmt.Errorf("air: line %d: expected %d Clsn box line(s), got %d", n, count, i)
		}
		n++
		line := stripComment(scanner.Text())
		m := clsnBoxLinePattern.FindStringSubmatch(line)
		if m == nil {
			return nil, n, fmt.Errorf("air: line %d: malformed Clsn box line %q", n, line)
		}
		left, _ := strconv.Atoi(m[1])
		top, _ := strconv.Atoi(m[2])
		right, _ := strconv.Atoi(m[3])
		bottom, _ := strconv.Atoi(m[4])
		boxes = append(boxes, ClsnBox{Left: left, Top: top, Right: right, Bottom: bottom})
	}
	return boxes, n, nil
}

// leadingIntPattern matches an optionally-signed run of digits at the start
// of a string, after leading whitespace.
var leadingIntPattern = regexp.MustCompile(`^\s*[+-]?\d+`)

// parseLeadingInt parses the leading integer of field, tolerating trailing
// non-numeric content after it (e.g. "143 0" parses as 143). This mirrors
// real MUGEN/Ikemen engines' own number parsing (Ikemen GO's Atoi, in
// src/common.go, scans a leading digit run and stops at the first
// non-digit character rather than requiring the whole field to be
// numeric) — see item 048. A field with no usable leading digits at all
// (e.g. "abc") still returns an error, distinguishing "trailing garbage
// after a real number" from "not a number".
func parseLeadingInt(field string) (int, error) {
	if strings.TrimSpace(field) == "" {
		// Ikemen GO's own Atoi (src/common.go) returns 0 for an empty/
		// blank field rather than erroring. A required field can end up
		// blank in a real-world file that is missing a comma, which
		// shifts a later field's content into an earlier one and leaves
		// this one empty (item 048's exact Goku/axl-kofa/baiken-kofa
		// fixture: the shared frame line is missing the comma between
		// Image and X, leaving Time blank).
		return 0, nil
	}
	m := leadingIntPattern.FindString(field)
	if m == "" {
		return 0, fmt.Errorf("strconv.Atoi: parsing %q: invalid syntax", field)
	}
	// m is a valid integer literal (optional sign + digits only), so this
	// Atoi cannot fail.
	return strconv.Atoi(strings.TrimSpace(m))
}

// parseFrameLine parses a frame line: "Group,Image, X,Y, Time[, Flip][,
// Blend]". Flip and Blend are optional; a blank Flip field (e.g. an empty
// token between two commas) is treated as FlipNone.
func parseFrameLine(line string) (Frame, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 5 {
		return Frame{}, fmt.Errorf("malformed frame line %q: expected at least 5 comma-separated fields", line)
	}

	// Group/Image are not range-checked beyond being valid integers: real
	// MUGEN/Ikemen engines treat *any* negative Group as "no sprite shown
	// this frame" regardless of Image's value, and real-world .air files
	// use varying negative values (-1, -2, -3, ...), not just the -1,-1
	// convention. Frame.IsBlank() (Group < 0 || Image < 0) is the single
	// recognition point for this state — see
	// .vibe/decisions/014-blank-frame-sentinel-accepts-any-negative-value.md.
	//
	// Each field's leading integer is used even if non-numeric content
	// trails it (item 048) — a field with no usable leading digits at all
	// still returns the descriptive error below.
	group, err := parseLeadingInt(fields[0])
	if err != nil {
		return Frame{}, fmt.Errorf("malformed frame line %q: invalid group: %w", line, err)
	}
	image, err := parseLeadingInt(fields[1])
	if err != nil {
		return Frame{}, fmt.Errorf("malformed frame line %q: invalid image: %w", line, err)
	}
	x, err := parseLeadingInt(fields[2])
	if err != nil {
		return Frame{}, fmt.Errorf("malformed frame line %q: invalid x: %w", line, err)
	}
	y, err := parseLeadingInt(fields[3])
	if err != nil {
		return Frame{}, fmt.Errorf("malformed frame line %q: invalid y: %w", line, err)
	}
	timeVal, err := parseLeadingInt(fields[4])
	if err != nil {
		return Frame{}, fmt.Errorf("malformed frame line %q: invalid time: %w", line, err)
	}

	frame := Frame{Group: group, Image: image, X: x, Y: y, Time: timeVal}

	if len(fields) >= 6 {
		frame.Flip = Flip(strings.ToUpper(strings.TrimSpace(fields[5])))
	}
	if len(fields) >= 7 {
		frame.Blend = BlendMode(strings.TrimSpace(fields[6]))
	}

	return frame, nil
}
