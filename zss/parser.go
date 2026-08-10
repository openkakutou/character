// Parse turns Ikemen GO .zss text into a Script — the read-path entry point
// for this package.
package zss

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// statedefHeaderPattern matches a "[Statedef N]" or
// "[Statedef N; key: value; ...]" line, capturing the state number and the
// raw text of the semicolon-separated header parameters (possibly empty).
// Case-insensitive: real .zss files use both "Statedef" and "StateDef".
var statedefHeaderPattern = regexp.MustCompile(`(?i)^\[\s*statedef\s+(-?\d+)\s*;?\s*(.*)\]$`)

// functionHeaderPattern matches a "[Function Name(params) ret]" line,
// capturing the function name, its raw comma-separated parameter list
// (possibly empty), and its raw comma-separated return variable list
// (possibly empty). Case-insensitive: real .zss files use both "Function"
// and "function".
var functionHeaderPattern = regexp.MustCompile(`(?i)^\[\s*function\s+([A-Za-z_]\w*)\s*\(([^)]*)\)\s*(.*)\]$`)

// Parse reads .zss text from r and returns the Script it describes: an
// ordered list of Blocks ("[Statedef N; ...]" or "[Function Name(...) ...]"
// headers), each carrying its script body as raw, unevaluated text (see the
// package doc comment), plus any Preamble content that precedes the first
// block header.
//
// A line starting with "[" that matches neither header pattern returns a
// descriptive, line-numbered error rather than being silently skipped:
// unlike .cns (used by both MUGEN and Ikemen GO, with decades of divergent
// real-world authoring habits to tolerate), .zss is Ikemen-GO-only and a
// corpus scan found no top-level content other than these two header
// shapes. An empty input returns an empty, nil-error Script.
func Parse(r io.Reader) (Script, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var script Script
	var current *Block
	var bodyLines []string
	var preambleLines []string
	inPreamble := true

	flushBody := func() {
		if current != nil {
			current.Body = joinLines(bodyLines)
		}
		bodyLines = nil
	}

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if strings.HasPrefix(line, "[") {
			startLine := lineNumber
			trimmed, err := readHeaderText(scanner, line, &lineNumber)
			if err != nil {
				return Script{}, fmt.Errorf("zss: line %d: %w", startLine, err)
			}

			if m := statedefHeaderPattern.FindStringSubmatch(trimmed); m != nil {
				flushBody()
				number, err := strconv.Atoi(m[1])
				if err != nil {
					return Script{}, fmt.Errorf("zss: line %d: invalid Statedef number %q: %w", lineNumber, m[1], err)
				}
				script.Blocks = append(script.Blocks, Block{
					Kind:         BlockKindStatedef,
					Number:       number,
					HeaderParams: parseHeaderParams(m[2]),
				})
				current = &script.Blocks[len(script.Blocks)-1]
				inPreamble = false
				continue
			}

			if m := functionHeaderPattern.FindStringSubmatch(trimmed); m != nil {
				flushBody()
				script.Blocks = append(script.Blocks, Block{
					Kind:   BlockKindFunction,
					Name:   m[1],
					Params: splitTrimmed(m[2], ","),
					Ret:    splitTrimmed(m[3], ","),
				})
				current = &script.Blocks[len(script.Blocks)-1]
				inPreamble = false
				continue
			}

			return Script{}, fmt.Errorf("zss: line %d: malformed block header %q", startLine, trimmed)
		}

		if inPreamble {
			preambleLines = append(preambleLines, line)
			continue
		}
		bodyLines = append(bodyLines, line)
	}
	flushBody()

	if err := scanner.Err(); err != nil {
		return Script{}, fmt.Errorf("zss: reading script source: %w", err)
	}

	script.Preamble = joinLines(preambleLines)
	return script, nil
}

// readHeaderText returns the logical single-line text of a block header
// that starts at first, joining continuation lines with a single space
// until a line ending in "]" is found. Real .zss files sometimes wrap a
// Statedef header's semicolon-separated parameters across several physical
// lines, closing the bracket only on a later line (e.g. a corpus-found
// "[StateDef 801;" / "type: S; ...;" / "anim: 801; ...ctrl: 0;]" split
// across three lines) — real Ikemen GO engines tolerate this, so this
// parser does too rather than treating the header as a single-line-only
// construct. *lineNumber is advanced to the line each continuation line was
// read from, so callers keep accurate line-numbered error reporting.
// Returns an error if the input ends before a closing "]" is found.
func readHeaderText(scanner *bufio.Scanner, first string, lineNumber *int) (string, error) {
	text := strings.TrimSpace(first)
	for !strings.HasSuffix(text, "]") {
		if !scanner.Scan() {
			return "", fmt.Errorf("unterminated block header %q", text)
		}
		*lineNumber++
		text += " " + strings.TrimSpace(scanner.Text())
	}
	return text, nil
}

// parseHeaderParams splits a Statedef header's raw semicolon-separated
// parameter text (e.g. "type: S; physics: N; velSet:0,0;") into a
// lowercase-keyed map, tolerating a trailing ";" and inconsistent spacing.
// A segment with no ":" (malformed, or a stray trailing empty segment) is
// skipped rather than erroring, the same tolerance cns.Parse already applies
// to unrecognized content. Returns nil for empty input.
func parseHeaderParams(raw string) map[string]string {
	segments := strings.Split(raw, ";")
	var params map[string]string
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		idx := strings.IndexByte(seg, ':')
		if idx == -1 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(seg[:idx]))
		if key == "" {
			continue
		}
		value := strings.TrimSpace(seg[idx+1:])
		if params == nil {
			params = make(map[string]string)
		}
		params[key] = value
	}
	return params
}

// splitTrimmed splits raw on sep, trims whitespace from each part, and
// drops empty parts — used for a Function header's parenthesized parameter
// list and return variable list. Returns nil for empty input.
func splitTrimmed(raw, sep string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// joinLines reassembles lines (as produced by bufio.Scanner, one per source
// line with no trailing newline) back into a single newline-terminated
// block of text — the raw form Block.Body/Script.Preamble store. Returns ""
// for no lines.
func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
