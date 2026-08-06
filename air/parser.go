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
// Clsn1Default/Clsn2Default declarations, indexed Clsn[i] box lines, and the
// Loopstart marker. Comment lines (';', whole-line or trailing) are ignored.
// An empty input returns an empty, non-nil-error result rather than an
// error. Malformed input — an unrecognized action header, a frame line with
// missing or non-numeric fields, or a negative group/image index — returns a
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

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := stripComment(scanner.Text())

		if line == "" {
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

// parseClsnDeclarationHeader recognizes a "Clsn1Default: N", "Clsn2Default:
// N", "Clsn1: N", or "Clsn2: N" line and returns its declaration type and
// the number of Clsn[i] lines that follow it.
func parseClsnDeclarationHeader(line string) (declType string, count int, ok bool) {
	for _, candidate := range []string{"Clsn1Default", "Clsn2Default", "Clsn1", "Clsn2"} {
		prefix := candidate + ":"
		if len(line) >= len(prefix) && strings.EqualFold(line[:len(prefix)], prefix) {
			n, err := strconv.Atoi(strings.TrimSpace(line[len(prefix):]))
			if err != nil {
				return "", 0, false
			}
			return candidate, n, true
		}
	}
	return "", 0, false
}

// clsnBoxLinePattern matches a "Clsn1[0] = L,T,R,B" (or Clsn2) line and
// captures the four coordinates.
var clsnBoxLinePattern = regexp.MustCompile(`(?i)^Clsn[12]\[\d+\]\s*=\s*(-?\d+)\s*,\s*(-?\d+)\s*,\s*(-?\d+)\s*,\s*(-?\d+)\s*$`)

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

// parseFrameLine parses a frame line: "Group,Image, X,Y, Time[, Flip][,
// Blend]". Flip and Blend are optional; a blank Flip field (e.g. an empty
// token between two commas) is treated as FlipNone.
func parseFrameLine(line string) (Frame, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 5 {
		return Frame{}, fmt.Errorf("malformed frame line %q: expected at least 5 comma-separated fields", line)
	}

	group, err := strconv.Atoi(strings.TrimSpace(fields[0]))
	if err != nil {
		return Frame{}, fmt.Errorf("malformed frame line %q: invalid group: %w", line, err)
	}
	if group < 0 {
		return Frame{}, fmt.Errorf("malformed frame line %q: group index must not be negative, got %d", line, group)
	}
	image, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	if err != nil {
		return Frame{}, fmt.Errorf("malformed frame line %q: invalid image: %w", line, err)
	}
	if image < 0 {
		return Frame{}, fmt.Errorf("malformed frame line %q: image index must not be negative, got %d", line, image)
	}
	x, err := strconv.Atoi(strings.TrimSpace(fields[2]))
	if err != nil {
		return Frame{}, fmt.Errorf("malformed frame line %q: invalid x: %w", line, err)
	}
	y, err := strconv.Atoi(strings.TrimSpace(fields[3]))
	if err != nil {
		return Frame{}, fmt.Errorf("malformed frame line %q: invalid y: %w", line, err)
	}
	timeVal, err := strconv.Atoi(strings.TrimSpace(fields[4]))
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
