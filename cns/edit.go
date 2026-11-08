package cns

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Format-preserving edits (see
// .vibe/decisions/033-format-preserving-edits-per-format-alignment.md).
//
// Rules, all deterministic:
//   - Comments and blank lines directly above a header ("leading trivia")
//     belong to that header's element and are deleted with it. A block
//     (Statedef or State) runs from its trivia to the trivia of the next
//     bracket line of any kind, so unrelated sections ([Data], [Size], ...)
//     are never part of, nor touched by, an edit.
//   - StateDefs are aligned by longest common subsequence on Number; the
//     unmatched runs between two anchors are paired positionally as
//     modifications (a renumbered state keeps its comments, only the number
//     in the header line is rewritten, header spelling kept). Surplus edited
//     StateDefs are inserted right after the block of the nearest preceding
//     edited StateDef that exists in the source, or just before the first
//     Statedef when there is none (at end of file when the source has no
//     Statedef), using Serialize's text. Surplus source StateDefs are
//     deleted with their controllers.
//   - Inside a modified StateDef each header field whose value changed is
//     patched on its (last) source line; a changed field with no source line
//     is added after the last key line of the header.
//   - Controllers are aligned the same way, by full equality. A modified
//     controller is patched in place: the `type` line; triggers matched by
//     index to their source lines (value replaced, key such as triggerall
//     kept; extra source lines deleted; extra triggers added after the last
//     trigger line as "trigger<N+1>", N being the number of the last kept
//     trigger key, or "trigger1" when that key has no number); parameters
//     (changed value patched, removed key's line deleted, new keys added
//     sorted after the last parameter line). Patching keeps key spelling,
//     the trailing ";" comment and the line ending. A controller whose
//     source repeats a parameter key is regenerated alone with Serialize's
//     controller text, keeping its leading trivia.
//   - Generated lines use the file's dominant line ending.

var (
	headerNumberPattern = regexp.MustCompile(`(?i)^(\s*\[\s*statedef\s+)-?\d+`)
	triggerNumPattern   = regexp.MustCompile(`(?i)^trigger(\d+)$`)
	headerKeys          = map[string]bool{"type": true, "movetype": true, "physics": true, "anim": true, "ctrl": true,
		"poweradd": true, "juggle": true, "facep2": true, "hitdefpersist": true, "movehitpersist": true,
		"hitcountpersist": true, "sprpriority": true}
)

type kvRef struct {
	line int
	key  string
}

type ctrlSpan struct {
	trivia, header, end int
	typeLine            int // -1 when absent
	triggers            []kvRef
	params              map[string][]int // lowercase key -> lines
	lastParam           int              // -1 when none
	lastKV              int              // last key line, header when none
}

type stateSpan struct {
	trivia, header, hdrEnd, end int
	fields                      map[string][]int // recognized header key -> lines
	lastKV                      int
	ctrls                       []ctrlSpan
}

// sourceIndex maps the model elements of a .cns source to its line ranges.
type sourceIndex struct {
	lines  []string // each keeps its line terminator
	eol    string
	states []stateSpan
}

const (
	kindNone = iota
	kindStatedef
	kindState
	kindOther
)

func classify(raw string) int {
	s := stripComment(raw)
	if !strings.HasPrefix(s, "[") {
		return kindNone
	}
	if !strings.HasSuffix(s, "]") {
		rec, ok := recoverMissingClosingBracket(s)
		if !ok {
			return kindOther
		}
		s = rec
	}
	switch {
	case statedefHeaderPattern.MatchString(s):
		return kindStatedef
	case statedefAttemptPattern.MatchString(s):
		return kindOther
	case stateHeaderPattern.MatchString(s):
		return kindState
	}
	return kindOther
}

func indexSource(src []byte) *sourceIndex {
	idx := &sourceIndex{eol: "\n"}
	if crlf := bytes.Count(src, []byte("\r\n")); crlf > bytes.Count(src, []byte("\n"))-crlf {
		idx.eol = "\r\n"
	}
	text := string(src)
	for text != "" {
		n := strings.IndexByte(text, '\n') + 1
		if n == 0 {
			n = len(text)
		}
		idx.lines = append(idx.lines, text[:n])
		text = text[n:]
	}

	var brackets []int
	kinds := make([]int, len(idx.lines))
	for i, l := range idx.lines {
		if kinds[i] = classify(l); kinds[i] != kindNone {
			brackets = append(brackets, i)
		}
	}
	trivia := func(i int) int {
		for i > 0 && stripComment(idx.lines[i-1]) == "" {
			i--
		}
		return i
	}

	inState := false
	for b, h := range brackets {
		end := len(idx.lines)
		if b+1 < len(brackets) {
			end = trivia(brackets[b+1])
		}
		switch kinds[h] {
		case kindStatedef:
			idx.states = append(idx.states, stateSpan{trivia: trivia(h), header: h, hdrEnd: end, end: end,
				fields: map[string][]int{}, lastKV: h})
			inState = true
			sp := &idx.states[len(idx.states)-1]
			for i := h + 1; i < end; i++ {
				if key, _, ok := parseKeyValueLine(stripComment(idx.lines[i])); ok {
					sp.lastKV = i
					if lk := strings.ToLower(key); headerKeys[lk] {
						sp.fields[lk] = append(sp.fields[lk], i)
					}
				}
			}
		case kindState:
			if !inState {
				continue
			}
			sp := &idx.states[len(idx.states)-1]
			cs := ctrlSpan{trivia: trivia(h), header: h, end: end, typeLine: -1, lastParam: -1, lastKV: h,
				params: map[string][]int{}}
			for i := h + 1; i < end; i++ {
				key, _, ok := parseKeyValueLine(stripComment(idx.lines[i]))
				if !ok {
					continue
				}
				cs.lastKV = i
				switch lk := strings.ToLower(key); {
				case strings.HasPrefix(lk, "trigger"):
					cs.triggers = append(cs.triggers, kvRef{i, key})
				case lk == "type":
					cs.typeLine = i
				default:
					cs.params[lk] = append(cs.params[lk], i)
					cs.lastParam = i
				}
			}
			sp.ctrls = append(sp.ctrls, cs)
			sp.end = end
		default:
			inState = false
		}
	}
	return idx
}

type plan struct {
	idx  *sourceIndex
	del  []bool
	repl map[int]string
	ins  map[int][]string
}

func (p *plan) insert(at int, text string) { p.ins[at] = append(p.ins[at], text) }

func (p *plan) remove(from, to int) {
	for i := from; i < to; i++ {
		p.del[i] = true
	}
}

func (p *plan) gen(text string) string {
	return strings.ReplaceAll(text, "\n", p.idx.eol)
}

// patchLine replaces the value of the "key = value" line raw with val,
// keeping everything else: indentation, key spelling, spacing, quotes, the
// trailing comment and the line terminator.
func patchLine(raw, val string) string {
	content := strings.TrimRight(raw, "\r\n")
	term := raw[len(content):]
	eq := strings.IndexByte(content, '=')
	end := len(content)
	if semi := strings.IndexByte(content, ';'); semi >= 0 {
		end = semi
	}
	vs := eq + 1
	for vs < end && (content[vs] == ' ' || content[vs] == '\t') {
		vs++
	}
	ve := end
	for ve > vs && (content[ve-1] == ' ' || content[ve-1] == '\t') {
		ve--
	}
	if old := content[vs:ve]; len(old) >= 2 && old[0] == '"' && old[len(old)-1] == '"' {
		val = `"` + val + `"`
	}
	return content[:vs] + val + content[ve:] + term
}

type pair struct{ i, j int }

// align pairs n baseline items with m edited items: an LCS under eq gives
// the anchors; unmatched runs between anchors are paired positionally; the
// surplus is {i,-1} (deleted) or {-1,j} (inserted). Result is in edited order.
func align(n, m int, eq func(i, j int) bool) []pair {
	l := make([][]int, n+1)
	for i := range l {
		l[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if eq(i, j) {
				l[i][j] = l[i+1][j+1] + 1
			} else {
				l[i][j] = max(l[i+1][j], l[i][j+1])
			}
		}
	}
	var anchors []pair
	for i, j := 0, 0; i < n && j < m; {
		switch {
		case eq(i, j) && l[i][j] == l[i+1][j+1]+1:
			anchors = append(anchors, pair{i, j})
			i, j = i+1, j+1
		case l[i+1][j] >= l[i][j+1]:
			i++
		default:
			j++
		}
	}
	anchors = append(anchors, pair{n, m})

	var out []pair
	i, j := 0, 0
	for _, a := range anchors {
		for ; i < a.i && j < a.j; i, j = i+1, j+1 {
			out = append(out, pair{i, j})
		}
		for ; i < a.i; i++ {
			out = append(out, pair{i, -1})
		}
		for ; j < a.j; j++ {
			out = append(out, pair{-1, j})
		}
		if a.i < n {
			out = append(out, a)
			i, j = i+1, j+1
		}
	}
	return out
}

func ctrlEqual(a, b Controller) bool {
	return a.Type == b.Type && slices.Equal(a.Triggers, b.Triggers) && maps.Equal(a.Parameters, b.Parameters)
}

func stateEqual(a, b StateDef) bool {
	return a.Number == b.Number && slices.Equal(headerFields(a), headerFields(b)) &&
		slices.EqualFunc(a.Controllers, b.Controllers, ctrlEqual)
}

// blockText renders text the way Serialize does, minus its trailing blank line.
func blockText(render func(w io.Writer) error) (string, error) {
	var buf bytes.Buffer
	if err := render(&buf); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

func (p *plan) editState(sp stateSpan, b, e StateDef) error {
	idx := p.idx
	if b.Number != e.Number {
		raw := idx.lines[sp.header]
		if m := headerNumberPattern.FindStringSubmatchIndex(raw); m != nil {
			p.repl[sp.header] = raw[:m[3]] + strconv.Itoa(e.Number) + raw[m[1]:]
		}
	}
	bf, ef := headerFields(b), headerFields(e)
	var added []string
	for k := range ef {
		if bf[k] == ef[k] {
			continue
		}
		if lines := sp.fields[ef[k].key]; len(lines) > 0 {
			last := lines[len(lines)-1]
			p.repl[last] = patchLine(idx.lines[last], ef[k].value)
		} else {
			added = append(added, p.gen(ef[k].key+" = "+ef[k].value+"\n"))
		}
	}
	for _, t := range added {
		p.insert(sp.lastKV+1, t)
	}

	lastSrc := -1
	for _, pr := range align(len(b.Controllers), len(e.Controllers), func(i, j int) bool {
		return ctrlEqual(b.Controllers[i], e.Controllers[j])
	}) {
		switch {
		case pr.j < 0:
			p.remove(sp.ctrls[pr.i].trivia, sp.ctrls[pr.i].end)
		case pr.i < 0:
			text, err := blockText(func(w io.Writer) error { return writeController(w, e.Number, e.Controllers[pr.j]) })
			if err != nil {
				return err
			}
			at := sp.hdrEnd
			if lastSrc >= 0 {
				at = sp.ctrls[lastSrc].end
			}
			p.insert(at, p.gen("\n"+text))
		default:
			lastSrc = pr.i
			if err := p.editController(sp.ctrls[pr.i], e.Number, b.Controllers[pr.i], e.Controllers[pr.j]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *plan) editController(cs ctrlSpan, number int, b, e Controller) error {
	if ctrlEqual(b, e) {
		return nil
	}
	idx := p.idx
	for _, lines := range cs.params {
		if len(lines) > 1 {
			text, err := blockText(func(w io.Writer) error { return writeController(w, number, e) })
			if err != nil {
				return err
			}
			p.remove(cs.header, cs.end)
			p.insert(cs.header, p.gen(text))
			return nil
		}
	}

	if b.Type != e.Type {
		if cs.typeLine >= 0 {
			p.repl[cs.typeLine] = patchLine(idx.lines[cs.typeLine], e.Type)
		} else {
			p.insert(cs.header+1, p.gen("type = "+e.Type+"\n"))
		}
	}

	// triggers: by index onto their source lines
	for k := 0; k < len(cs.triggers) && k < len(e.Triggers); k++ {
		if b.Triggers[k] != e.Triggers[k] {
			p.repl[cs.triggers[k].line] = patchLine(idx.lines[cs.triggers[k].line], e.Triggers[k])
		}
	}
	for k := len(e.Triggers); k < len(cs.triggers); k++ {
		p.remove(cs.triggers[k].line, cs.triggers[k].line+1)
	}
	if len(e.Triggers) > len(cs.triggers) {
		anchor, next := cs.header, 1
		if cs.typeLine >= 0 {
			anchor = cs.typeLine
		}
		if n := len(cs.triggers); n > 0 {
			anchor = cs.triggers[n-1].line
			if m := triggerNumPattern.FindStringSubmatch(cs.triggers[n-1].key); m != nil {
				num, _ := strconv.Atoi(m[1])
				next = num + 1
			}
		}
		for k := len(cs.triggers); k < len(e.Triggers); k++ {
			p.insert(anchor+1, p.gen(fmt.Sprintf("trigger%d = %s\n", next, e.Triggers[k])))
			next++
		}
	}

	// parameters
	var added []string
	for key, val := range e.Parameters {
		old, had := b.Parameters[key]
		switch {
		case !had:
			added = append(added, key)
		case old != val:
			line := cs.params[key][0]
			p.repl[line] = patchLine(idx.lines[line], val)
		}
	}
	for key := range b.Parameters {
		if _, still := e.Parameters[key]; !still {
			p.remove(cs.params[key][0], cs.params[key][0]+1)
		}
	}
	sort.Strings(added)
	anchor := cs.lastParam
	if anchor < 0 {
		anchor = cs.lastKV
	}
	for _, key := range added {
		p.insert(anchor+1, p.gen(key+" = "+e.Parameters[key]+"\n"))
	}
	return nil
}

func (idx *sourceIndex) edit(w io.Writer, base, edited []StateDef) error {
	p := &plan{idx: idx, del: make([]bool, len(idx.lines)), repl: map[int]string{}, ins: map[int][]string{}}
	lastSrc := -1
	for _, pr := range align(len(base), len(edited), func(i, j int) bool { return base[i].Number == edited[j].Number }) {
		switch {
		case pr.j < 0:
			p.remove(idx.states[pr.i].trivia, idx.states[pr.i].end)
		case pr.i < 0:
			text, err := blockText(func(w io.Writer) error { return Serialize(w, []StateDef{edited[pr.j]}) })
			if err != nil {
				return err
			}
			switch {
			case lastSrc >= 0:
				p.insert(idx.states[lastSrc].end, p.gen("\n"+text))
			case len(idx.states) > 0:
				p.insert(idx.states[0].trivia, p.gen(text+"\n"))
			case len(idx.lines) > 0:
				p.insert(len(idx.lines), p.gen("\n"+text))
			default:
				p.insert(0, p.gen(text))
			}
		default:
			lastSrc = pr.i
			if stateEqual(base[pr.i], edited[pr.j]) {
				continue
			}
			if err := p.editState(idx.states[pr.i], base[pr.i], edited[pr.j]); err != nil {
				return err
			}
		}
	}

	var out bytes.Buffer
	put := func(s string) {
		if out.Len() > 0 && out.Bytes()[out.Len()-1] != '\n' {
			out.WriteString(idx.eol)
		}
		out.WriteString(s)
	}
	for i := 0; i <= len(idx.lines); i++ {
		for _, t := range p.ins[i] {
			put(t)
		}
		switch {
		case i == len(idx.lines) || p.del[i]:
		case p.repl[i] != "":
			put(p.repl[i])
		default:
			put(idx.lines[i])
		}
	}
	if _, err := w.Write(out.Bytes()); err != nil {
		return fmt.Errorf("cns: writing edited document: %w", err)
	}
	return nil
}

// SerializeEdited writes the Document's source to w with states applied:
// only what differs from the parsed baseline is rewritten, everything else
// (comments, blank lines, order, unrelated sections, line endings) stays
// byte-exact. states equal to the parsed content reproduces the source
// exactly. The rules are documented at the top of edit.go.
func (d *Document) SerializeEdited(w io.Writer, states []StateDef) error {
	if d.index == nil {
		d.index = indexSource(d.source)
	}
	return editSource(w, d.source, d.index, states)
}

// EditSource is SerializeEdited for raw .cns text that is not held in a
// Document (e.g. the Statedef -1 block embedded in a .cmd file): it writes
// source to w with states applied, leaving every unrelated line untouched.
func EditSource(w io.Writer, source []byte, states []StateDef) error {
	return editSource(w, source, indexSource(source), states)
}

func editSource(w io.Writer, source []byte, idx *sourceIndex, states []StateDef) error {
	base, err := Parse(bytes.NewReader(source))
	if err != nil {
		return err
	}
	if slices.EqualFunc(base, states, stateEqual) {
		if _, err := w.Write(source); err != nil {
			return fmt.Errorf("cns: writing edited document: %w", err)
		}
		return nil
	}
	return idx.edit(w, base, states)
}
