package cmd

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/openkakutou/character/cns"
)

// Format-preserving edits (see
// .vibe/decisions/033-format-preserving-edits-per-format-alignment.md).
//
// Rules, all deterministic. Two passes run in sequence on disjoint lines:
// first the linked state (see below), then the Remap/Defaults/Command
// sections; neither touches the other's lines. Unrelated sections, comments,
// blank lines, order and line endings stay byte-exact; an edited file equal
// to the parsed content reproduces the source exactly.
//   - Leading trivia (comments and blank lines directly above a header)
//     belongs to that header's section and is deleted with it. A section runs
//     from its trivia to the trivia of the next bracket line of any kind.
//   - [Remap]: a changed value is patched on its (last) source line, keeping
//     key spelling, spacing, the trailing ";" comment and the line ending; a
//     removed key's lines are deleted; new keys are appended, sorted, after
//     the last key line of the last [Remap] section. A missing section is
//     created (Serialize's text) just before the first [Defaults], [Command],
//     [Statedef]/[State] section, or at end of file when there is none.
//   - [Defaults]: command.time / command.buffer.time (any key casing) are
//     patched in place; a newly needed key is appended after the last key
//     line (time first); a missing section is created before the first
//     [Command]/[Statedef]/[State] section, or at end of file.
//   - [Command]: commands are aligned by longest common subsequence on full
//     Command equality; unmatched runs are paired positionally as
//     modifications: only the name/command/time/buffer.time lines that
//     changed are patched (the last line when a key repeats), a zero
//     time/buffer.time deletes its line (zero means "not set"), a newly
//     needed key is appended after the section's last key line (name,
//     command, time, buffer.time order). A surplus edited command is inserted,
//     as Serialize writes it, right after the section of the nearest
//     preceding surviving command, else before the first [Command], else
//     before the first [Statedef]/[State] section, else at end of file. A
//     surplus source command is deleted with its leading trivia. Same-name
//     commands (alternative inputs) are ordinary distinct elements.
//   - States: delegated to cns.EditSource over the same source, so the
//     [Statedef -1] block is patched in place. When the source has no
//     [Statedef -1] (implicit header, see ensureImplicitStatedef) a marked
//     synthetic header is given to cns and removed again from its output, so
//     it is never written.
//   - Generated lines use the file's dominant line ending.

const implicitMarker = ";kakutou-implicit-statedef"

var implicitHeaderPattern = regexp.MustCompile(`(?i)^\s*\[\s*statedef\s+-1\s*\]`)

type sectionKind int

const (
	kindOther sectionKind = iota
	kindRemap
	kindDefaults
	kindCommand
	kindState // [Statedef ...] / [State ...]
)

type section struct {
	kind                sectionKind
	trivia, header, end int
	keys                map[string][]int // lowercase key -> lines
	lastKV              int              // last key line, header when none
}

type textIndex struct {
	lines    []string // each keeps its terminator
	eol      string
	sections []section
}

func splitLines(src string) []string {
	var lines []string
	for src != "" {
		n := strings.IndexByte(src, '\n') + 1
		if n == 0 {
			n = len(src)
		}
		lines = append(lines, src[:n])
		src = src[n:]
	}
	return lines
}

func dominantEOL(src []byte) string {
	if crlf := bytes.Count(src, []byte("\r\n")); crlf > bytes.Count(src, []byte("\n"))-crlf {
		return "\r\n"
	}
	return "\n"
}

