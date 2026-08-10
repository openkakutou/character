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

// stateHeaderPattern matches any "[State ...]" line — a numbered
// "[State N]"/"[State N, label]" header, or a bare label with no number at
// all (e.g. "[State removeexplod]"), which real MUGEN/Ikemen engines accept
// and treat identically. The content is never captured: it is not stored on
// Controller (see cns.Controller) and has no effect other than marking the
// start of a controller block attached to whichever Statedef is currently
// open. It never fires on a "[Statedef ...]" line: "statedef" has no space
// (or "]") right after "state", so it fails the (\s|\]) requirement below.
// See .vibe/decisions/022-cns-parse-state-header-accepts-any-label.md.
var stateHeaderPattern = regexp.MustCompile(`(?i)^\[\s*state(\s.*)?\]$`)

// stateAttemptPattern recognizes any bracket line that starts with the
// "state" keyword (but not "statedef" — see stateHeaderPattern's own note),
// whether or not it goes on to match stateHeaderPattern. Used only to detect
// a "[State ...]" header missing its closing "]" (see
// recoverMissingClosingBracket) — unlike statedefAttemptPattern, it plays no
// role in error reporting once a header has a closing bracket, since
// stateHeaderPattern already accepts any bracketed content after "state".
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
// number) returns a descriptive, line-numbered error rather than being
// silently skipped, as does a "[State N]" block with no enclosing Statedef
// and a reader that fails outright. A missing closing bracket on an
// otherwise-recognizable Statedef/State header attempt is recovered rather
// than treated as an error (see recoverMissingClosingBracket); only a
// bracket line missing "]" that isn't recognizable as either header attempt
// returns the "malformed section header" error. See
// .vibe/decisions/012-cns-parse-header-detection-strategy.md.
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
				recovered, ok := recoverMissingClosingBracket(line)
				if !ok {
					return nil, fmt.Errorf("cns: line %d: malformed section header %q", lineNumber, line)
				}
				line = recovered
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

			if stateHeaderPattern.MatchString(line) {
				if current == nil {
					return nil, fmt.Errorf("cns: line %d: [State ...] block found outside of any Statedef: %q", lineNumber, line)
				}
				current.Controllers = append(current.Controllers, Controller{})
				currentCtrl = &current.Controllers[len(current.Controllers)-1]
				continue
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

// recoverMissingClosingBracket handles a bracket line with no closing "]" —
// a real-world .cns authoring typo (see backlog item 042: a full corpus scan
// found "[State N, label"/"[Statedef N, label" missing their closing
// bracket is the single most common character.Load failure, ~15% of a
// 717-file corpus) that real MUGEN/Ikemen engines tolerate. When line looks
// like an attempt at a "[Statedef ...]" or "[State ...]" header (matches
// statedefAttemptPattern/stateAttemptPattern), the missing "]" is appended
// so the normal header-matching logic below runs exactly as it would on a
// correctly bracketed line, producing equivalent StateDef/Controller data —
// including still returning a "malformed Statedef header" error for a
// recognizable-but-invalid attempt (e.g. a non-numeric Statedef number).
// Any other bracket line missing "]" is genuinely unrelated content, not a
// typo in either header shape this parser is responsible for, and is left
// unrecovered — the caller then returns the "malformed section header"
// error, matching .vibe/decisions/012's distinction.
func recoverMissingClosingBracket(line string) (string, bool) {
	if statedefAttemptPattern.MatchString(line) || stateAttemptPattern.MatchString(line) {
		return line + "]", true
	}
	return line, false
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
// being built. An unrecognized key is ignored. A recognized boolean field
// with a value that doesn't parse as a bool returns a descriptive error. A
// recognized numeric field with a value that doesn't parse as a literal
// integer never errors: it is treated as an unevaluated MUGEN/Ikemen trigger
// expression and stored in HeaderExprs instead — see
// .vibe/decisions/023-statedef-numeric-header-fields-unevaluated-expression-escape-hatch.md.
func applyStatedefField(s *StateDef, key, value string) error {
	switch strings.ToLower(key) {
	case "type":
		s.Type = StateType(value)
	case "movetype":
		s.MoveType = MoveType(value)
	case "physics":
		s.Physics = PhysicsType(value)
	case "anim":
		applyIntOrExprField(s, "anim", value, &s.Anim)
	case "ctrl":
		b, err := parseBoolField("ctrl", value)
		if err != nil {
			return err
		}
		s.Ctrl = b
	case "poweradd":
		applyIntOrExprField(s, "poweradd", value, &s.PowerAdd)
	case "juggle":
		applyIntOrExprField(s, "juggle", value, &s.Juggle)
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
		applyIntOrExprField(s, "sprpriority", value, &s.SprPriority)
	}
	// Any other key is unrecognized within a Statedef header and ignored.
	return nil
}

// applyIntOrExprField sets *dst from value when value parses as a literal
// integer. Otherwise it leaves *dst untouched (at its zero value) and
// records value verbatim in s.HeaderExprs under name instead — value is
// treated as an unevaluated MUGEN/Ikemen trigger expression, since Parse has
// no expression grammar to tell a real expression apart from a malformed
// value. See
// .vibe/decisions/023-statedef-numeric-header-fields-unevaluated-expression-escape-hatch.md.
func applyIntOrExprField(s *StateDef, name, value string, dst *int) {
	n, err := strconv.Atoi(value)
	if err != nil {
		if s.HeaderExprs == nil {
			s.HeaderExprs = make(map[string]string)
		}
		s.HeaderExprs[name] = value
		return
	}
	*dst = n
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
