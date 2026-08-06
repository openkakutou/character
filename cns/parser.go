// Parse turns MUGEN/Ikemen GO .cns text into StateDefs — the read-path entry
// point for this package.
package cns

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// statedefHeaderPattern matches a "[Statedef N]" line, capturing the state
// number. A trailing ", label" comment (e.g. "[Statedef 0, Standing]") is
// tolerated and ignored, matching common .cns authoring style.
var statedefHeaderPattern = regexp.MustCompile(`(?i)^\[\s*statedef\s+(-?\d+)\s*(?:,[^\]]*)?\]$`)

// statedefAttemptPattern recognizes any bracket line that starts with the
// "statedef" keyword, whether or not it goes on to match
// statedefHeaderPattern — used to tell a malformed Statedef header apart
// from a genuinely unrelated section. See
// .vibe/decisions/012-cns-parse-header-detection-strategy.md.
var statedefAttemptPattern = regexp.MustCompile(`(?i)^\[\s*statedef(\s|\])`)

// stateHeaderPattern matches a "[State N]" line, capturing the number of the
// Statedef it belongs to (not stored on Controller — see cns.Controller). A
// trailing ", label" comment is tolerated and ignored, same as
// statedefHeaderPattern.
var stateHeaderPattern = regexp.MustCompile(`(?i)^\[\s*state\s+(-?\d+)\s*(?:,[^\]]*)?\]$`)

// stateAttemptPattern recognizes any bracket line that starts with the
// "state" keyword. It never fires on a "[Statedef ...]" line: "statedef" has
// no space (or "]") right after "state", so it fails the (\s|\]) lookahead
// that both attempt patterns require.
var stateAttemptPattern = regexp.MustCompile(`(?i)^\[\s*state(\s|\])`)

// Parse reads .cns text from r and returns the StateDefs ("[Statedef N]"
// blocks) it describes, in file order, each carrying its state controllers
// ("[State N]" blocks) in file order.
//
// Trigger keys ("trigger1", "trigger2", ..., "triggerall", matched by a
// "trigger" prefix) are collected into a Controller's Triggers in file
// order rather than evaluated; the "type" key sets Controller.Type; every
// other key becomes a Parameters entry, normalized to lowercase for
// predictable lookup. A bracket section that is neither a valid
// "[Statedef N]" nor "[State N]" header is skipped without validating its
// content, matching def.Parse's tolerance for sections outside this
// package's scope — but a bracket line that looks like an attempted
// Statedef/State header yet fails to parse (a non-numeric or missing state
// number, or a missing closing bracket) returns a descriptive,
// line-numbered error rather than being silently skipped, as does a
// "[State N]" block with no enclosing Statedef and a reader that fails
// outright. See .vibe/decisions/012-cns-parse-header-detection-strategy.md.
// Comment lines (';', whole-line or trailing) are ignored. An empty input
// returns an empty, nil-error result.
func Parse(r io.Reader) ([]StateDef, error) {
	scanner := bufio.NewScanner(r)

	var states []StateDef
	var current *StateDef
	var currentCtrl *Controller

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := stripComment(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return nil, fmt.Errorf("cns: line %d: malformed section header %q", lineNumber, line)
			}

			if m := statedefHeaderPattern.FindStringSubmatch(line); m != nil {
				number, err := strconv.Atoi(m[1])
				if err != nil {
					return nil, fmt.Errorf("cns: line %d: invalid Statedef number %q: %w", lineNumber, m[1], err)
				}
				states = append(states, StateDef{Number: number})
				current = &states[len(states)-1]
				currentCtrl = nil
				continue
			}
			if statedefAttemptPattern.MatchString(line) {
				return nil, fmt.Errorf("cns: line %d: malformed Statedef header %q", lineNumber, line)
			}

			if m := stateHeaderPattern.FindStringSubmatch(line); m != nil {
				if current == nil {
					return nil, fmt.Errorf("cns: line %d: [State N] block found outside of any Statedef: %q", lineNumber, line)
				}
				current.Controllers = append(current.Controllers, Controller{})
				currentCtrl = &current.Controllers[len(current.Controllers)-1]
				continue
			}
			if stateAttemptPattern.MatchString(line) {
				return nil, fmt.Errorf("cns: line %d: malformed State header %q", lineNumber, line)
			}

			// An unrecognized section: its content has no place in this
			// model, so following lines aren't attributed to any
			// Statedef/Controller until the next recognized header.
			current = nil
			currentCtrl = nil
			continue
		}

		if current == nil {
			// Content before the first Statedef, or inside an
			// unrecognized section, has nothing to attach to.
			continue
		}

		key, value, ok := parseKeyValueLine(line)
		if !ok {
			return nil, fmt.Errorf("cns: line %d: malformed key=value line %q", lineNumber, line)
		}

		if currentCtrl != nil {
			applyControllerField(currentCtrl, key, value)
			continue
		}

		if err := applyStatedefField(current, key, value); err != nil {
			return nil, fmt.Errorf("cns: line %d: %w", lineNumber, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cns: reading combat logic source: %w", err)
	}

	return states, nil
}