func indexText(src []byte) *textIndex {
	ix := &textIndex{lines: splitLines(string(src)), eol: dominantEOL(src)}
	var brackets []int
	for i, l := range ix.lines {
		if strings.HasPrefix(stripComment(l), "[") {
			brackets = append(brackets, i)
		}
	}
	trivia := func(i int) int {
		for i > 0 && stripComment(ix.lines[i-1]) == "" {
			i--
		}
		return i
	}
	for b, h := range brackets {
		end := len(ix.lines)
		if b+1 < len(brackets) {
			end = trivia(brackets[b+1])
		}
		line := stripComment(ix.lines[h])
		s := section{kind: kindOther, trivia: trivia(h), header: h, end: end, keys: map[string][]int{}, lastKV: h}
		if strings.HasSuffix(line, "]") {
			switch strings.ToLower(strings.TrimSpace(line[1 : len(line)-1])) {
			case "remap":
				s.kind = kindRemap
			case "defaults":
				s.kind = kindDefaults
			case "command":
				s.kind = kindCommand
			}
		}
		if s.kind == kindOther && (statedefHeaderAttemptPattern.MatchString(line) || stateHeaderAttemptPattern.MatchString(line)) {
			s.kind = kindState
		}
		for i := h + 1; i < end; i++ {
			if key, _, ok := parseKeyValueLine(stripComment(ix.lines[i])); ok {
				s.keys[normalizeCommandKey(key)] = append(s.keys[normalizeCommandKey(key)], i)
				s.lastKV = i
			}
		}
		ix.sections = append(ix.sections, s)
	}
	return ix
}

// firstSection returns the trivia line of the first section whose kind is in
// kinds, or the line count when there is none.
func (ix *textIndex) firstSection(kinds ...sectionKind) int {
	for _, s := range ix.sections {
		if slices.Contains(kinds, s.kind) {
			return s.trivia
		}
	}
	return len(ix.lines)
}

type editPlan struct {
	ix   *textIndex
	del  []bool
	repl map[int]string
	ins  map[int][]string
}

func (p *editPlan) insert(at int, text string) { p.ins[at] = append(p.ins[at], text) }
func (p *editPlan) remove(from, to int) {
	for i := from; i < to; i++ {
		p.del[i] = true
	}
}

// insertSection inserts a whole section (text ends with its last line's
// terminator) at line at, a section's trivia start or the end of file,
// separated from its neighbours by one blank line.
func (p *editPlan) insertSection(at int, text string) {
	if at > 0 {
		text = "\n" + text
	}
	if at < len(p.ix.lines) && stripComment(p.ix.lines[at]) != "" {
		text += "\n"
	}
	p.insert(at, p.gen(text))
}
func (p *editPlan) gen(text string) string { return strings.ReplaceAll(text, "\n", p.ix.eol) }

// patchLine replaces the value of the "key = value" line raw with val,
// keeping indentation, key spelling, spacing, quotes, the trailing comment
// and the line terminator. shortcut: same logic as cns's unexported
// patchLine, duplicated to avoid widening cns's API; merge if cns exports it.
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

func lastLine(lines []int) int { return lines[len(lines)-1] }

// setKey patches the line Parse retains for the key: the last line across
// all spellings in names (the last one wins at parse time), or returns false
// when none exists.
func (p *editPlan) setKey(s section, val string, names ...string) bool {
	line := -1
	for _, n := range names {
		if lines := s.keys[n]; len(lines) > 0 {
			line = max(line, lastLine(lines))
		}
	}
	if line < 0 {
		return false
	}
	p.repl[line] = patchLine(p.ix.lines[line], val)
	return true
}

func (p *editPlan) deleteKey(s section, names ...string) {
	for _, n := range names {
		for _, l := range s.keys[n] {
			p.del[l] = true
		}
	}
}

