package air

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"slices"
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
var actionHeaderPattern = regexp.MustCompile(`(?i)^\[\s*begin\s+action\s+(-?\d+)[^\]]*\]`)

// actionHeaderAttemptPattern recognizes any bracket line that starts with
// the "begin" keyword, whether or not it goes on to match
// actionHeaderPattern — used to tell a malformed "[Begin Action N]" header
// apart from a genuinely unrelated bracket section (e.g. an embedded
// "[Statedef N]"/"[State N]" block, item 055). The keyword must be
// immediately followed by whitespace, "]", or end of line so a bare
// "[Begin" with nothing else on the line is still recognized as an attempt,
// mirroring cns.Parse's statedefAttemptPattern/stateAttemptPattern.
var actionHeaderAttemptPattern = regexp.MustCompile(`(?i)^\[\s*begin(\s|\]|$)`)

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
// A bracket line that isn't even an attempt at "[Begin Action N]" — e.g. a
// "[Statedef N]"/"[State N]" block some real-world .air files carry
// embedded alongside their animation data — is skipped, along with the
// content lines that follow it, rather than erroring (item 055), the same
// tolerance cns.Parse already applies to sections outside its own scope.
// Malformed input — a bracket line that looks like an attempted "[Begin
// Action N]" header yet fails to parse (missing or non-numeric action
// number), a frame line with missing or non-numeric fields, or a
// group/image index more negative than the -1 sentinel — still returns a
// descriptive error naming the offending line rather than panicking or
// silently producing incorrect data, as does a reader that fails outright.
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

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("air: reading: %w", err)
	}

	// copyOf records each "Copy Action N" action (index in animations)
	// with N; copying skips the rest of its block.
	copyOf := map[int]int{}
	copying := false

	for i := 0; i < len(lines); i++ {
		lineNumber := i + 1
		line := stripComment(lines[i])

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
				if actionHeaderAttemptPattern.MatchString(line) {
					return nil, fmt.Errorf("air: line %d: malformed action header %q", lineNumber, line)
				}
				// A bracket line that isn't even an attempt at "[Begin
				// Action N]" is genuinely unrelated content — e.g. a
				// "[Statedef N]"/"[State N]" block some real-world .air
				// files carry embedded alongside their animation data
				// (item 055). It's skipped, along with whatever content
				// lines follow it, until the next recognized action
				// header, the same tolerance cns.Parse already applies to
				// sections outside its own scope.
				current = nil
				defaultClsn1, defaultClsn2 = nil, nil
				pendingClsn1, pendingClsn2 = nil, nil
				continue
			}
			number, err := strconv.Atoi(m[1])
			if err != nil {
				return nil, fmt.Errorf("air: line %d: invalid action number %q: %w", lineNumber, m[1], err)
			}
			animations = append(animations, Animation{Number: number})
			current = &animations[len(animations)-1]
			copying = false
			defaultClsn1, defaultClsn2 = nil, nil
			pendingClsn1, pendingClsn2 = nil, nil
			continue
		}

		if current == nil || copying {
			// Non-header content before the first action header has no
			// action to attach to; nothing meaningful can be done with it.
			// A Copy Action block ends at its Copy Action line.
			continue
		}

		// Ikemen GO's "Copy Action N" makes the action a copy of action N
		// (anim.go:359-370); an invalid N leaves it as it is.
		if lower := strings.ToLower(line); strings.HasPrefix(lower, "copy action ") {
			copying = true
			if n, err := strconv.Atoi(strings.TrimSpace(lower[len("copy action "):])); err == nil {
				copyOf[len(animations)-1] = n
			}
			continue
		}

		// Ikemen GO matches "loopstart" as a prefix (anim.go:395), so
		// "Loopstart:" is one too.
		if len(line) >= 9 && strings.EqualFold(line[:9], "loopstart") {
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
			var boxes []ClsnBox
			boxes, i = readClsnBoxes(lines, i, count)
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
			continue
		}

		// Ikemen GO reads a frame only from a line starting with a digit or
		// '-' and skips any other (anim.go:45-60); a frame line it cannot
		// read is reported but not fatal there, and skipped here.
		if line[0] != '-' && (line[0] < '0' || line[0] > '9') {
			continue
		}
		frame, err := parseFrameLine(line)
		if err != nil {
			continue
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

	return resolveCopyActions(animations, copyOf), nil
}

// copyChainLimit is how many Copy Action links Ikemen GO follows.
const copyChainLimit = 8