// applyControllerField applies a single key=value line to a Controller
// being built: a "trigger..." key appends to Triggers, "type" sets Type,
// and any other key becomes a lowercase-normalized Parameters entry. This
// can never fail — Controller stores everything as unevaluated strings.
func applyControllerField(c *Controller, key, value string) {
	switch {
	case strings.HasPrefix(strings.ToLower(key), "trigger"):
		c.Triggers = append(c.Triggers, value)
	case strings.EqualFold(key, "type"):
		c.Type = value
	default:
		if c.Parameters == nil {
			c.Parameters = make(map[string]string)
		}
		c.Parameters[strings.ToLower(key)] = value
	}
}

// applyStatedefField applies a single key=value line to a StateDef header
// being built. An unrecognized key is ignored. A recognized key with a
// value that doesn't parse as the field's type (int or bool) returns a
// descriptive error.
func applyStatedefField(s *StateDef, key, value string) error {
	switch strings.ToLower(key) {
	case "type":
		s.Type = StateType(value)
	case "movetype":
		s.MoveType = MoveType(value)
	case "physics":
		s.Physics = PhysicsType(value)
	case "anim":
		n, err := parseIntField("anim", value)
		if err != nil {
			return err
		}
		s.Anim = n
	case "ctrl":
		b, err := parseBoolField("ctrl", value)
		if err != nil {
			return err
		}
		s.Ctrl = b
	case "poweradd":
		n, err := parseIntField("poweradd", value)
		if err != nil {
			return err
		}
		s.PowerAdd = n
	case "juggle":
		n, err := parseIntField("juggle", value)
		if err != nil {
			return err
		}
		s.Juggle = n
	case "facep2":
		b, err := parseBoolField("facep2", value)
		if err != nil {
			return err
		}
		s.FaceP2 = b
	case "hitdefpersist":
		b, err := parseBoolField("hitdefpersist", value)
		if err != nil {
			return err
		}
		s.HitDefPersist = b
	case "movehitpersist":
		b, err := parseBoolField("movehitpersist", value)
		if err != nil {
			return err
		}
		s.MoveHitPersist = b
	case "hitcountpersist":
		b, err := parseBoolField("hitcountpersist", value)
		if err != nil {
			return err
		}
		s.HitCountPersist = b
	case "sprpriority":
		n, err := parseIntField("sprpriority", value)
		if err != nil {
			return err
		}
		s.SprPriority = n
	}
	// Any other key is unrecognized within a Statedef header and ignored.
	return nil
}

// parseIntField parses value as an int for the Statedef header field named
// name, returning a descriptive error on failure.
func parseIntField(name, value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", name, value, err)
	}
	return n, nil
}

// parseBoolField parses value as a bool for the Statedef header field named
// name, returning a descriptive error on failure.
func parseBoolField(name, value string) (bool, error) {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s %q: %w", name, value, err)
	}
	return b, nil
}

// stripComment removes a ".cns" comment from line — everything from the
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