func render(fn func(w io.Writer) error) (string, error) {
	var buf bytes.Buffer
	if err := fn(&buf); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

type pair struct{ i, j int }

// align pairs n baseline items with m edited ones: LCS anchors, unmatched
// runs between anchors paired positionally, the surplus is {i,-1} (deleted)
// or {-1,j} (inserted). Result is in edited order.
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

func (p *editPlan) editRemap(base, edited map[string]string) {
	var secs []section
	for _, s := range p.ix.sections {
		if s.kind == kindRemap {
			secs = append(secs, s)
		}
	}
	var added []string
	for k, v := range edited {
		old, had := base[k]
		if !had {
			added = append(added, k)
			continue
		}
		if old == v {
			continue
		}
		for i := len(secs) - 1; i >= 0; i-- { // last section wins, like Parse
			if p.setKey(secs[i], v, k) {
				break
			}
		}
	}
	for k := range base {
		if _, still := edited[k]; !still {
			for _, s := range secs {
				p.deleteKey(s, k)
			}
		}
	}
	sort.Strings(added)
	if len(added) == 0 {
		return
	}
	if len(secs) > 0 {
		last := secs[len(secs)-1]
		for _, k := range added {
			p.insert(last.lastKV+1, p.gen(k+" = "+edited[k]+"\n"))
		}
		return
	}
	text, _ := render(func(w io.Writer) error { return writeRemapSection(w, edited) })
	p.insertSection(p.ix.firstSection(kindDefaults, kindCommand, kindState), text)
}

func (p *editPlan) editDefaults(base, edited CommandDefaults) error {
	if base == edited {
		return nil
	}
	var secs []section
	for _, s := range p.ix.sections {
		if s.kind == kindDefaults {
			secs = append(secs, s)
		}
	}
	if len(secs) == 0 {
		text, err := render(func(w io.Writer) error { return writeDefaultsSection(w, edited) })
		if err != nil {
			return err
		}
		p.insertSection(p.ix.firstSection(kindCommand, kindState), text)
		return nil
	}
	last := secs[len(secs)-1]
	fields := []struct {
		old, val int
		names    []string
	}{
		{base.Time, edited.Time, []string{"command.time"}},
		{base.BufferTime, edited.BufferTime, []string{"command.buffer.time"}},
	}
	for _, f := range fields {
		if f.old == f.val {
			continue
		}
		patched := false
		for i := len(secs) - 1; i >= 0 && !patched; i-- {
			patched = p.setKey(secs[i], fmt.Sprint(f.val), f.names...)
		}
		if !patched {
			p.insert(last.lastKV+1, p.gen(fmt.Sprintf("%s = %d\n", f.names[0], f.val)))
		}
	}
	return nil
}

func (p *editPlan) editCommand(s section, b, e Command) {
	if v := e.Name; b.Name != v {
		if !p.setKey(s, v, "name") {
			p.insert(s.lastKV+1, p.gen(fmt.Sprintf("name = %q\n", v)))
		}
	}
	if b.Input != e.Input {
		if !p.setKey(s, e.Input, "command") {
			p.insert(s.lastKV+1, p.gen("command = "+e.Input+"\n"))
		}
	}
	numeric := []struct {
		old, val int
		names    []string
	}{
		{b.Time, e.Time, []string{"time", "command.time"}},
		{b.BufferTime, e.BufferTime, []string{"buffer.time", "command.buffer.time"}},
	}
	for _, f := range numeric {
		switch {
		case f.old == f.val:
		case f.val == 0:
			p.deleteKey(s, f.names...)
		case !p.setKey(s, fmt.Sprint(f.val), f.names...):
			p.insert(s.lastKV+1, p.gen(fmt.Sprintf("%s = %d\n", f.names[0], f.val)))
		}
	}
}

func (p *editPlan) editCommands(base, edited []Command) error {
	var secs []section
	for _, s := range p.ix.sections {
		if s.kind == kindCommand {
			secs = append(secs, s)
		}
	}
	if len(secs) != len(base) {
		return fmt.Errorf("cmd: source has %d [Command] sections, parsed %d", len(secs), len(base))
	}
	lastSrc := -1
	for _, pr := range align(len(base), len(edited), func(i, j int) bool { return base[i] == edited[j] }) {
		switch {
		case pr.j < 0:
			p.remove(secs[pr.i].trivia, secs[pr.i].end)
		case pr.i < 0:
			text, err := render(func(w io.Writer) error { return writeCommandSection(w, edited[pr.j]) })
			if err != nil {
				return err
			}
			switch {
			case lastSrc >= 0:
				p.insert(secs[lastSrc].end, p.gen("\n"+text))
			case len(secs) > 0:
				p.insertSection(secs[0].trivia, text)
			default:
				p.insertSection(p.ix.firstSection(kindState), text)
			}
		default:
			lastSrc = pr.i
			p.editCommand(secs[pr.i], base[pr.i], edited[pr.j])
		}
	}
	return nil
}

func (ix *textIndex) apply(p *editPlan) []byte {
	var out bytes.Buffer
	put := func(s string) {
		if out.Len() > 0 && out.Bytes()[out.Len()-1] != '\n' {
			out.WriteString(ix.eol)
		}
		out.WriteString(s)
	}
	for i := 0; i <= len(ix.lines); i++ {
		for _, t := range p.ins[i] {
			put(t)
		}
		switch {
		case i == len(ix.lines) || p.del[i]:
		default:
			if r, ok := p.repl[i]; ok {
				put(r)
			} else {
				put(ix.lines[i])
			}
		}
	}
	return out.Bytes()
}

// editStates applies states to source through cns.EditSource, handling the
// implicit [Statedef -1] header.
func editStates(source []byte, states []cns.StateDef) ([]byte, error) {
	withHeader := ensureImplicitStatedef(source)
	if bytes.Equal(withHeader, source) {
		var buf bytes.Buffer
		if err := cns.EditSource(&buf, source, states); err != nil {
			return nil, fmt.Errorf("cmd: editing linked state: %w", err)
		}
		return buf.Bytes(), nil
	}
	// Implicit header: give cns a marked synthetic header at the same spot.
	srcLines, synLines := splitLines(string(source)), splitLines(string(withHeader))
	k := 0
	for k < len(srcLines) && srcLines[k] == synLines[k] {
		k++
	}
	term := srcLines[k][len(strings.TrimRight(srcLines[k], "\r\n")):]
	if term == "" {
		term = "\n"
	}
	marked := slices.Concat(srcLines[:k], []string{"[Statedef -1] " + implicitMarker + term}, srcLines[k:])
	var buf bytes.Buffer
	if err := cns.EditSource(&buf, []byte(strings.Join(marked, "")), states); err != nil {
		return nil, fmt.Errorf("cmd: editing linked state: %w", err)
	}
	var out []string
	outLines := splitLines(buf.String())
	for i, l := range outLines {
		if strings.Contains(l, implicitMarker) {
			// Keep the header only when cns changed it (renumbered, or header
			// key lines added right after it); otherwise it stays implicit.
			changed := !implicitHeaderPattern.MatchString(l) || (i+1 < len(outLines) && k < len(srcLines) && outLines[i+1] != srcLines[k])
			if changed {
				out = append(out, strings.Replace(l, " "+implicitMarker, "", 1))
			}
			continue
		}
		out = append(out, l)
	}
	return []byte(strings.Join(out, "")), nil
}

// SerializeEdited writes the Document's source to w with file applied: only
// what differs from the parsed content is rewritten, everything else
// (comments, blank lines, order, unrelated sections, line endings) stays
// byte-exact. A file equal to the parsed content reproduces the source
// exactly. The rules are documented at the top of edit.go.
func (d *Document) SerializeEdited(w io.Writer, file CommandFile) error {
	base, err := Parse(bytes.NewReader(d.source))
	if err != nil {
		return err
	}
	text, err := editStates(d.source, file.States)
	if err != nil {
		return err
	}
	ix := indexText(text)
	p := &editPlan{ix: ix, del: make([]bool, len(ix.lines)), repl: map[int]string{}, ins: map[int][]string{}}
	if !maps.Equal(base.Remap, file.Remap) {
		p.editRemap(base.Remap, file.Remap)
	}
	if err := p.editDefaults(base.Defaults, file.Defaults); err != nil {
		return err
	}
	if !slices.Equal(base.Commands, file.Commands) {
		if err := p.editCommands(base.Commands, file.Commands); err != nil {
			return err
		}
	}
	if _, err := w.Write(ix.apply(p)); err != nil {
		return fmt.Errorf("cmd: writing edited document: %w", err)
	}
	return nil
}