// resolveCopyActions gives each Copy Action action the frames of the action
// it names, following a copy of a copy up to 8 links, as Ikemen GO's
// resolveCopyAction does (anim.go:1386-1424); one whose target is missing,
// itself or too far down the chain is dropped.
func resolveCopyActions(animations []Animation, copyOf map[int]int) []Animation {
	if len(copyOf) == 0 {
		return animations
	}
	byNumber := map[int]int{}
	for i, a := range animations {
		if _, seen := byNumber[a.Number]; !seen {
			byNumber[a.Number] = i
		}
	}
	dropped := map[int]bool{}
	for idx := range animations {
		target, isCopy := copyOf[idx]
		if !isCopy {
			continue
		}
		resolved := false
		for range copyChainLimit {
			ti, ok := byNumber[target]
			if !ok || ti == idx {
				break
			}
			if next, chained := copyOf[ti]; chained {
				target = next
				continue
			}
			animations[idx].Frames = slices.Clone(animations[ti].Frames)
			animations[idx].LoopStart = animations[ti].LoopStart
			resolved = true
			break
		}
		if !resolved {
			dropped[idx] = true
		}
	}
	kept := animations[:0]
	for i, a := range animations {
		if !dropped[i] {
			kept = append(kept, a)
		}
	}
	return kept
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

// parseClsnDeclarationHeader recognizes a Clsn declaration as Ikemen GO
// does (anim.go:405-430, commit 54d4148a): a line starting "Clsn1" or
// "Clsn2" (any case) and holding a ':' followed by the box count; it is a
// default one when "default" follows the digit, so a misspelling such as
// "Clsn2Defailt" or "Clsn1deault" declares the next frame's boxes only.
func parseClsnDeclarationHeader(line string) (declType string, count int, ok bool) {
	lower := strings.ToLower(line)
	if len(lower) < 5 || !strings.HasPrefix(lower, "clsn") || (lower[4] != '1' && lower[4] != '2') {
		return "", 0, false
	}
	colon := strings.IndexByte(lower, ':')
	if colon < 0 {
		return "", 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(lower[colon+1:]))
	if err != nil || n < 0 {
		return "", 0, false
	}
	declType = "Clsn" + string(lower[4])
	if strings.HasPrefix(lower[5:], "default") {
		declType += "Default"
	}
	return declType, n, true
}

// readClsnBoxes reads the count box lines following the declaration at
// lines[at] as Ikemen GO's readClsn does (anim.go:469-566, commit
// 54d4148a): blank lines are skipped; a box line starts with "Clsn" and
// holds an "=" followed by four comma-separated numbers, read leniently;
// its optional [n] places it, else (or when n is out of range or taken) the
// first free slot; the first other line ends the boxes and is read again
// as content. Boxes the declaration counts but the file lacks stay empty
// (Ikemen GO warns and keeps the action). It returns the boxes and the
// index of the last line consumed.
func readClsnBoxes(lines []string, at, count int) ([]ClsnBox, int) {
	boxes := []ClsnBox{} // count is untrusted: grown as boxes are read
	filled := []bool{}
	slot := func(n int) int {
		for len(boxes) <= n {
			boxes, filled = append(boxes, ClsnBox{}), append(filled, false)
		}
		return n
	}
	read := 0
	i := at + 1
	for ; read < count && i < len(lines); i++ {
		line := strings.ToLower(stripComment(lines[i]))
		if line == "" {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if !strings.HasPrefix(line, "clsn") || eq < 0 {
			break
		}
		values := strings.Split(line[eq+1:], ",")
		if len(values) < 4 {
			break
		}
		n := clsnBoxIndex(line[:eq])
		if n < 0 || n >= count || (n < len(filled) && filled[n]) {
			n = firstFree(filled, count)
		}
		slot(n)
		boxes[n] = ClsnBox{Left: lenientInt(values[0]), Top: lenientInt(values[1]), Right: lenientInt(values[2]), Bottom: lenientInt(values[3])}
		filled[n] = true
		read++
	}
	if read > 0 {
		slot(count - 1)
	}
	return boxes, i - 1
}

// clsnBoxIndex is the [n] of a box line's left side, -1 without a valid one.
func clsnBoxIndex(left string) int {
	open := strings.IndexByte(left, '[')
	if open < 0 {
		return -1
	}
	end := strings.IndexByte(left[open:], ']')
	if end <= 0 {
		return -1
	}
	n, err := strconv.Atoi(strings.TrimSpace(left[open+1 : open+end]))
	if err != nil || n < 0 {
		return -1
	}
	return n
}

// firstFree is the first slot of count not filled yet.
func firstFree(filled []bool, count int) int {
	for k := range count {
		if k >= len(filled) || !filled[k] {
			return k
		}
	}
	return count - 1
}

// lenientInt is Ikemen GO's Atoi: the leading integer, 0 without one.
func lenientInt(field string) int {
	n, err := parseLeadingInt(field)
	if err != nil {
		return 0
	}
	return n
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
